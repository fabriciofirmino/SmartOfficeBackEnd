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
// @Failure 401 {object} map[string]string "Token inválido"
// @Failure 402 {object} map[string]string "Créditos insuficientes"
// @Router /api/renew [post]
func RenewAccount(c *gin.Context) {
	// 🔹 1️⃣ Autenticação e validação do token
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token não fornecido"})
		return
	}

	claims, err := utils.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token inválido"})
		return
	}

	memberIDFloat, exists := claims["member_id"].(float64)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "MemberID não encontrado no token"})
		return
	}
	memberID := int(memberIDFloat)

	log.Printf("🔍 MemberID extraído do token: %d\n", memberID)

	// 🔹 2️⃣ Ler o corpo da requisição
	var req RenewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos"})
		return
	}

	// 🔹 3️⃣ Validar se o cliente pertence ao `member_id`
	var userID, userMemberID, maxConnections int
	var currentExpDate sql.NullInt64 // Agora armazenamos como timestamp Unix

	query := `
		SELECT id, member_id, exp_date, max_connections
		FROM streamcreed_db.users
		WHERE id = ? AND member_id = ?
	`
	log.Printf("🔍 Executando query para buscar cliente: %s\n", query)
	err = config.DB.QueryRow(query, req.IDCliente, memberID).Scan(&userID, &userMemberID, &currentExpDate, &maxConnections)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("❌ Nenhum cliente encontrado para o ID %d e MemberID %d\n", req.IDCliente, memberID)
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "Cliente não pertence a este MemberID"})
			return
		}
		log.Printf("❌ Erro ao buscar cliente: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar cliente"})
		return
	}

	// 🔹 4️⃣ Validar créditos do `member_id`
	var creditosDisponiveis int
	creditQuery := "SELECT credits FROM streamcreed_db.reg_users WHERE id = ?"
	log.Printf("🔍 Executando query para buscar créditos: %s | Param: %d\n", creditQuery, memberID)

	err = config.DB.QueryRow(creditQuery, memberID).Scan(&creditosDisponiveis)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("⚠️ Nenhum crédito encontrado para MemberID %d. Definindo como 0.\n", memberID)
			creditosDisponiveis = 0
		} else {
			log.Printf("❌ Erro ao buscar créditos do MemberID %d: %v", memberID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar créditos"})
			return
		}
	}

	log.Printf("✅ Créditos disponíveis para MemberID %d: %d\n", memberID, creditosDisponiveis)

	// 🔹 5️⃣ Buscar o custo por período nas variáveis de ambiente
	var custoPorPeriodo int
	var diasRenovacao int

	switch req.QuantidadeRenovacaoMes {
	case 1:
		custoPorPeriodo, _ = strconv.Atoi(os.Getenv("CREDITO_1_MES"))
		diasRenovacao = 31
	case 3:
		custoPorPeriodo, _ = strconv.Atoi(os.Getenv("CREDITO_3_MESES"))
		diasRenovacao = 93
	case 6:
		custoPorPeriodo, _ = strconv.Atoi(os.Getenv("CREDITO_6_MESES"))
		diasRenovacao = 186
	case 12:
		custoPorPeriodo, _ = strconv.Atoi(os.Getenv("CREDITO_12_MESES"))
		diasRenovacao = 365
	default:
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Quantidade de meses inválida"})
		return
	}

	if custoPorPeriodo == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Configuração de créditos não encontrada"})
		return
	}

	// 🔹 6️⃣ Calcular o custo total considerando `max_connections`
	custoTotal := custoPorPeriodo * maxConnections
	log.Printf("💰 Custo total para renovação (considerando conexões): %d créditos\n", custoTotal)

	if creditosDisponiveis < custoTotal {
		c.JSON(http.StatusPaymentRequired, gin.H{
			"erro":                 "Créditos insuficientes para renovação",
			"creditos_disponiveis": creditosDisponiveis,
			"creditos_necessarios": custoTotal,
			"mensagem": fmt.Sprintf(
				"Você tem %d créditos, mas para essa renovação são necessários %d créditos. Faça uma recarga e tente novamente.",
				creditosDisponiveis, custoTotal),
		})
		return
	}

	// 🔹 7️⃣ Criar transação para garantir consistência
	tx, err := config.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao iniciar transação"})
		return
	}

	// 🔹 8️⃣ Calcular a nova `exp_date` em `epoch`, ajustando para sempre ser às 23h00 no timezone correto
	now := time.Now().Unix()
	var newExpDate time.Time

	if currentExpDate.Valid && currentExpDate.Int64 >= now {
		// 🔥 Se a data de expiração ainda for válida, soma os dias a partir dela
		newExpDate = time.Unix(currentExpDate.Int64, 0).AddDate(0, 0, diasRenovacao)
	} else {
		// 🔥 Se já venceu, começa a contar a partir de agora
		newExpDate = time.Now().AddDate(0, 0, diasRenovacao)
	}

	// 🔥 Definir o horário fixo para 23h00 no timezone local do servidor
	location, _ := time.LoadLocation("America/Sao_Paulo") // 🔥 Ajuste para o fuso correto, se necessário
	newExpDate = time.Date(newExpDate.Year(), newExpDate.Month(), newExpDate.Day(), 23, 0, 0, 0, location)

	// Converter para Unix Timestamp (epoch)
	newExpDateEpoch := newExpDate.Unix()

	log.Printf("⏳ Nova expiração em UNIX Timestamp (23h00 horário correto): %d\n", newExpDateEpoch)

	// 🔹 9️⃣ Atualizar `exp_date` no banco dentro da transação
	_, err = tx.Exec("UPDATE streamcreed_db.users SET exp_date = ?, is_trial = '0' WHERE id = ?", newExpDateEpoch, userID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao atualizar data de expiração"})
		return
	}

	// 🔹 🔟 Debitar créditos dentro da transação
	_, err = tx.Exec("UPDATE streamcreed_db.reg_users SET credits = credits - ? WHERE id = ?", custoTotal, memberID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao debitar créditos"})
		return
	}

	// 🔹 ✅ Confirmar transação
	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao confirmar renovação"})
		return
	}

	// 🔹 🔥 Retorno de sucesso
	c.JSON(http.StatusOK, gin.H{
		"status":             "Renovação concluída com sucesso",
		"id_cliente":         userID,
		"novo_exp_date":      newExpDateEpoch, // ✅ Agora está em epoch
		"creditos_gastos":    custoTotal,
		"creditos_restantes": creditosDisponiveis - custoTotal,
	})
}
