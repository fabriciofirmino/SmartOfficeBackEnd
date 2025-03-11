package controllers

import (
	"apiBackEnd/config"
	"apiBackEnd/models"
	"apiBackEnd/utils"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

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

	// 📌 Contagem total de registros filtrados pelo `member_id`
	var total int
	countQuery := `SELECT COUNT(*) FROM users WHERE member_id = ?`
	countArgs := []interface{}{memberID}

	// 📌 Aplica filtro de pesquisa na contagem
	if search != "" {
		countQuery += ` AND (username LIKE ? OR reseller_notes LIKE ?)`
		countArgs = append(countArgs, "%"+search+"%", "%"+search+"%")
	}

	err = config.DB.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		fmt.Println("❌ Erro ao contar registros:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao contar registros"})
		return
	}

	// 📌 Calcula total de páginas corretamente
	totalPages := (total + limit - 1) / limit

	// **🔹 Debug para verificar os cálculos**
	//fmt.Println("📊 DEBUG PAGINAÇÃO:")
	//fmt.Println("🔹 Total de registros:", total)
	//fmt.Println("🔹 Limite por página:", limit)
	//fmt.Println("🔹 Total de páginas calculadas:", totalPages)
	//fmt.Println("🔹 Página solicitada:", page)

	// **🔹 Ajuste da página para evitar erro**
	if totalPages == 0 {
		totalPages = 1 // Evita divisão por zero
	}
	if page > totalPages {
		page = totalPages // 🔹 Ajusta para última página disponível
	}

	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	fmt.Println("🔹 Offset calculado:", offset) // Verificando o valor final de offset

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

	// 📌 Paginação segura
	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		fmt.Println("❌ Erro ao buscar clientes:", err)
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
			fmt.Println("❌ Erro ao processar os dados:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao processar os dados"})
			return
		}

		// 📌 Convertendo NULL para string vazia
		if !client.ExpDate.Valid {
			client.ExpDate.String = ""
		}
		if !client.ResellerNotes.Valid {
			client.ResellerNotes.String = ""
		}

		clients = append(clients, client)
	}

	// 📌 Retorno formatado
	c.JSON(http.StatusOK, gin.H{
		"total_paginas":   totalPages,
		"pagina_atual":    page,
		"total_registros": total,
		"clientes":        clients,
	})
}
