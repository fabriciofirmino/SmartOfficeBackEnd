package controllers

import (
	"apiBackEnd/config"
	"apiBackEnd/models"
	"apiBackEnd/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetClientsTable retorna clientes paginados e filtrados para o DataTable
// @Summary Retorna clientes paginados e filtrados
// @Description Retorna uma lista de clientes paginada e filtrada para uso em DataTables, associados ao member_id do token.
// @Tags ClientsTable
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "Número da página (padrão: 1)"
// @Param limit query int false "Limite de registros por página (padrão: 10)"
// @Param search query string false "Termo de pesquisa para filtrar por username ou reseller_notes"
// @Success 200 {object} map[string]interface{} "Retorna a lista de clientes paginada e informações de paginação"
// @Failure 401 {object} map[string]string "Token inválido ou não fornecido"
// @Failure 500 {object} map[string]string "Erro interno ao buscar ou processar os dados"
// @Router /api/clients-table [get]
// GetClientsTable retorna clientes paginados e filtrados para o DataTable
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
	memberID := int(memberIDFloat) // 🔹 Converte para inteiro

	// 📌 Parâmetros de paginação e pesquisa
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search") // 🔹 Termo de pesquisa opcional

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	offset := (page - 1) * limit

	// 📌 Consulta base (Filtrando pelo `member_id`)
	query := `SELECT id, username, password, exp_date, enabled, admin_enabled, max_connections, created_at, reseller_notes, is_trial 
			  FROM users WHERE member_id = ?`
	var args []interface{}
	args = append(args, memberID) // 🔹 Garante que apenas os clientes desse membro são retornados

	// 📌 Adiciona filtro de pesquisa se necessário
	if search != "" {
		query += ` AND (username LIKE ? OR reseller_notes LIKE ?)`
		args = append(args, "%"+search+"%", "%"+search+"%")
	}

	// 📌 Paginação
	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar clientes"})
		return
	}
	defer rows.Close()

	var clients []models.ClientTableData
	for rows.Next() {
		var client models.ClientTableData
		if err := rows.Scan(
			&client.ID, &client.Username, &client.Password, &client.ExpDate, &client.Enabled,
			&client.AdminEnabled, &client.MaxConnections, &client.CreatedAt, &client.ResellerNotes, &client.IsTrial,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao processar os dados"})
			return
		}
		clients = append(clients, client)
	}

	// 📌 Contagem total de registros filtrados pelo `member_id`
	var total int
	countQuery := `SELECT COUNT(*) FROM users WHERE member_id = ?`
	countArgs := []interface{}{memberID}

	// 📌 Aplica filtro de pesquisa na contagem
	if search != "" {
		countQuery += ` AND (username LIKE ? OR reseller_notes LIKE ?)`
		countArgs = append(countArgs, "%"+search+"%", "%"+search+"%")
	}

	config.DB.QueryRow(countQuery, countArgs...).Scan(&total)

	totalPages := (total + limit - 1) / limit // 🔹 Calcula total de páginas

	// 📌 Retorno formatado
	c.JSON(http.StatusOK, gin.H{
		"total_paginas":   totalPages,
		"pagina_atual":    page,
		"total_registros": total,
		"clientes":        clients,
	})
}
