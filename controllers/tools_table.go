package controllers

import (
	"apiBackEnd/config"
	"apiBackEnd/models"
	"apiBackEnd/utils"
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// 📌 Função para validar se o usuário pertence ao `member_id` do token
func validateUserAccess(c *gin.Context, userID int) (int, error) {
	// 📌 Extrai o token do cabeçalho
	tokenString := c.GetHeader("Authorization")
	claims, _, err := utils.ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}

	memberID := int(claims["member_id"].(float64))

	// 📌 Valida se o usuário pertence ao `member_id`
	var userMemberID int
	err = config.DB.QueryRow("SELECT member_id FROM users WHERE id = ?", userID).Scan(&userMemberID)
	if err != nil || userMemberID != memberID {
		return 0, err
	}

	return memberID, nil
}

// saveAuditLog salva logs de ações no MongoDB corretamente
func saveAuditLog(action string, userID int, details interface{}) {
	if config.MongoDB == nil {
		fmt.Println("⚠️ MongoDB não inicializado, ignorando log!")
		return
	}

	// Define a estrutura do log
	logEntry := bson.M{
		"user_id":   userID,
		"action":    action,
		"details":   details,
		"timestamp": time.Now(),
	}

	// Obtém a referência para a coleção `audit_logs`
	collection := config.MongoDB.Database("Logs").Collection("Telas")

	// Insere o log no MongoDB
	_, err := collection.InsertOne(context.TODO(), logEntry)
	if err != nil {
		fmt.Println("❌ Erro ao salvar log no MongoDB:", err)
	} else {
		fmt.Println("✅ Log salvo no MongoDB:", logEntry)
	}
}

// @Summary Adiciona uma nova tela ao usuário
// @Description Aumenta o número máximo de conexões do usuário e desconta créditos se aplicável
// @Tags ToolsTable
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.ScreenRequest true "JSON contendo o ID do usuário"
// @Success 200 {object} map[string]interface{} "Retorna o novo total de telas e o saldo de créditos atualizado"
// @Failure 400 {object} map[string]string "Erro nos parâmetros ou créditos insuficientes"
// @Failure 401 {object} map[string]string "Token inválido"
// @Failure 500 {object} map[string]string "Erro interno ao adicionar tela"
// @Router /api/tools-table/add-screen [post]
// 📌 Adicionar Tela ao usuário
func AddScreen(c *gin.Context) {
	var req models.ScreenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos"})
		return
	}

	// 📌 Extrair `member_id` do token
	tokenString := c.GetHeader("Authorization")
	claims, _, err := utils.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token inválido"})
		return
	}
	memberID := int(claims["member_id"].(float64))

	// 📌 Validar se o usuário pertence ao `member_id` autenticado
	var userMemberID int
	err = config.DB.QueryRow("SELECT member_id FROM users WHERE id = ?", req.UserID).Scan(&userMemberID)
	if err != nil {
		log.Printf("❌ ERRO ao buscar usuário %d: %v", req.UserID, err)
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Usuário não encontrado"})
		return
	}

	// 🔒 Garantir que o usuário pertence à revenda correta
	if userMemberID != memberID {
		log.Printf("🚨 ALERTA! Tentativa de alteração indevida! (Usuário: %d, Revenda Token: %d, Revenda Usuário: %d)", req.UserID, memberID, userMemberID)
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Usuário não pertence à sua revenda"})
		return
	}

	// 📌 Obtém o número atual de telas
	var totalTelas int
	err = config.DB.QueryRow("SELECT max_connections FROM users WHERE id = ?", req.UserID).Scan(&totalTelas)
	if err != nil {
		log.Printf("❌ ERRO ao buscar total de telas: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar total de telas"})
		return
	}

	if totalTelas >= 3 {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Limite máximo de telas atingido"})
		return
	}

	// 📌 Obtém data de vencimento
	var expDate int64
	err = config.DB.QueryRow("SELECT exp_date FROM users WHERE id = ?", req.UserID).Scan(&expDate)
	if err != nil {
		log.Printf("❌ ERRO ao buscar data de vencimento: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar data de vencimento"})
		return
	}

	// 📌 Calcula o custo da nova tela com base nos dias restantes
	diasRestantes := (expDate - time.Now().Unix()) / 86400
	var valorCobrado float64

	if diasRestantes <= 15 {
		valorCobrado = 0.5 // Contas com menos de 15 dias pagam meio crédito
	} else if diasRestantes > 30 {
		valorCobrado = math.Ceil(float64(diasRestantes) / 30) // Divide total de dias por 30
	} else {
		valorCobrado = 1.0 // Contas normais (até 30 dias)
	}

	// 📌 Obtém créditos do **MEMBER_ID** na tabela `reg_users`
	var creditosAtuais float64
	err = config.DB.QueryRow("SELECT credits FROM reg_users WHERE id = ?", memberID).Scan(&creditosAtuais)
	if err != nil {
		log.Printf("❌ ERRO ao buscar créditos da revenda %d: %v", memberID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar créditos do revendedor"})
		return
	}

	// 🔍 LOGS PARA DEBUG
	log.Printf("🟢 DEBUG - Créditos da revenda %d obtidos com sucesso!", memberID)
	log.Printf("🔹 Créditos antes da compra: %.2f", creditosAtuais)
	log.Printf("🔹 Dias restantes para expiração: %d", diasRestantes)
	log.Printf("🔹 Valor da tela a ser cobrado: %.2f", valorCobrado)

	// 📌 Verifica se há créditos suficientes
	if creditosAtuais < valorCobrado {
		log.Println("❌ ERRO: Créditos insuficientes!")
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Créditos insuficientes"})
		return
	}

	// 📌 Atualiza **os créditos da revenda** e aumenta telas do usuário
	tx, err := config.DB.Begin()
	if err != nil {
		log.Printf("❌ ERRO ao iniciar transação: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao adicionar tela"})
		return
	}

	// 📌 Atualiza a quantidade de telas no usuário
	_, err = tx.Exec("UPDATE users SET max_connections = max_connections + 1 WHERE id = ?", req.UserID)
	if err != nil {
		tx.Rollback()
		log.Printf("❌ ERRO ao atualizar telas do usuário %d: %v", req.UserID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao adicionar tela"})
		return
	}

	// 📌 Atualiza os créditos na `reg_users`
	_, err = tx.Exec("UPDATE reg_users SET credits = credits - ? WHERE id = ?", valorCobrado, memberID)
	if err != nil {
		tx.Rollback()
		log.Printf("❌ ERRO ao atualizar créditos da revenda %d: %v", memberID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao descontar créditos"})
		return
	}

	// 📌 Confirma a transação
	err = tx.Commit()
	if err != nil {
		log.Printf("❌ ERRO ao confirmar transação: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao adicionar tela"})
		return
	}

	// 📌 Salva log no MongoDB após adicionar tela
	saveAuditLog("add_screen", req.UserID, bson.M{
		"total_telas_antes": totalTelas,
		"total_telas_atual": totalTelas + 1,
		"valor_cobrado":     valorCobrado,
		"creditos_antes":    creditosAtuais,
		"creditos_atuais":   creditosAtuais - valorCobrado,
	})

	// 📌 Retorna resposta
	c.JSON(http.StatusOK, models.ScreenResponse{
		TotalTelas:     totalTelas + 1,
		ValorCobrado:   valorCobrado,
		CreditosAntes:  creditosAtuais,
		CreditosAtuais: creditosAtuais - valorCobrado,
	})
}

