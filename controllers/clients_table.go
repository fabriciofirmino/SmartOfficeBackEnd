package controllers

import (
	"apiBackEnd/config"
	"apiBackEnd/models"
	"apiBackEnd/utils"
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetClientsTable retorna clientes paginados e filtrados para o DataTable, incluindo status online e expiração
// @Summary Retorna clientes paginados e filtrados
// @Description Retorna uma lista de clientes paginada e filtrada para uso em DataTables, associados ao member_id do token. Inclui filtro de status online e expiração.
// @Tags ClientsTable
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "Número da página (padrão: 1)"
// @Param limit query int false "Limite de registros por página (padrão: 10)"
// @Param search query string false "Termo de pesquisa para filtrar por username ou reseller_notes"
// @Param online query bool false "Filtrar usuários online (true para listar apenas online, false para todos)"
// @Param expiration_filter query int false "Filtrar clientes por vencimento (7, 15, 30, custom até 90 ou \'0\' para vencidos)"
// @Param franquia_member_id query int false "Filtrar por ID do membro da franquia"
// @Param is_trial query string false "Filtrar por status de trial (0 para não trial, 1 para trial)"
// @Success 200 {object} map[string]interface{} "Retorna a lista de clientes paginada e informações de paginação"
// @Failure 401 {object} map[string]string "Token inválido ou não fornecido"
// @Failure 500 {object} map[string]string "Erro interno ao buscar ou processar os dados"
// @Router /api/clients-table [get]
// GetClientsTable retorna clientes paginados e filtrados para o DataTable, incluindo status online e expiração
func GetClientsTable(c *gin.Context) {
	// 📌 Extrair `member_id` do token
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token não fornecido"})
		return
	}

	claims, _, err := utils.ValidateToken(tokenString)
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

	// 📌 Obtém o parâmetro `online` (true/false)
	onlineFilter, _ := strconv.ParseBool(c.DefaultQuery("online", "false"))

	// 📌 Obtém o parâmetro `expiration_filter` (dias até expiração ou `-1` para vencidos)
	expirationFilter, _ := strconv.Atoi(c.DefaultQuery("expiration_filter", "0"))

	// 📌 Obtém o parâmetro `franquia_member_id` para filtro
	franquiaMemberIDFilterStr := c.Query("franquia_member_id")
	var franquiaMemberIDFilter sql.NullInt64
	if franquiaMemberIDFilterStr != "" {
		fmID, err := strconv.ParseInt(franquiaMemberIDFilterStr, 10, 64)
		if err == nil {
			franquiaMemberIDFilter.Int64 = fmID
			franquiaMemberIDFilter.Valid = true
		} else {
			log.Printf("⚠️ Aviso: Valor inválido para o filtro franquia_member_id: %s", franquiaMemberIDFilterStr)
			// Considerar retornar um erro 400 se o valor for inválido e o filtro for crucial
		}
	}

	// 📌 Parâmetros de paginação e pesquisa
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	offset := (page - 1) * limit

	// 📌 Obtém status online de todos os usuários ANTES da paginação
	onlineStatuses, err := getAllUsersOnlineStatus(memberID) // TODO: Revisar se este memberID é o correto para buscar status online quando filtrando por franquia_member_id
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar status online"})
		return
	}

	// 📌 Consulta base para buscar todos os usuários do membro
	query := `SELECT id, username, password, exp_date, enabled, admin_enabled, max_connections, created_at, reseller_notes, is_trial, Aplicativo, franquia_member_id 
			FROM users WHERE member_id = ? and deleted != '1'`
	var args []interface{}
	args = append(args, memberID)

	// 📌 Aplica filtro de pesquisa
	if search != "" {
		query += ` AND (username LIKE ? OR reseller_notes LIKE ?)`
		args = append(args, "%"+search+"%", "%"+search+"%")
	}
	// 📌 Aplica filtro `is_trial=0` ou `is_trial=1` (caso informado)
	isTrialFilter := c.Query("is_trial")
	if isTrialFilter != "" {
		query += ` AND is_trial = ?`
		args = append(args, isTrialFilter)
	}

	// 📌 Aplica filtro de franquia_member_id (se fornecido)
	if franquiaMemberIDFilter.Valid {
		query += ` AND franquia_member_id = ?`
		args = append(args, franquiaMemberIDFilter.Int64)
	}

	// 📌 Ordenação antes da paginação
	query += " ORDER BY created_at DESC"

	// 📌 Executa busca de todos os usuários, sem paginação inicial
	rows, err := config.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar clientes"})
		return
	}
	defer rows.Close()

	var allClients []models.ClientTableData

	// 🔄 Processa todos os usuários antes de paginar
	for rows.Next() {
		var client models.ClientTableData
		var expDate, createdAt, resellerNotes sql.NullString
		var aplicativo sql.NullString
		var franquiaMemberIDScanned sql.NullInt64 // Variável para escanear o franquia_member_id

		if err := rows.Scan(
			&client.ID, &client.Username, &client.Password, &expDate, &client.Enabled,
			&client.AdminEnabled, &client.MaxConnections, &createdAt, &resellerNotes, &client.IsTrial,
			&aplicativo, &franquiaMemberIDScanned, // Adicionado para scan
		); err != nil {
			log.Printf("❌ Erro ao escanear dados do cliente: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao processar os dados"})
			return
		}
		client.ExpDate = expDate
		client.CreatedAt = createdAt
		client.ResellerNotes = resellerNotes
		client.Aplicativo = aplicativo.String
		client.FranquiaMemberID = franquiaMemberIDScanned // Atribui o valor escaneado

		// 📌 Associa status online se existir
		if status, exists := onlineStatuses[client.ID]; exists {
			client.Online = map[string]interface{}{
				"username":            status.Username,
				"stream_display_name": status.StreamDisplayName,
				"date_start":          status.DateStart,
				"tempo_online":        status.TempoOnline,
				"user_agent":          status.UserAgent,
				"user_ip":             status.UserIP,
				"container":           status.Container,
				"geoip_country_code":  status.GeoIPCountryCode,
				"isp":                 status.ISP,
				"city":                status.City,
				"divergence":          status.Divergence,
				"stream_icon":         status.StreamIcon,
			}
		} else {
			client.Online = map[string]interface{}{} // 🔹 Retorna `{}` se não estiver online
		}

		allClients = append(allClients, client)
	}

	// 📌 Filtro `expiration_filter`
	if expirationFilter != 0 { // 🔹 Só filtra se `expiration_filter` for diferente de 0
		filteredClients := make([]models.ClientTableData, 0, len(allClients))
		currentTime := time.Now().Unix()
		expirationDays := int64(expirationFilter) * 86400 // 🔹 Converte dias para segundos

		for _, client := range allClients {
			expDateInt, err := strconv.ParseInt(client.ExpDate.String, 10, 64)
			if err != nil {
				expDateInt = 0 // 🔹 Se falhar, assume vencido
			}

			// 📌 Se `expiration_filter = -1` ou `0`, traz apenas vencidos (`exp_date < agora`)
			if (expirationFilter == -1 || expirationFilter == 0) && expDateInt < currentTime {
				filteredClients = append(filteredClients, client)
			} else if expirationFilter > 0 && expDateInt >= currentTime && expDateInt <= (currentTime+expirationDays) {
				filteredClients = append(filteredClients, client) // 🔹 Próximos X dias
			}
		}
		allClients = filteredClients // 🔹 Substitui a lista original
	}

	// 📌 Filtro `online=true`
	if onlineFilter {
		filteredClients := make([]models.ClientTableData, 0, len(allClients))
		for _, client := range allClients {
			if _, exists := onlineStatuses[client.ID]; exists {
				filteredClients = append(filteredClients, client)
			}
		}
		allClients = filteredClients
	}

	// 📌 Atualiza total de registros após filtros
	total := len(allClients)
	totalPages := (total + limit - 1) / limit

	// 📌 Paginação final
	start := offset
	end := offset + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	// 📌 Retorno formatado
	c.JSON(http.StatusOK, gin.H{
		"total_paginas":   totalPages,
		"pagina_atual":    page,
		"total_registros": total,
		"clientes":        allClients[start:end],
	})
}

