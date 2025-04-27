package controllers

import (
	"apiBackEnd/config"
	"apiBackEnd/utils"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
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

	// Validation for status removed.

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

	// 🔹 5️⃣ Validar créditos disponíveis (antes da renovação)
	creditsFloat, exists := claims["credits"].(float64)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Créditos não encontrados no token"})
		return
	}
	creditosDisponiveis := int(creditsFloat)

	// 🔹 6️⃣ Definir custo e duração da renovação
	diasRenovacao := req.QuantidadeRenovacaoMes * 30
	log.Printf("Quantidade de meses para renovação: %d, Dias de renovação: %d", req.QuantidadeRenovacaoMes, diasRenovacao)

	// 🔹 **Calcular custo total com base em quantidade de meses e telas (max_connections)**
	custoTotal := req.QuantidadeRenovacaoMes * maxConnections
	log.Printf("Custo total calculado: %d (Quantidade de meses: %d * Máximo de conexões: %d)", custoTotal, req.QuantidadeRenovacaoMes, maxConnections)

	if creditosDisponiveis < custoTotal {
		log.Printf("Créditos insuficientes. Disponível: %d, Necessário: %d", creditosDisponiveis, custoTotal)
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
		// Se exp_date for maior ou igual à data atual, assume exp_date
		newExpDate = time.Unix(currentExpDate.Int64, 0).AddDate(0, 0, diasRenovacao)
	} else {
		// Se exp_date for menor ou igual à data atual, assume a data de hoje
		newExpDate = time.Now().AddDate(0, 0, diasRenovacao)
	}

	// Ajustar a nova data de expiração para sempre ser às 23h00
	newExpDate = time.Date(newExpDate.Year(), newExpDate.Month(), newExpDate.Day(), 23, 0, 0, 0, time.Local)
	newExpDateEpoch := newExpDate.Unix()

	// 🔹 8️⃣ Transação para atualização segura
	tx, err := config.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao iniciar transação"})
		return
	}

	// 🔹 **Debitar créditos antes de renovar**
	log.Printf("Tentando debitar créditos. MemberID: %d, CustoTotal: %d", memberID, custoTotal)
	result, err := tx.Exec("UPDATE streamcreed_db.reg_users SET credits = credits - ? WHERE id = ? AND credits >= ?", custoTotal, memberID, custoTotal)
	if err != nil {
		tx.Rollback()
		log.Printf("Erro ao debitar créditos para memberID %d: %v", memberID, err)
		c.JSON(http.StatusPaymentRequired, gin.H{"erro": "Não foi possível debitar os créditos para renovação"})
		return
	}

	// 🔹 **Verificar se o débito foi realizado**
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		tx.Rollback()
		log.Printf("Erro: débito de créditos não realizado para memberID %d. RowsAffected: %d", memberID, rowsAffected)
		c.JSON(http.StatusPaymentRequired, gin.H{"erro": "Créditos insuficientes ou débito não realizado"})
		return
	}

	log.Printf("Créditos debitados com sucesso para memberID %d. RowsAffected: %d", memberID, rowsAffected)

	// 🔹 **Verificar se os créditos foram debitados corretamente**
	var creditosRestantes int
	err = tx.QueryRow("SELECT credits FROM streamcreed_db.reg_users WHERE id = ?", memberID).Scan(&creditosRestantes)
	if err != nil || creditosRestantes < 0 {
		tx.Rollback()
		log.Printf("Erro ao verificar créditos restantes para memberID %d: %v", memberID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao verificar créditos após débito"})
		return
	}

	log.Printf("Créditos restantes para memberID %d: %d", memberID, creditosRestantes)

	// 🔹 **Renovar assinatura**
	_, err = tx.Exec("UPDATE streamcreed_db.users SET exp_date = ?, is_trial = '0' WHERE id = ?", newExpDateEpoch, userID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao atualizar exp_date"})
		return
	}

	// 🔹 **Finalizar transação**
	err = tx.Commit()
	if err != nil {
		log.Printf("Erro ao finalizar transação para memberID %d: %v", memberID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao finalizar transação"})
		return
	}

	// 🔹 **Salvar log no MongoDB**
	saveRenewLog(memberID, userID, currentExpDate.Int64, newExpDateEpoch, custoTotal)

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
		"creditos_restantes": creditosRestantes,
		"token_expira_em":    tempoRestanteFormatado, // 🔥 Agora formatado corretamente!
	})
}

// saveRenewLog salva os detalhes da renovação no MongoDB
func saveRenewLog(memberID, userID int, oldExpDate, newExpDate int64, creditsSpent int) {
	if config.MongoDB == nil {
		log.Println("⚠️ MongoDB não inicializado, ignorando log!")
		return
	}

	// Define o fuso horário UTC-3
	location, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		log.Printf("❌ Erro ao carregar fuso horário UTC-3: %v", err)
		return
	}

	// Ajusta o timestamp para UTC-3
	timestamp := time.Now().In(location)

	// Define a estrutura do log
	logEntry := bson.M{
		"member_id":     memberID,
		"user_id":       userID,
		"old_exp_date":  time.Unix(oldExpDate, 0).In(location).Format("2006-01-02 15:04:05"),
		"new_exp_date":  time.Unix(newExpDate, 0).In(location).Format("2006-01-02 15:04:05"),
		"credits_spent": creditsSpent,
		"timestamp":     timestamp.Format("2006-01-02 15:04:05"), // Salva como string formatada
	}

	// Obtém a referência para a coleção "renew"
	collection := config.MongoDB.Database("Logs").Collection("renew")

	// Insere o log no MongoDB
	_, err = collection.InsertOne(context.TODO(), logEntry)
	if err != nil {
		log.Printf("❌ Erro ao salvar log de renovação no MongoDB: %v", err)
	} else {
		log.Printf("✅ Log de renovação salvo no MongoDB: %+v", logEntry)
	}
}
