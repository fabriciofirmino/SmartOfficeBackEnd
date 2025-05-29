package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math" // Mantido por enquanto, AddScreen usa math.Ceil
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	// Tentativa de correção dos caminhos de importação com base no nome do módulo "apiBackEnd"
	"apiBackEnd/config" // Alterado de "apiBackEnd/db" para "apiBackEnd/config"
	"apiBackEnd/models"
	"apiBackEnd/utils"
)

// validateAndSanitizeField é uma função auxiliar para validar e sanitizar campos.
// Retorna uma string de erro se a validação falhar, ou uma string vazia se for bem-sucedido.
// Adicionei o parâmetro gin.Context para consistência, embora não é usado aqui.
func validateAndSanitizeField(fieldName, value string, minLen, maxLen int, c *gin.Context) string {
	if len(value) < minLen || len(value) > maxLen {
		return fmt.Sprintf("O campo %s deve ter entre %d e %d caracteres", fieldName, minLen, maxLen)
	}
	// A sanitização foi comentada pois a regexp não estava sendo usada e causava erro de import.
	// Se a sanitização for necessária, a lógica e a importação de "regexp" devem ser revisadas.
	return "" // Sem erros
}

// EditUser godoc
// @Summary Edita um usuário existente
// @Description Edita um usuário com base no ID fornecido. Permite a atualização de vários campos, incluindo nome de usuário, senha, notas do revendedor, número do WhatsApp, nome para aviso, envio de notificação, bouquet, aplicativos, preferências de notificação (Notificacao_conta, Notificacao_vods, Notificacao_jogos) e valor do plano.
// @Tags Tools Table
// @Security BearerAuth
// @Accept  json
// @Produce  json
// @Param id path int true "ID do Usuário"
// @Param user body models.EditUserRequest true "Dados do Usuário para Editar. Campos como 'Notificacao_conta', 'Notificacao_vods', 'Notificacao_jogos' esperam true/false e são armazenados como 1/0. 'Valor_plano' espera um valor decimal."
// @Success 200 {object} map[string]interface{} "Usuário editado com sucesso. Inclui todos os campos atualizados, como 'Valor_plano'."
// @Failure 400 {object} map[string]string "Erro: Requisição inválida ou dados ausentes"
// @Failure 404 {object} map[string]string "Erro: Usuário não encontrado"
// @Failure 500 {object} map[string]string "Erro: Erro interno do servidor"
// @Router /api/tools-table/edit/{id} [put]
// @Example request.body.todos_campos
//
//	{
//	  "username": "usuario_atualizado",
//	  "password": "senha_nova_123",
//	  "reseller_notes": "Notas importantes sobre este usuário.",
//	  "numero_whats": "5511987654321",
//	  "nome_para_aviso": "Nome de Aviso Atualizado",
//	  "enviar_notificacao": true,
//	  "bouquet": "[1, 5, 10]",
//	  "aplicativos": [
//	    {"app_id": 1, "name": "Meu App Favorito", "url": "https://meuapp.com/user", "active": true},
//	    {"app_id": 3, "name": "Outro App Util", "url": "https://outroapp.net", "active": true}
//	  ],
//	  "Notificacao_conta": true,
//	  "Notificacao_vods": false,
//	  "Notificacao_jogos": true,
//	  "franquia_member_id": 25,
//	  "Valor_plano": 99.99
//	}
//
// @Example request.body.apenas_notificacoes_e_valor
//
//	{
//	  "Notificacao_conta": false,
//	  "Notificacao_vods": true,
//	  "Notificacao_jogos": false,
//	  "Valor_plano": 29.50
//	}
func EditUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuário inválido"})
		return
	}

	var req models.EditUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Erro ao fazer bind do JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos", "details": err.Error()})
		return
	}

	// Validação dos campos
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		log.Printf("Erro de validação: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro de validação", "details": err.Error()})
		return
	}

	// Buscar dados antigos para o log de auditoria
	oldData, err := getUserDataForAudit(userID)
	if err != nil {
		log.Printf("Erro ao buscar dados antigos para auditoria do usuário %d: %v", userID, err)
		// Continuar mesmo se não encontrar dados antigos, mas logar o erro
	}

	tx, err := config.DB.Begin() // Alterado de db.DB para config.DB
	if err != nil {
		log.Printf("Erro ao iniciar transação: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno do servidor"})
		return
	}
	defer tx.Rollback() // Rollback em caso de erro

	var querySetters []string
	var queryArgs []interface{}
	// argCounter não é mais necessário para formatar os placeholders na query string

	if req.Username != "" {
		if errMsg := validateAndSanitizeField("username", req.Username, 4, 15, c); errMsg != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
			return
		}
		querySetters = append(querySetters, "username = ?")
		queryArgs = append(queryArgs, req.Username)
	}
	if req.Password != "" {
		if errMsg := validateAndSanitizeField("password", req.Password, 4, 15, c); errMsg != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
			return
		}
		querySetters = append(querySetters, "password = ?")
		queryArgs = append(queryArgs, req.Password) // Idealmente, a senha seria hasheada aqui
	}
	if req.ResellerNotes != "" {
		querySetters = append(querySetters, "reseller_notes = ?")
		queryArgs = append(queryArgs, req.ResellerNotes)
	}
	if req.NumeroWhats != nil {
		querySetters = append(querySetters, "numero_whats = ?")
		queryArgs = append(queryArgs, *req.NumeroWhats)
	}
	if req.NomeParaAviso != nil {
		querySetters = append(querySetters, "nome_para_aviso = ?")
		queryArgs = append(queryArgs, *req.NomeParaAviso)
	}
	if req.EnviarNotificacao != nil {
		querySetters = append(querySetters, "enviar_notificacao = ?")
		queryArgs = append(queryArgs, *req.EnviarNotificacao)
	}
	if req.Bouquet != "" {
		querySetters = append(querySetters, "bouquet = ?")
		queryArgs = append(queryArgs, req.Bouquet)
	}
	if len(req.Aplicativos) > 0 {
		aplicativosJSON, err := json.Marshal(req.Aplicativos)
		if err != nil {
			log.Printf("Erro ao fazer marshal dos aplicativos: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar aplicativos"})
			return
		}
		querySetters = append(querySetters, "aplicativo = ?")
		queryArgs = append(queryArgs, string(aplicativosJSON))
	}
	if req.Notificacao_conta != nil {
		querySetters = append(querySetters, "Notificacao_conta = ?")
		queryArgs = append(queryArgs, *req.Notificacao_conta)
	}
	if req.Notificacao_vods != nil {
		querySetters = append(querySetters, "Notificacao_vods = ?")
		queryArgs = append(queryArgs, *req.Notificacao_vods)
	}
	if req.Notificacao_jogos != nil {
		querySetters = append(querySetters, "Notificacao_jogos = ?")
		queryArgs = append(queryArgs, *req.Notificacao_jogos)
	}
	if req.FranquiaMemberID != nil {
		querySetters = append(querySetters, "franquia_member_id = ?")
		queryArgs = append(queryArgs, *req.FranquiaMemberID)
	}
	if req.Valor_plano != nil {
		querySetters = append(querySetters, "Valor_plano = ?")
		queryArgs = append(queryArgs, *req.Valor_plano)
	}

	if len(querySetters) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nenhum campo fornecido para atualização"})
		return
	}

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = ?", strings.Join(querySetters, ", "))
	finalQueryArgs := make([]interface{}, len(queryArgs))
	copy(finalQueryArgs, queryArgs)
	finalQueryArgs = append(finalQueryArgs, userID)

	// A lógica de loggableQuery não é mais necessária, pois 'query' já usa '?'
	log.Printf("Executando query: %s com args: %v", query, finalQueryArgs)

	_, err = tx.Exec(query, finalQueryArgs...)
	if err != nil {
		log.Printf("Erro ao executar a query de atualização para o usuário %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao editar usuário"})
		return
	}

	// Buscar os dados atualizados após o UPDATE, pois RETURNING não é suportado
	selectQuery := `SELECT username, password, reseller_notes, numero_whats, nome_para_aviso, enviar_notificacao, bouquet, aplicativo, Notificacao_conta, Notificacao_vods, Notificacao_jogos, franquia_member_id, Valor_plano FROM users WHERE id = ?`
	var updatedUsername, updatedPassword, updatedResellerNotes, updatedBouquet sql.NullString
	var updatedNumeroWhats, updatedNomeParaAviso sql.NullString
	var updatedEnviarNotificacao, updatedNotificacaoConta, updatedNotificacaoVods, updatedNotificacaoJogos sql.NullBool
	var updatedAplicativosJSON sql.NullString
	var updatedFranquiaMemberID sql.NullInt64
	var updatedValorPlano sql.NullFloat64 // Novo campo

	err = tx.QueryRow(selectQuery, userID).Scan(
		&updatedUsername,
		&updatedPassword,
		&updatedResellerNotes,
		&updatedNumeroWhats,
		&updatedNomeParaAviso,
		&updatedEnviarNotificacao,
		&updatedBouquet,
		&updatedAplicativosJSON,
		&updatedNotificacaoConta,
		&updatedNotificacaoVods,
		&updatedNotificacaoJogos,
		&updatedFranquiaMemberID,
		&updatedValorPlano, // Novo campo
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Usuário com ID %d não encontrado após atualização (isso não deveria acontecer).", userID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado após atualização"})
			return
		}
		log.Printf("Erro ao buscar dados atualizados do usuário %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar dados atualizados do usuário"})
		return
	}

	// Log de auditoria
	newData := map[string]interface{}{
		"username":           updatedUsername.String,
		"password":           "********", // Não logar a senha diretamente
		"reseller_notes":     updatedResellerNotes.String,
		"numero_whats":       updatedNumeroWhats.String,
		"nome_para_aviso":    updatedNomeParaAviso.String,
		"enviar_notificacao": updatedEnviarNotificacao.Bool,
		"bouquet":            updatedBouquet.String,
		"Notificacao_conta":  updatedNotificacaoConta.Bool,
		"Notificacao_vods":   updatedNotificacaoVods.Bool,
		"Notificacao_jogos":  updatedNotificacaoJogos.Bool,
		"franquia_member_id": updatedFranquiaMemberID.Int64,
		"Valor_plano":        updatedValorPlano.Float64, // Novo campo
	}
	if updatedAplicativosJSON.Valid {
		var apps []models.AplicativoInfo
		if err := json.Unmarshal([]byte(updatedAplicativosJSON.String), &apps); err == nil {
			newData["aplicativos"] = apps
		} else {
			newData["aplicativos"] = updatedAplicativosJSON.String // Fallback para string se não puder unmarshal
			log.Printf("Erro ao fazer unmarshal dos aplicativos atualizados para log: %v", err)
		}
	} else {
		newData["aplicativos"] = nil
	}

	// Comentando a chamada para saveAuditLog que usa SQL Tx
	/*
		if auditErr := saveAuditLog(tx, userID, "edit_user", oldData, newData); auditErr != nil {
			log.Printf("Erro ao salvar log de auditoria para edição do usuário %d: %v", userID, auditErr)
		}
	*/

	// Nova chamada para saveAuditLog com MongoDB
	if err := saveAuditLogToMongo(c, userID, "edit-user-specific-action", oldData, newData); err != nil {
		log.Printf("Erro ao salvar log de auditoria no MongoDB para edição do usuário %d: %v", userID, err)
		// Decidir se este erro deve impedir o commit da transação principal ou apenas ser logado
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Erro ao commitar transação: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno do servidor ao salvar alterações"})
		return
	}

	response := gin.H{
		"message":            "Usuário editado com sucesso",
		"id":                 userID,
		"username":           updatedUsername.String,
		"reseller_notes":     updatedResellerNotes.String,
		"numero_whats":       updatedNumeroWhats.String,
		"nome_para_aviso":    updatedNomeParaAviso.String,
		"enviar_notificacao": updatedEnviarNotificacao.Bool,
		"bouquet":            updatedBouquet.String,
		"Notificacao_conta":  updatedNotificacaoConta.Bool,
		"Notificacao_vods":   updatedNotificacaoVods.Bool,
		"Notificacao_jogos":  updatedNotificacaoJogos.Bool,
		"franquia_member_id": updatedFranquiaMemberID.Int64,
		"Valor_plano":        updatedValorPlano.Float64, // Adicionado aqui
	}
	if updatedAplicativosJSON.Valid {
		var apps []models.AplicativoInfo
		if err := json.Unmarshal([]byte(updatedAplicativosJSON.String), &apps); err == nil {
			response["aplicativos"] = apps
		} else {
			response["aplicativos"] = updatedAplicativosJSON.String
		}
	}

	log.Printf("Usuário %d editado com sucesso. Novos dados: %+v", userID, response)
	c.JSON(http.StatusOK, response)
}

