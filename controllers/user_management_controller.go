package controllers

import (
	"apiBackEnd/config"
	"apiBackEnd/models"
	"apiBackEnd/utils"
	"database/sql" // Necessário para sql.NullString, etc.
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// UpdateUserStatusHandler godoc
// @Summary Ativar/Desativar Conta
// @Description Altera o campo "enabled" do usuário.
// @Tags Gerenciamento de Usuários
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param user_id path int true "ID do Usuário"
// @Param body body models.UserStatusPayload true "Payload contendo o novo status do usuário"
// @Success 200 {object} map[string]interface{} "Exemplo: {\"message\": \"Status do usuário atualizado com sucesso\"}"
// @Failure 400 {object} map[string]string "ID ou payload inválido"
// @Failure 401 {object} map[string]string "Token inválido"
// @Failure 403 {object} map[string]string "Usuário não tem permissão para alterar este usuário"
// @Failure 404 {object} map[string]string "Usuário não encontrado"
// @Failure 500 {object} map[string]string "Erro interno"
// @Router /api/users/{user_id}/status [patch]
func UpdateUserStatusHandler(c *gin.Context) {
	tokenInfo, ok := utils.ValidateAndExtractToken(c)
	if !ok {
		return
	}
	adminID := tokenInfo.MemberID

	userIDStr := c.Param("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuário inválido"})
		return
	}

	// Verificar permissão para modificar este usuário
	hasPermission, _, err := utils.VerificaPermissaoUsuario(userID, adminID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar permissões"})
		}
		return
	}

	if !hasPermission {
		c.JSON(http.StatusForbidden, gin.H{"error": "Você não tem permissão para alterar este usuário"})
		return
	}

	var payload models.UserStatusPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload inválido: " + err.Error()})
		return
	}

	previousState, err := utils.GetUserCurrentState(userID)
	if err != nil {
		log.Printf("Erro ao obter estado anterior do usuário %d para log: %v", userID, err)
	}

	query := "UPDATE streamcreed_db.users SET enabled = ? WHERE id = ?"
	result, err := config.DB.Exec(query, payload.Enabled, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar status do usuário"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado ou nenhuma alteração realizada"})
		return
	}

	action := "activate_user"
	if !payload.Enabled {
		action = "deactivate_user"
	}
	details := map[string]interface{}{
		"to": gin.H{"enabled": payload.Enabled},
	}
	if previousState != nil {
		details["from"] = previousState
	}
	utils.SaveAccountManagementAction(c.Request.Context(), action, userID, adminID, details)

	c.JSON(http.StatusOK, gin.H{"message": "Status do usuário atualizado com sucesso"})
}

// GetAllowedRegionsHandler godoc
// @Summary Obter Regiões Permitidas
// @Description Retorna as regiões permitidas configuradas na tabela settings como array de siglas.
// @Tags Gerenciamento de Regiões
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string][]string "Exemplo: {\"allowed_countries\": [\"AR\", \"AU\", \"BR\", \"CA\", \"EU\", \"FR\", \"DE\", \"JP\", \"PT\", \"RU\", \"SA\", \"ZA\", \"UA\", \"US\"]}"
// @Failure 500 {object} map[string]string "Erro interno"
// @Router /api/regions/allowed [get]
func GetAllowedRegionsHandler(c *gin.Context) {
	var allowCountries sql.NullString
	query := "SELECT allow_countries FROM streamcreed_db.settings LIMIT 1"
	err := config.DB.QueryRow(query).Scan(&allowCountries)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusOK, gin.H{"allowed_countries": []string{}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao consultar países permitidos"})
		return
	}

	var allowedList []string
	if allowCountries.Valid && allowCountries.String != "" {
		// Tenta decodificar o JSON armazenado no banco
		if err := json.Unmarshal([]byte(allowCountries.String), &allowedList); err != nil {
			// Caso a decodificação falhe, pode-se retornar uma lista vazia ou o valor original tratado como uma string
			allowedList = []string{}
		}
	}
	c.JSON(http.StatusOK, gin.H{"allowed_countries": allowedList})
}