// getAllUsersOnlineStatus busca o status online de todos os clientes do membro
func getAllUsersOnlineStatus(memberID int) (map[int]models.OnlineStatusData, error) {
	query := "CALL getUserOnlineStatus(0, ?);"

	// 📡 Log da query sendo executada
	//log.Printf("📡 DEBUG: Executando Query: %s | Params: (0, %d)", query, memberID)

	rows, err := config.DB.Query(query, memberID)
	if err != nil {
		log.Printf("❌ ERRO ao executar a procedure: %v", err)
		return nil, err
	}
	defer rows.Close()

	onlineUsers := make(map[int]models.OnlineStatusData)

	for rows.Next() {
		var onlineData models.OnlineStatusData

		err := rows.Scan(
			&onlineData.Id,
			&onlineData.Username,
			&onlineData.StreamDisplayName,
			&onlineData.DateStart,
			&onlineData.TempoOnline,
			&onlineData.UserAgent,
			&onlineData.UserIP,
			&onlineData.Container,
			&onlineData.GeoIPCountryCode,
			&onlineData.ISP,
			&onlineData.City,
			&onlineData.Divergence,
			&onlineData.StreamIcon,
		)

		if err != nil {
			log.Printf("❌ ERRO ao escanear os dados retornados: %v", err)
			return nil, err
		}

		// 📡 Log de cada usuário online retornado
		//log.Printf("✅ Usuário Online Carregado: ID %d | Canal: %s | IP: %s", onlineData.Id, onlineData.StreamDisplayName, onlineData.UserIP)

		onlineUsers[onlineData.Id] = onlineData
	}

	// 📡 Log final do total de usuários online carregados
	//log.Printf("🔍 Total de usuários online carregados: %d", len(onlineUsers))

	return onlineUsers, nil
}