// saveAuditLogToMongo registra uma ação no log de auditoria usando MongoDB.
func saveAuditLogToMongo(c *gin.Context, targetUserID int, action string, oldData, newData map[string]interface{}) error {
	adminID := 0 // Valor padrão se não encontrar no contexto
	if memberIDVal, exists := c.Get("member_id"); exists {
		if id, ok := memberIDVal.(float64); ok { // JWT armazena números como float64
			adminID = int(id)
		} else if idInt, ok := memberIDVal.(int); ok { // Caso já seja int
			adminID = idInt
		}
	}

	var details map[string]interface{}
	if action == "remove_screen" {
		details = map[string]interface{}{}
		if newData != nil {
			if val, ok := newData["total_telas_antes"]; ok {
				details["total_telas_antes"] = val
			}
			if val, ok := newData["total_telas_atual"]; ok {
				details["total_telas_atual"] = val
			}
		}
	} else if action == "add_screen" {
		details = map[string]interface{}{}
		if newData != nil {
			if val, ok := newData["valor_cobrado"]; ok {
				details["valor_cobrado"] = val
			}
			if val, ok := newData["creditos_antes"]; ok {
				details["creditos_antes"] = val
			}
			if val, ok := newData["creditos_atuais"]; ok {
				details["creditos_atuais"] = val
			}
			if val, ok := newData["total_telas_antes"]; ok {
				details["total_telas_antes"] = val
			}
			if val, ok := newData["total_telas_atual"]; ok {
				details["total_telas_atual"] = val
			}
		}
	} else if action == "edit-user-specific-action" {
		details = map[string]interface{}{
			"old_value": oldData,
			"new_value": newData,
		}
	} else {
		details = map[string]interface{}{}
	}

	logEntry := models.AuditLogEntry{
		Action:    action, // A ação original ainda é logada no documento
		UserID:    targetUserID,
		AdminID:   adminID, // ID do admin que realizou a ação
		Timestamp: time.Now(),
		Details:   details,
	}

	var collectionName string
	if action == "edit-user-specific-action" {
		collectionName = "Edit"       // Nome exato da coleção para esta ação específica
		logEntry.Action = "edit_user" // Mantém a ação como "edit_user" no log salvo
	} else {
		// Lógica padrão para outras ações
		collectionName = strings.ReplaceAll(action, "-", "_") + "_logs"
	}

	log.Printf("Tentando salvar log de auditoria. Banco de Dados: Logs, Coleção: %s, Ação no Documento: %s, Usuário: %d", collectionName, logEntry.Action, targetUserID)

	// Usa o banco de dados "Logs" e a coleção determinada
	collection := config.MongoDB.Database("Logs").Collection(collectionName)
	_, err := collection.InsertOne(context.TODO(), logEntry)
	if err != nil {
		log.Printf("ERRO ao inserir log de auditoria no MongoDB (db: Logs, collection: %s): %v", collectionName, err)
		return fmt.Errorf("erro ao inserir log de auditoria no MongoDB (db: Logs, collection: %s): %w", collectionName, err)
	}
	log.Printf("Log de auditoria salvo com sucesso no MongoDB (db: Logs, collection: %s) para ação no documento '%s', usuário %d", collectionName, logEntry.Action, targetUserID)
	return nil
}