// ForceUserRegionHandler godoc
// @Summary Forçar Região para Usuário
// @Description Altera o campo "forced_country" do usuário.
// @Tags Gerenciamento de Usuários
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param user_id path int true "ID do Usuário"
// @Param body body models.UserRegionPayload true "Payload com o país forçado (ex: US)"
// @Success 200 {object} map[string]interface{} "Exemplo: {\"message\": \"Região do usuário atualizada com sucesso\"}"
// @Failure 400 {object} map[string]string "ID ou payload inválido"
// @Failure 401 {object} map[string]string "Token inválido"
// @Failure 403 {object} map[string]string "Usuário não tem permissão para alterar este usuário"
// @Failure 404 {object} map[string]string "Usuário não encontrado"
// @Failure 500 {object} map[string]string "Erro interno"
// @Router /api/users/{user_id}/region [patch]
func ForceUserRegionHandler(c *gin.Context) {
	tokenInfo, ok := utils.ValidateAndExtractToken(c)
	if !ok {
		return
	}
	adminID := tokenInfo.MemberID
	// log.Printf("[DEBUG] ForceUserRegionHandler: Token validado, adminID: %d", adminID)

	userIDStr := c.Param("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuário inválido"})
		return
	}
	// log.Printf("[DEBUG] ForceUserRegionHandler: Processando userID: %d", userID)

	// Verificação de segurança: apenas o membro responsável pode alterar a região
	var responsibleMemberID int
	var enabled bool
	checkQuery := "SELECT member_id, enabled FROM streamcreed_db.users WHERE id = ?"
	err = config.DB.QueryRow(checkQuery, userID).Scan(&responsibleMemberID, &enabled)
	if err != nil {
		if err == sql.ErrNoRows {
			// log.Printf("[ERROR] ForceUserRegionHandler: Usuário %d não encontrado na tabela", userID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		} else {
			// log.Printf("[ERROR] ForceUserRegionHandler: Erro ao consultar usuário %d: %v", userID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar permissões"})
		}
		return
	}

	// log.Printf("[DEBUG] ForceUserRegionHandler: Usuário %d encontrado, responsibleMemberID: %d, adminID: %d, enabled: %v",
	//	userID, responsibleMemberID, adminID, enabled)

	// Se não for admin e não for o membro responsável, não permite a alteração
	isAdmin := adminID == 1 // Exemplo: assumindo que adminID 1 é super-admin
	if !isAdmin && adminID != responsibleMemberID {
		// log.Printf("[WARN] ForceUserRegionHandler: Usuário %d (memberID: %d) não tem permissão para alterar o usuário %d (memberID: %d)",
		//	adminID, adminID, userID, responsibleMemberID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Você não tem permissão para alterar este usuário"})
		return
	}

	var payload models.UserRegionPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		// log.Printf("[ERROR] ForceUserRegionHandler: Payload inválido: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload inválido: " + err.Error()})
		return
	}
	// log.Printf("[DEBUG] ForceUserRegionHandler: Payload recebido, forçando país: %s", payload.ForcedCountry)

	// Obter as regiões permitidas do settings
	var allowedJSON sql.NullString
	queryAllowed := "SELECT allow_countries FROM streamcreed_db.settings LIMIT 1"
	err = config.DB.QueryRow(queryAllowed).Scan(&allowedJSON)
	if err != nil && err != sql.ErrNoRows {
		// log.Printf("[ERROR] ForceUserRegionHandler: Erro ao consultar regiões permitidas: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao obter regiões permitidas"})
		return
	}

	var allowedList []string
	if allowedJSON.Valid && allowedJSON.String != "" {
		if err := json.Unmarshal([]byte(allowedJSON.String), &allowedList); err != nil {
			// log.Printf("[ERROR] ForceUserRegionHandler: Erro ao decodificar JSON de regiões permitidas: %v", err)
			allowedList = []string{}
		}
	}
	// log.Printf("[DEBUG] ForceUserRegionHandler: Regiões permitidas: %v", allowedList)

	// Verificar se payload.ForcedCountry está na lista permitida
	permitted := false
	for _, region := range allowedList {
		if region == payload.ForcedCountry {
			permitted = true
			break
		}
	}
	if !permitted {
		// log.Printf("[WARN] ForceUserRegionHandler: Região %s não está na lista de permitidas", payload.ForcedCountry)
		c.JSON(http.StatusForbidden, gin.H{"error": "Região não permitida"})
		return
	}

	previousState, err := utils.GetUserCurrentState(userID)
	if err != nil {
		// log.Printf("[WARN] ForceUserRegionHandler: Erro ao obter estado anterior do usuário %d: %v", userID, err)
	} else {
		// log.Printf("[DEBUG] ForceUserRegionHandler: Estado anterior do usuário %d: %v", userID, previousState)
	}

	query := "UPDATE streamcreed_db.users SET forced_country = ? WHERE id = ?"
	result, err := config.DB.Exec(query, payload.ForcedCountry, userID)
	if err != nil {
		// log.Printf("[ERROR] ForceUserRegionHandler: Erro ao atualizar região do usuário %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar região do usuário"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	// log.Printf("[DEBUG] ForceUserRegionHandler: Update realizado, linhas afetadas: %d", rowsAffected)
	if rowsAffected == 0 {
		// log.Printf("[ERROR] ForceUserRegionHandler: Nenhuma linha afetada para o usuário %d", userID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado ou nenhuma alteração realizada"})
		return
	}

	details := map[string]interface{}{
		"to": gin.H{"forced_country": payload.ForcedCountry},
	}
	if previousState != nil && previousState["forced_country"] != nil {
		details["from"] = gin.H{"forced_country": previousState["forced_country"]}
	} else if previousState != nil {
		details["from"] = gin.H{"forced_country": nil}
	}
	utils.SaveAccountManagementAction(c.Request.Context(), "force_region", userID, adminID, details)

	// log.Printf("[INFO] ForceUserRegionHandler: Região do usuário %d alterada para %s pelo adminID %d",
	//	userID, payload.ForcedCountry, adminID)

	c.JSON(http.StatusOK, gin.H{"message": "Região do usuário atualizada com sucesso"})
}

