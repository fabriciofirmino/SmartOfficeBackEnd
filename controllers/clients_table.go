package controllers

import (
	"apiBackEnd/config"
	"apiBackEnd/models"
	"apiBackEnd/utils"
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
// @Param expiration_filter query int false "Filtrar clientes por vencimento (7, 15, 30, custom até 90 ou '0' para vencidos)"
// @Success 200 {object} map[string]interface{} "Retorna a lista de clientes paginada e informações de paginação"
// @Failure 401 {object} map[string]string "Token inválido ou não fornecido"
// @Failure 500 {object} map[string]string "Erro interno ao buscar ou processar os dados"
// @Router /api/clients-table [get]
// GetClientsTable retorna clientes paginados e filtrados para o DataTable, incluindo status online e expiração
func GetClientsTable(c *gin.Context) {
	// 📌 Extrai `member_id` do token
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

	// 📌 Obtém os parâmetros da requisição
	onlineFilter := c.DefaultQuery("online", "") == "true"
	search := c.Query("search")
	expirationFilter, _ := strconv.Atoi(c.DefaultQuery("expiration_filter", "-1")) // -1 significa "sem filtro"

	// 📌 Limita o valor máximo do filtro de expiração para 90 dias
	if expirationFilter > 90 {
		expirationFilter = 90
	}

	// 📌 Define a condição de expiração
	currentTime := time.Now().Unix()
	var expirationCondition string
	var expirationArgs []interface{}

	if expirationFilter > 0 {
		expirationCondition = "AND exp_date BETWEEN ? AND ?"
		expirationArgs = append(expirationArgs, currentTime, currentTime+int64(expirationFilter*86400))
	} else if expirationFilter == 0 { // Apenas vencidos
		expirationCondition = "AND exp_date < ?"
		expirationArgs = append(expirationArgs, currentTime)
	}

	// 📌 Parâmetros de paginação
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	offset := (page - 1) * limit

	// 📌 Obtém status online de **todos os usuários do membro**
	onlineStatuses, err := getAllUsersOnlineStatus(memberID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar status online"})
		return
	}

	// 📌 Query base para buscar usuários já com todos os filtros aplicados
	query := `SELECT id, username, password, exp_date, enabled, admin_enabled, max_connections, created_at, reseller_notes, is_trial 
			FROM users WHERE member_id = ? ` + expirationCondition
	var args []interface{}
	args = append(args, memberID)
	args = append(args, expirationArgs...)

	// 📌 Aplica filtro de pesquisa
	if search != "" {
		query += ` AND (username LIKE ? OR reseller_notes LIKE ?)`
		args = append(args, "%"+search+"%", "%"+search+"%")
	}

	// 📌 Mantém a ordenação
	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	// 📌 Executa busca de clientes com paginação
	rows, err := config.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar clientes"})
		return
	}
	defer rows.Close()

	var clients []models.ClientTableData

	// 🔄 Processa os clientes aplicando filtros e status online
	for rows.Next() {
		var client models.ClientTableData
		if err := rows.Scan(
			&client.ID, &client.Username, &client.Password, &client.ExpDate, &client.Enabled,
			&client.AdminEnabled, &client.MaxConnections, &client.CreatedAt, &client.ResellerNotes, &client.IsTrial,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao processar os dados"})
			return
		}

		// 📌 Associa status online
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

		clients = append(clients, client)
	}

	// 📌 Contagem total de registros (corrigida)
	var total int
	countQuery := `SELECT COUNT(*) FROM users WHERE member_id = ? ` + expirationCondition
	countArgs := []interface{}{memberID}
	countArgs = append(countArgs, expirationArgs...)

	config.DB.QueryRow(countQuery, countArgs...).Scan(&total)

	// 📌 Filtragem final se `online=true`
	if onlineFilter {
		filteredClients := make([]models.ClientTableData, 0, len(clients))
		for _, client := range clients {
			if _, exists := onlineStatuses[client.ID]; exists {
				filteredClients = append(filteredClients, client)
			}
		}
		clients = filteredClients
		total = len(clients) // Atualiza a contagem total após filtrar os online
	}

	totalPages := (total + limit - 1) / limit

	// 📌 Retorno formatado
	c.JSON(http.StatusOK, gin.H{
		"total_paginas":   totalPages,
		"pagina_atual":    page,
		"total_registros": total,
		"clientes":        clients,
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