// 📌 Adicionar Tela ao usuário
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
		c.JSON(http.StatusNotFound, gin.H{"erro": "Usuário não encontrado"})
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
	txCtx, err := config.DB.Begin()
	if err != nil {
		log.Printf("❌ ERRO ao iniciar transação: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao adicionar tela"})
		return
	}
	defer txCtx.Rollback() // Defer Rollback after successful Begin

	// 📌 Atualiza a quantidade de telas no usuário
	_, err = txCtx.Exec("UPDATE users SET max_connections = max_connections + 1 WHERE id = ?", req.UserID)
	if err != nil {
		log.Printf("❌ ERRO ao atualizar telas do usuário %d: %v", req.UserID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao adicionar tela"})
		return
	}

	// 📌 Atualiza os créditos na `reg_users`
	_, err = txCtx.Exec("UPDATE reg_users SET credits = credits - ? WHERE id = ?", valorCobrado, memberID)
	if err != nil {
		log.Printf("❌ ERRO ao atualizar créditos da revenda %d: %v", memberID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao descontar créditos"})
		return
	}

	// 📌 Confirma a transação
	err = txCtx.Commit()
	if err != nil {
		log.Printf("❌ ERRO ao confirmar transação: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao adicionar tela"})
		return
	}

	// Recalcula valores finais para o log
	totalTelasAtual := totalTelas + 1
	creditosDepois := creditosAtuais - valorCobrado

	// Dados para o log de auditoria
	newAuditData := map[string]interface{}{
		"valor_cobrado":     valorCobrado,
		"creditos_antes":    creditosAtuais,
		"creditos_atuais":   creditosDepois,
		"total_telas_antes": totalTelas,
		"total_telas_atual": totalTelasAtual,
	}

	if err := saveAuditLogToMongo(c, req.UserID, "add_screen", nil, newAuditData); err != nil {
		log.Printf("⚠️ AVISO: Falha ao salvar log de auditoria para add_screen do usuário %d: %v", req.UserID, err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Tela adicionada com sucesso!",
		"total_telas":     totalTelasAtual,
		"creditos_atuais": creditosDepois,
		"valor_cobrado":   valorCobrado,
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

	// 📌 Extrair `member_id` do token para validação e log
	tokenString := c.GetHeader("Authorization")
	claims, _, err := utils.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token inválido"})
		return
	}
	memberID := int(claims["member_id"].(float64))

	// 📌 Validar se o usuário pertence ao `member_id` autenticado
	var userMemberID int
	errDb := config.DB.QueryRow("SELECT member_id FROM users WHERE id = ?", req.UserID).Scan(&userMemberID)
	if errDb != nil {
		if errDb == sql.ErrNoRows {
			log.Printf("❌ ERRO ao buscar usuário %d para RemoveScreen: %v", req.UserID, errDb)
			c.JSON(http.StatusNotFound, gin.H{"erro": "Usuário não encontrado"})
			return
		}
		log.Printf("❌ ERRO ao buscar usuário %d para RemoveScreen: %v", req.UserID, errDb)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar informações do usuário"})
		return
	}
	if userMemberID != memberID {
		log.Printf("🚨 ALERTA! Tentativa de remoção de tela indevida! (Usuário: %d, Revenda Token: %d, Revenda Usuário: %d)", req.UserID, memberID, userMemberID)
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Usuário não pertence à sua revenda"})
		return
	}

	// 📌 Obtém total de telas
	var totalTelas int
	err = config.DB.QueryRow("SELECT max_connections FROM users WHERE id = ?", req.UserID).Scan(&totalTelas)
	if err != nil {
		log.Printf("Erro ao buscar total de telas para usuário %d: %v", req.UserID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar informações do usuário"})
		return
	}
	if totalTelas <= 1 {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "O usuário deve ter pelo menos 1 tela ativa"})
		return
	}

	// 📌 Atualiza banco de dados
	_, err = config.DB.Exec("UPDATE users SET max_connections = max_connections - 1 WHERE id = ?", req.UserID)
	if err != nil {
		log.Printf("Erro ao remover tela para usuário %d: %v", req.UserID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao remover tela"})
		return
	}

	// Log de auditoria para RemoveScreen
	newAuditData := map[string]interface{}{
		"total_telas_antes": totalTelas,
		"total_telas_atual": totalTelas - 1,
	}
	if auditErr := saveAuditLogToMongo(c, req.UserID, "remove_screen", nil, newAuditData); auditErr != nil {
		log.Printf("Erro ao salvar log de auditoria para remove_screen (usuário %d): %v", req.UserID, auditErr)
	}

	c.JSON(http.StatusOK, gin.H{
		"sucesso":     "Tela removida com sucesso",
		"total_telas": totalTelas - 1,
	})
}

func getUserDataForAudit(userID int) (map[string]interface{}, error) {
	// Altera o placeholder de $1 para ?
	query := `SELECT username, password, reseller_notes, numero_whats, nome_para_aviso, enviar_notificacao, bouquet, aplicativo, Notificacao_conta, Notificacao_vods, Notificacao_jogos, franquia_member_id, Valor_plano FROM users WHERE id = ?`
	row := config.DB.QueryRow(query, userID)

	var username, password, resellerNotes, bouquet, numeroWhats, nomeParaAviso sql.NullString
	var enviarNotificacao, notificacaoConta, notificacaoVods, notificacaoJogos sql.NullBool
	var aplicativosJSON sql.NullString
	var franquiaMemberID sql.NullInt64
	var valorPlano sql.NullFloat64 // Novo campo

	err := row.Scan(
		&username,
		&password,
		&resellerNotes,
		&numeroWhats,
		&nomeParaAviso,
		&enviarNotificacao,
		&bouquet,
		&aplicativosJSON,
		&notificacaoConta,
		&notificacaoVods,
		&notificacaoJogos,
		&franquiaMemberID,
		&valorPlano, // Novo campo
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("usuário %d não encontrado", userID)
		}
		return nil, fmt.Errorf("erro ao buscar dados do usuário %d: %w", userID, err)
	}

	data := map[string]interface{}{
		"username":           username.String,
		"password":           "********", // Não expor a senha
		"reseller_notes":     resellerNotes.String,
		"numero_whats":       numeroWhats.String,
		"nome_para_aviso":    nomeParaAviso.String,
		"enviar_notificacao": enviarNotificacao.Bool,
		"bouquet":            bouquet.String,
		"Notificacao_conta":  notificacaoConta.Bool,
		"Notificacao_vods":   notificacaoVods.Bool,
		"Notificacao_jogos":  notificacaoJogos.Bool,
		"franquia_member_id": franquiaMemberID.Int64,
		"Valor_plano":        valorPlano.Float64, // Novo campo
	}

	if aplicativosJSON.Valid && aplicativosJSON.String != "" {
		var apps []models.AplicativoInfo
		if errUnmarshal := json.Unmarshal([]byte(aplicativosJSON.String), &apps); errUnmarshal == nil {
			data["aplicativos"] = apps
		} else {
			data["aplicativos"] = aplicativosJSON.String // Fallback para string se não puder unmarshal
			log.Printf("Erro ao fazer unmarshal dos aplicativos para auditoria (usuário %d): %v", userID, errUnmarshal)
		}
	} else {
		data["aplicativos"] = nil
	}

	return data, nil
}