// @Summary Remove uma tela do usuário
// @Description Diminui o número máximo de conexões do usuário, garantindo que tenha pelo menos uma tela ativa
// @Tags ToolsTable
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.ScreenRequest true "JSON contendo o ID do usuário"
// @Success 200 {object} map[string]interface{} "Retorna o novo total de telas"
// @Failure 400 {object} map[string]string "Erro nos parâmetros ou limite mínimo atingido"
// @Failure 401 {object} map[string]string "Token inválido"
// @Failure 500 {object} map[string]string "Erro interno ao remover tela"
// @Router /api/tools-table/remove-screen [post]
// 📌 Remover Tela do usuário
func RemoveScreen(c *gin.Context) {
	var req models.ScreenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos"})
		return
	}

	// 📌 Valida o token e se o usuário pertence ao `member_id`
	_, err := validateUserAccess(c, req.UserID) // Remove `memberID`
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Acesso negado"})
		return
	}

	// 📌 Obtém total de telas
	var totalTelas int
	err = config.DB.QueryRow("SELECT max_connections FROM users WHERE id = ?", req.UserID).Scan(&totalTelas)
	if err != nil || totalTelas <= 1 {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "O usuário deve ter pelo menos 1 tela ativa"})
		return
	}

	// 📌 Atualiza banco de dados
	_, err = config.DB.Exec("UPDATE users SET max_connections = max_connections - 1 WHERE id = ?", req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao remover tela"})
		return
	}

	// 📌 Salva log no MongoDB após remover tela
	saveAuditLog("remove_screen", req.UserID, bson.M{
		"total_telas_antes": totalTelas,
		"total_telas_atual": totalTelas - 1,
	})

	// 📌 Retorna resposta
	c.JSON(http.StatusOK, gin.H{
		"sucesso":     "Tela removida com sucesso",
		"total_telas": totalTelas - 1,
	})
}