// KickUserSessionHandler godoc
// @Summary Expulsar Usuário (Kick)
// @Description Remove a sessão ativa do usuário da tabela "user_activity_now".
// @Tags Gerenciamento de Usuários
// @Security BearerAuth
// @Param user_id path int true "ID do Usuário"
// @Success 200 {object} map[string]interface{} "Exemplo: {\"message\": \"Sessão do usuário removida com sucesso\"}"
// @Failure 400 {object} map[string]string "ID inválido"
// @Failure 401 {object} map[string]string "Token inválido"
// @Failure 403 {object} map[string]string "Usuário não tem permissão para expulsar este usuário"
// @Failure 404 {object} map[string]string "Usuário não encontrado"
// @Failure 500 {object} map[string]string "Erro interno"
// @Router /api/users/{user_id}/session [delete]
func KickUserSessionHandler(c *gin.Context) {
	// Obter e validar o token do usuário
	tokenInfo, ok := utils.ValidateAndExtractToken(c)
	if !ok {
		return
	}
	adminID := tokenInfo.MemberID

	userIDStr := c.Param("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuário inválido"})
		return
	}

	// Verificar permissão para expulsar este usuário
	hasPermission, _, err := utils.VerificaPermissaoUsuario(userID, adminID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar permissões"})
		}
		return
	}

	if !hasPermission {
		c.JSON(http.StatusForbidden, gin.H{"error": "Você não tem permissão para expulsar este usuário"})
		return
	}

	// Verificar se o usuário tem sessão ativa antes de tentar remover
	var sessionCount int
	checkQuery := "SELECT COUNT(*) FROM streamcreed_db.user_activity_now WHERE user_id = ?"
	err = config.DB.QueryRow(checkQuery, userID).Scan(&sessionCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar sessões do usuário"})
		return
	}

	if sessionCount == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "Usuário não tem sessão ativa para ser removida"})
		return
	}

	// Remover sessão
	query := "DELETE FROM streamcreed_db.user_activity_now WHERE user_id = ?"
	result, err := config.DB.Exec(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao remover sessão do usuário"})
		return
	}

	rowsAffected, _ := result.RowsAffected()

	// Salvar log da ação
	details := map[string]interface{}{
		"sessions_removed": rowsAffected,
	}
	utils.SaveAccountManagementAction(c.Request.Context(), "kick_user", userID, adminID, details)

	c.JSON(http.StatusOK, gin.H{"message": "Sessão do usuário removida com sucesso", "sessions_removed": rowsAffected})
}

