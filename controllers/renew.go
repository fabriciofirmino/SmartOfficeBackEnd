package controllers

import (
	"apiBackEnd/config"
	"apiBackEnd/utils"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Estrutura para receber a requisição de renovação
type RenewRequest struct {
	IDCliente              int `json:"id_cliente"`
	QuantidadeRenovacaoMes int `json:"quantidade_renovacao_em_meses"`
}

// RenewAccount renova a conta de um cliente.
//
// @Summary Renovar Conta
// @Description Atualiza a data de expiração da conta com base no tempo selecionado.
// @Tags Renovação
// @Security BearerAuth
// @Accept  json
// @Produce  json
// @Param renew body controllers.RenewRequest true "Dados para renovação"
// @Success 200 {object} map[string]interface{} "Conta renovada com sucesso"
// @Failure 400 {object} map[string]string "Erro na requisição"
// @Failure 401 {object} map[string]string "Token inválido ou conta bloqueada"
// @Failure 402 {object} map[string]string "Créditos insuficientes"
// @Router /api/renew [post]
func RenewAccount(c *gin.Context) {
	// 🔹 1️⃣ Autenticação e validação do token
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token não fornecido"})
		return
	}

	claims, timeRemaining, err := utils.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token inválido"})
		return
	}

	// 🔹 2️⃣ Extração dos dados do token (evita consultas desnecessárias ao banco)
	memberIDFloat, exists := claims["member_id"].(float64)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "MemberID não encontrado no token"})
		return
	}
	memberID := int(memberIDFloat)

	statusFloat, exists := claims["status"].(float64)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Status da conta não encontrado no token"})
		return
	}
	status := int(statusFloat)

	// 🔥 **Se o status for diferente de 1, bloqueia a renovação**
	if status != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Conta bloqueada para renovação"})
		return
	}

	log.Printf("🔍 MemberID extraído do token: %d | Status: %d\n", memberID, status)

	// 🔹 3️⃣ Ler o corpo da requisição
	var req RenewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos"})
		return
	}

	// 🔹 4️⃣ Validar se o cliente pertence ao `member_id`
	var userID, maxConnections int
	var currentExpDate sql.NullInt64

	query := `SELECT id, exp_date, max_connections FROM streamcreed_db.users WHERE id = ? AND member_id = ?`

	log.Printf("🔍 Executando query para buscar cliente: %s\n", query)

	err = config.DB.QueryRow(query, req.IDCliente, memberID).Scan(&userID, &currentExpDate, &maxConnections)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "Cliente não pertence a este MemberID"})
			return
		}
		log.Printf("❌ Erro ao buscar cliente: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar cliente"})
		return
	}

	// 🔹 5️⃣ Validar créditos disponíveis (vêm do token, evitando consulta extra)
	creditsFloat, exists := claims["credits"].(float64)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Créditos não encontrados no token"})
		return
	}
	creditosDisponiveis := int(creditsFloat)

	// 🔹 6️⃣ Definir custo e duração da renovação
	diasRenovacao := req.QuantidadeRenovacaoMes * 30
	custoPorPeriodo, _ := strconv.Atoi(os.Getenv(fmt.Sprintf("CREDITO_%d_MESES", req.QuantidadeRenovacaoMes)))
	custoTotal := custoPorPeriodo * maxConnections

	if creditosDisponiveis < custoTotal {
		c.JSON(http.StatusPaymentRequired, gin.H{
			"erro":                 "Créditos insuficientes para renovação",
			"creditos_disponiveis": creditosDisponiveis,
			"creditos_necessarios": custoTotal,
		})
		return
	}

	// 🔹 7️⃣ Calcular nova data de expiração (sempre às 23h00)
	now := time.Now().Unix()
	var newExpDate time.Time
	if currentExpDate.Valid && currentExpDate.Int64 >= now {
		newExpDate = time.Unix(currentExpDate.Int64, 0).AddDate(0, 0, diasRenovacao)
	} else {
		newExpDate = time.Now().AddDate(0, 0, diasRenovacao)
	}

	newExpDate = time.Date(newExpDate.Year(), newExpDate.Month(), newExpDate.Day(), 23, 0, 0, 0, time.Local)
	newExpDateEpoch := newExpDate.Unix()

	// 🔹 8️⃣ Transação para atualização segura
	tx, err := config.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao iniciar transação"})
		return
	}

	_, err = tx.Exec("UPDATE streamcreed_db.users SET exp_date = ?, is_trial = '0' WHERE id = ?", newExpDateEpoch, userID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao atualizar exp_date"})
		return
	}

	_, err = tx.Exec("UPDATE streamcreed_db.reg_users SET credits = credits - ? WHERE id = ?", custoTotal, memberID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao debitar créditos"})
		return
	}

	tx.Commit()
	// Converter `timeRemaining` (segundos) para dias, horas, minutos e segundos
	dias := timeRemaining / 86400
	horas := (timeRemaining % 86400) / 3600
	minutos := (timeRemaining % 3600) / 60
	segundos := timeRemaining % 60

	// Formatar a string de tempo restante
	tempoRestanteFormatado := fmt.Sprintf("%d dias, %d horas, %d minutos, %d segundos", dias, horas, minutos, segundos)

	// 🔹 ✅ Retorno da renovação
	c.JSON(http.StatusOK, gin.H{
		"status":             "Renovação concluída com sucesso",
		"id_cliente":         userID,
		"novo_exp_date":      newExpDateEpoch,
		"creditos_gastos":    custoTotal,
		"creditos_restantes": creditosDisponiveis - custoTotal,
		"token_expira_em":    tempoRestanteFormatado, // 🔥 Agora formatado corretamente!
	})

}
