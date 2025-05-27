package controllers

import (
	"apiBackEnd/config"
	"apiBackEnd/models"
	"apiBackEnd/utils"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"database/sql"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// validateAndSanitizeField valida, sanitiza e retorna um campo limpo
func validateAndSanitizeField(value string, fieldName string, minLen, maxLen int, c *gin.Context) string {
	if value == "" {
		return value
	}

	// 🔹 Remove espaços e caracteres especiais, mantendo apenas letras e números
	reg, _ := regexp.Compile("[^a-zA-Z0-9]")
	sanitizedValue := reg.ReplaceAllString(value, "")

	// 🔹 Validação de tamanho
	if len(sanitizedValue) < minLen || len(sanitizedValue) > maxLen {
		c.JSON(http.StatusBadRequest, gin.H{"erro": fmt.Sprintf("O campo %s deve ter entre %d e %d caracteres", fieldName, minLen, maxLen)})
		return ""
	}

	return sanitizedValue
}

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

	// Escolhe a coleção com base na ação
	var collectionName string
	switch action {
	case "add_screen", "remove_screen":
		collectionName = "Telas"
	case "edit_user":
		collectionName = "Edit"
	default:
		collectionName = "LogsGerais" // Se for outra ação, salva em uma coleção genérica
	}

	// Obtém a referência para a coleção correta
	collection := config.MongoDB.Database("Logs").Collection(collectionName)

	// Insere o log no MongoDB
	_, err := collection.InsertOne(context.TODO(), logEntry)
	if err != nil {
		fmt.Println("❌ Erro ao salvar log no MongoDB:", err)
	} else {
		fmt.Printf("✅ Log salvo na coleção '%s': %+v\n", collectionName, logEntry)
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

// EditUser permite editar os dados de um usuário
// @Summary Edita um usuário
// @Description Permite a edição de dados de um usuário na revenda autenticada
// @Tags ToolsTable
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "ID do usuário a ser editado"
// @Param request body models.EditUserRequest true "Dados do usuário a serem editados"
// @example
//
//	{
//	  "username": "usuario",
//	  "aplicativos": [
//	    {
//	      "device_id": "6541464646",
//	      "mac": "7C:0A:3F:D5:71:D8",
//	      "nome_do_aplicativo": "Duplecast",
//	      "vencimento_aplicativo": "2025-12-31"
//	    },
//	    {
//	      "device_id": "DVC002",
//	      "mac": "00:1A:2B:3C:4D:5F",
//	      "nome_do_aplicativo": "StreamPro",
//	      "vencimento_aplicativo": "2024-12-30"
//	    }
//	  ]
//	}
//
// @Success 200 {object} map[string]interface{} "Usuário editado com sucesso"
// @Failure 400 {object} map[string]string "Erro na requisição"
// @Failure 401 {object} map[string]string "Token inválido ou acesso negado"
// @Failure 500 {object} map[string]string "Erro interno ao processar a requisição"
// @Router /api/tools-table/edit/{id} [put]
func EditUser(c *gin.Context) {
	var req models.EditUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ Erro no bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos"})
		return
	}

	// 📌 Obtém o ID do usuário pela URL
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "ID inválido"})
		return
	}

	// 📌 Valida se o usuário pertence à revenda autenticada (AGORA BLOQUEIA SE FOR OUTRO MEMBER)
	memberID, err := validateUserAccess(c, userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": err.Error()})
		return
	}

	log.Printf("🔍 DEBUG - Editando usuário ID: %d (Revenda: %d)", userID, memberID)

	// 📌 Verifica novamente no banco se o usuário pertence à revenda correta antes do UPDATE
	var dbMemberID int
	err = config.DB.QueryRow("SELECT member_id FROM users WHERE id = ?", userID).Scan(&dbMemberID)
	if err != nil {
		log.Printf("❌ ERRO: Usuário ID %d não encontrado!", userID)
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Usuário não encontrado"})
		return
	}

	// 🚫 Se o usuário pertence a outra revenda, bloquear atualização
	if dbMemberID != memberID {
		log.Printf("🚨 ALERTA! Tentativa de edição de outro membro! (Usuário: %d, Revenda do Token: %d, Revenda do Usuário: %d)", userID, memberID, dbMemberID)
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Você não pode editar usuários de outra revenda!"})
		return
	}

	// 📌 Valida e limpa username
	if req.Username != "" {
		minUserLength := 4
		maxUserLength := 15

		req.Username = validateAndSanitizeField(req.Username, "username", minUserLength, maxUserLength, c)
		if req.Username == "" {
			return
		}

		// 📌 Verifica se o username já existe em toda a base
		var existingID int
		err = config.DB.QueryRow("SELECT id FROM users WHERE username = ? AND id != ?", req.Username, userID).Scan(&existingID)
		err = config.DB.QueryRow("SELECT id FROM users WHERE username = ? AND id != ?", req.Username, userID).Scan(&existingID)
		if err == nil {
			log.Println("❌ ERRO - Username já está em uso globalmente!")
			c.JSON(400, gin.H{"erro": "Username já está em uso!"})
			return
		}

	}

	// 📌 Valida e limpa password
	if req.Password != "" {
		req.Password = validateAndSanitizeField(req.Password, "senha", 6, 15, c)
		if req.Password == "" {
			return
		}
	}

	// 📌 Monta a query dinâmica de atualização
	updateFields := []string{}
	args := []interface{}{}

	if req.Username != "" {
		updateFields = append(updateFields, "username = ?")
		args = append(args, req.Username)
	}
	if req.Password != "" {
		updateFields = append(updateFields, "password = ?")
		args = append(args, req.Password)
	}
	if req.ResellerNotes != "" {
		updateFields = append(updateFields, "reseller_notes = ?")
		args = append(args, req.ResellerNotes)
	}
	if req.NumeroWhats != nil {
		updateFields = append(updateFields, "NUMERO_WHATS = ?")
		args = append(args, *req.NumeroWhats)
	}
	if req.NomeParaAviso != nil {
		updateFields = append(updateFields, "NOME_PARA_AVISO = ?")
		args = append(args, *req.NomeParaAviso)
	}
	if req.Bouquet != "" {
		updateFields = append(updateFields, "bouquet = ?")
		args = append(args, req.Bouquet)
	}

	// 📌 Processa os dados dos aplicativos e salva como JSON no banco de dados
	var aplicativosJSON string
	if req.Aplicativos != nil && len(req.Aplicativos) > 0 {
		appDataBytes, err := json.Marshal(req.Aplicativos)
		if err != nil {
			log.Println("❌ Erro ao converter array de aplicativos para JSON:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao processar os dados dos aplicativos"})
			return
		}
		aplicativosJSON = string(appDataBytes)
		updateFields = append(updateFields, "aplicativo = ?")
		args = append(args, aplicativosJSON)
	}

	if req.EnviarNotificacao != nil {
		val := 0
		if *req.EnviarNotificacao {
			val = 1
		}
		updateFields = append(updateFields, "ENVIAR_NOTIFICACAO = ?")
		args = append(args, val)
	}

	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Nenhum campo válido para atualização"})
		return
	}

	// 📌 Finaliza a query de atualização
	args = append(args, userID)
	query := fmt.Sprintf("UPDATE users SET %s WHERE id = ?", strings.Join(updateFields, ", "))

	// 📌 Executa o update
	res, err := config.DB.Exec(query, args...)
	if err != nil {
		log.Printf("❌ ERRO ao atualizar usuário ID %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao atualizar usuário"})
		return
	}

	// 📌 Verifica se o usuário foi atualizado
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("⚠️ Nenhuma alteração realizada para o usuário ID %d", userID)
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Nenhuma alteração realizada"})
		return
	}

	log.Printf("✅ Usuário ID %d atualizado com sucesso!", userID)

	// 📌 Obtém os dados antigos do usuário para log
	var oldUser models.EditUserRequest
	var aplicativoNull sql.NullString // Use NullString para lidar com valores NULL

	err = config.DB.QueryRow(`
	SELECT username, password, reseller_notes, NUMERO_WHATS, NOME_PARA_AVISO, 
	ENVIAR_NOTIFICACAO, bouquet, aplicativo 
	FROM users WHERE id = ?`, userID).
		Scan(&oldUser.Username, &oldUser.Password, &oldUser.ResellerNotes, &oldUser.NumeroWhats,
			&oldUser.NomeParaAviso, &oldUser.EnviarNotificacao, &oldUser.Bouquet, &aplicativoNull)

	if err != nil {
		log.Printf("❌ ERRO ao buscar dados antigos do usuário ID %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar dados antigos do usuário"})
		return
	}

	// Atribui o valor de aplicativo apenas se for válido
	if aplicativoNull.Valid {
		oldUser.Aplicativos = nil
		_ = json.Unmarshal([]byte(aplicativoNull.String), &oldUser.Aplicativos)
	} else {
		oldUser.Aplicativos = nil
	}

	// Variáveis temporárias para log
	var numeroWhatsLog, nomeParaAvisoLog string
	if req.NumeroWhats != nil {
		numeroWhatsLog = *req.NumeroWhats
	} else if oldUser.NumeroWhats != nil {
		numeroWhatsLog = *oldUser.NumeroWhats
	}
	if req.NomeParaAviso != nil {
		nomeParaAvisoLog = *req.NomeParaAviso
	} else if oldUser.NomeParaAviso != nil {
		nomeParaAvisoLog = *oldUser.NomeParaAviso
	}

	// Para o log, ajuste para mostrar o valor inteiro (0/1) ou nil
	var enviarNotificacaoLog interface{}
	if req.EnviarNotificacao != nil {
		if *req.EnviarNotificacao {
			enviarNotificacaoLog = 1
		} else {
			enviarNotificacaoLog = 0
		}
	} else if oldUser.EnviarNotificacao != nil {
		if *oldUser.EnviarNotificacao {
			enviarNotificacaoLog = 1
		} else {
			enviarNotificacaoLog = 0
		}
	} else {
		enviarNotificacaoLog = nil
	}

	// 📌 Salva Log no MongoDB com valores antigos e novos
	saveAuditLog("edit_user", userID, bson.M{
		"valores_anteriores": bson.M{
			"username":           oldUser.Username,
			"password":           oldUser.Password,
			"reseller_notes":     oldUser.ResellerNotes,
			"numero_whats":       oldUser.NumeroWhats,
			"nome_para_aviso":    oldUser.NomeParaAviso,
			"enviar_notificacao": oldUser.EnviarNotificacao,
			"bouquet":            oldUser.Bouquet,
			"aplicativos":        oldUser.Aplicativos,
		},
		"valores_novos": bson.M{
			"username":           req.Username,
			"password":           req.Password,
			"reseller_notes":     req.ResellerNotes,
			"numero_whats":       numeroWhatsLog,
			"nome_para_aviso":    nomeParaAvisoLog,
			"enviar_notificacao": enviarNotificacaoLog,
			"bouquet":            req.Bouquet,
			"aplicativos":        req.Aplicativos,
		},
		"timestamp": time.Now(),
	})

	// 📌 Retorna resposta
	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"message":     "Usuário atualizado com sucesso!",
		"aplicativos": req.Aplicativos,
	})
}