// SoftDeleteUserHandler godoc
// @Summary Exclusão Lógica de Conta
// @Description Marca o usuário como desativado e registra a exclusão lógica.
// @Tags Gerenciamento de Usuários
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param user_id path int true "ID do Usuário"
// @Success 200 {object} map[string]interface{} "Exemplo: {\"message\": \"Usuário excluído logicamente com sucesso\"}"
// @Failure 400 {object} map[string]string "ID inválido"
// @Failure 401 {object} map[string]string "Token inválido"
// @Failure 404 {object} map[string]string "Usuário não encontrado"
// @Failure 500 {object} map[string]string "Erro interno"
// @Router /api/users/{user_id} [delete]
func SoftDeleteUserHandler(c *gin.Context) {
	tokenInfo, ok := utils.ValidateAndExtractToken(c)
	if !ok {
		return
	}
	adminID := tokenInfo.MemberID

	userIDStr := c.Param("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuário inválido"})
		return
	}

	// Verificar permissão para excluir este usuário
	hasPermission, _, err := utils.VerificaPermissaoUsuario(userID, adminID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar permissões"})
		}
		return
	}

	if !hasPermission {
		c.JSON(http.StatusForbidden, gin.H{"error": "Você não tem permissão para excluir este usuário"})
		return
	}

	deletedAt := time.Now()
	query := `
		UPDATE streamcreed_db.users
		SET enabled = 0, date_deleted = ?, deleted = 1
		WHERE id = ?`
	result, err := config.DB.Exec(query, deletedAt, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir logicamente o usuário"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		return
	}

	details := map[string]interface{}{
		"deleted_at": deletedAt.Format(time.RFC3339),
	}
	utils.SaveAccountManagementAction(c.Request.Context(), "soft_delete_user", userID, adminID, details)

	c.JSON(http.StatusOK, gin.H{"message": "Usuário excluído logicamente com sucesso"})
}

// ListDeletedUsersHandler godoc
// @Summary Listar Contas Excluídas
// @Description Retorna a lista de usuários que foram excluídos logicamente nos últimos 30 dias.
// @Tags Gerenciamento de Usuários
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.DeletedUser "Lista de usuários excluídos"
// @Failure 500 {object} map[string]string "Erro interno"
// @Router /api/users/deleted [get]
func ListDeletedUsersHandler(c *gin.Context) {
	tokenInfo, ok := utils.ValidateAndExtractToken(c)
	if !ok {
		return
	}
	adminID := tokenInfo.MemberID

	// Verificar se é superadmin
	isAdmin := adminID == 1 // Super admin pode ver todos usuários excluídos

	// Query atualizada para incluir member_id e todos campos necessários
	baseQuery := `
		SELECT id, username, member_id, exp_date, max_connections, created_at, date_deleted, deleted 
		FROM streamcreed_db.users 
		WHERE deleted = 1 
		AND date_deleted IS NOT NULL`

	// Se não for admin, filtra apenas usuários da revenda
	query := baseQuery
	var args []interface{}
	if !isAdmin {
		query += " AND member_id = ?"
		args = append(args, adminID)
	}

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		log.Printf("ERRO SQL ao listar usuários excluídos: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar usuários excluídos"})
		return
	}
	defer rows.Close()

	// Em vez de usar diretamente o modelo DeletedUser, vamos criar uma estrutura temporária
	// que contém apenas os campos que queremos retornar
	type DeletedUserResponse struct {
		ID             int    `json:"id"`
		Username       string `json:"username"`
		Email          string `json:"email"`
		MemberID       int    `json:"member_id"`
		DeletedAt      *int64 `json:"deleted_at,omitempty"` // Timestamp UNIX
		ExpDate        *int64 `json:"exp_date,omitempty"`   // Timestamp UNIX
		MaxConnections int    `json:"max_connections"`
		CreatedAt      *int64 `json:"created_at,omitempty"` // Timestamp UNIX
	}

	var deletedUsers []DeletedUserResponse
	for rows.Next() {
		var u models.DeletedUser
		var expDate sql.NullInt64
		// Usar NullString para as datas que vêm como string do banco
		var dateDeletedStr sql.NullString
		var createdAtStr sql.NullString
		var deleted sql.NullInt64

		// Scan os valores como strings para os campos de data
		err := rows.Scan(
			&u.ID, &u.Username, &u.MemberID, &expDate, &u.MaxConnections,
			&createdAtStr, &dateDeletedStr, &deleted,
		)
		if err != nil {
			log.Printf("Erro ao escanear usuário: %v", err)
			continue
		}

		// Criar uma resposta simples com os campos necessários
		response := DeletedUserResponse{
			ID:             u.ID,
			Username:       u.Username,
			Email:          "",
			MemberID:       u.MemberID,
			MaxConnections: u.MaxConnections,
		}

		// Converter date_deleted string para timestamp UNIX
		if dateDeletedStr.Valid && dateDeletedStr.String != "" {
			// Tenta parsear a data no formato do MySQL
			t, err := time.Parse("2006-01-02 15:04:05", dateDeletedStr.String)
			if err == nil {
				unixTime := t.Unix()
				response.DeletedAt = &unixTime
				log.Printf("Data convertida: %s -> %d", dateDeletedStr.String, unixTime)
			} else {
				log.Printf("Erro ao converter data_deleted '%s': %v", dateDeletedStr.String, err)
			}
		}

		// Converter created_at string para timestamp UNIX
		if createdAtStr.Valid && createdAtStr.String != "" {
			// Tenta parsear a data no formato do MySQL
			t, err := time.Parse("2006-01-02 15:04:05", createdAtStr.String)
			if err == nil {
				unixTime := t.Unix()
				response.CreatedAt = &unixTime
				log.Printf("created_at convertido: %s -> %d", createdAtStr.String, unixTime)
			} else {
				// Se falhar na conversão, adicionar mesmo assim (como timestamp atual)
				currentTime := time.Now().Unix()
				response.CreatedAt = &currentTime
				log.Printf("Falha ao converter created_at '%s', usando timestamp atual: %d", createdAtStr.String, currentTime)
			}
		} else {
			// Mesmo se o campo for NULL, adicionar um valor (timestamp atual)
			currentTime := time.Now().Unix()
			response.CreatedAt = &currentTime
			log.Printf("Campo created_at é NULL, usando timestamp atual: %d", currentTime)
		}

		// ExpDate já está como int64, usar diretamente
		if expDate.Valid {
			response.ExpDate = &expDate.Int64
		}

		deletedUsers = append(deletedUsers, response)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Erro após iteração de resultados: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar resultados"})
		return
	}

	// Verificação após processar todos os usuários
	if len(deletedUsers) == 0 {
		// Retornar uma mensagem personalizada em vez de um array vazio
		c.JSON(http.StatusOK, gin.H{
			"message": "Nenhum usuário excluído para listar",
			"data":    []interface{}{}, // Array vazio para manter compatibilidade com as interfaces que esperam um array
		})
		return
	}

	// Antes de retornar, verificar e logar os dados
	log.Printf("Retornando %d usuários deletados", len(deletedUsers))
	for i, user := range deletedUsers {
		log.Printf("Usuário %d: ID=%d, created_at=%v", i, user.ID, user.CreatedAt)
	}

	c.JSON(http.StatusOK, deletedUsers)
}

// RestoreUserHandler godoc
// @Summary Restaurar Conta
// @Description Restaura a conta do usuário removendo os campos de exclusão lógica.
// @Tags Gerenciamento de Usuários
// @Security BearerAuth
// @Param user_id path int true "ID do Usuário"
// @Success 200 {object} map[string]interface{} "Exemplo: {\"message\": \"Usuário restaurado com sucesso\"}"
// @Failure 400 {object} map[string]string "ID inválido"
// @Failure 401 {object} map[string]string "Token inválido"
// @Failure 404 {object} map[string]string "Usuário não encontrado ou não está excluído"
// @Failure 500 {object} map[string]string "Erro interno"
// @Router /api/users/{user_id}/restore [patch]
func RestoreUserHandler(c *gin.Context) {
	tokenInfo, ok := utils.ValidateAndExtractToken(c)
	if !ok {
		return
	}
	adminID := tokenInfo.MemberID

	userIDStr := c.Param("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuário inválido"})
		return
	}

	// Verificar se o usuário existe e está excluído
	var responsibleMemberID int
	var isDeleted bool
	checkQuery := "SELECT member_id, deleted = 1 FROM streamcreed_db.users WHERE id = ?"
	err = config.DB.QueryRow(checkQuery, userID).Scan(&responsibleMemberID, &isDeleted)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar usuário"})
		}
		return
	}

	if !isDeleted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Usuário não está excluído logicamente"})
		return
	}

	// Verificar permissão para restaurar este usuário
	isAdmin := adminID == 1 // Super admin
	if !isAdmin && adminID != responsibleMemberID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Você não tem permissão para restaurar este usuário"})
		return
	}

	query := `
		UPDATE streamcreed_db.users
		SET enabled = 1, date_deleted = NULL, deleted = 0
		WHERE id = ? AND deleted = 1`
	result, err := config.DB.Exec(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao restaurar o usuário"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado ou não está excluído logicamente"})
		return
	}

	utils.SaveAccountManagementAction(c.Request.Context(), "restore_user", userID, adminID, nil)

	c.JSON(http.StatusOK, gin.H{"message": "Usuário restaurado com sucesso"})
}
