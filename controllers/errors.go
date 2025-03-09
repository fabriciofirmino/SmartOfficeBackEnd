package controllers

import (
	"apiBackEnd/config"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Estrutura para representar um erro do usuário
type ErrorResponse struct {
	TotalPaginas int         `json:"total_paginas"`
	PaginaAtual  int         `json:"pagina_atual"`
	TotalErros   int         `json:"total_erros"`
	Erros        []UserError `json:"erros"`
}

type UserError struct {
	NomeCanal     string `json:"nome_canal"`
	Usuario       string `json:"usuario"`
	MensagemErro  string `json:"mensagem_erro"`
	UserAgent     string `json:"user_agent"`
	IP            string `json:"ip"`
	DataFormatada string `json:"data_formatada"`
}

// GetUserErrors retorna os erros associados a um usuário específico, com paginação.
//
// @Summary Detalhes dos erros do usuário com paginação
// @Description Retorna os erros registrados na conta de um usuário, incluindo IP, dispositivo e motivo do erro.
// @Tags Erros
// @Security BearerAuth
// @Accept  json
// @Produce  json
// @Param id_usuario path int true "ID do usuário"
// @Param limit query int false "Número de registros por página (padrão: 10)"
// @Param page query int false "Número da página (padrão: 1)"
// @Success 200 {object} map[string]interface{} "Lista paginada de erros do usuário"
// @Failure 400 {object} map[string]string "ID inválido"
// @Failure 404 {object} map[string]string "Nenhum erro encontrado"
// @Failure 500 {object} map[string]string "Erro interno"
// @Router /api/details-error/{id_usuario} [get]
func GetUserErrors(c *gin.Context) {
	// 📌 Capturar o ID do usuário da URL
	userIDStr := c.Param("id_usuario")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "ID do usuário inválido"})
		return
	}

	// 📌 Capturar parâmetros de paginação da URL (com valores padrão)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	// 📌 Query para contar o total de erros do usuário
	var totalErros int
	countQuery := "SELECT COUNT(*) FROM streamcreed_db.client_logs WHERE user_id = ? AND client_status IS NOT NULL"
	err = config.DB.QueryRow(countQuery, userID).Scan(&totalErros)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao contar os erros do usuário"})
		return
	}

	// 📌 Se não houver erros, retornar mensagem amigável
	if totalErros == 0 {
		c.JSON(http.StatusNotFound, gin.H{"mensagem": "Nenhum erro encontrado para este usuário."})
		return
	}

	// 📌 Calcular total de páginas
	totalPages := int(math.Ceil(float64(totalErros) / float64(limit)))

	// 📌 Query para buscar erros paginados
	query := `
		SELECT
			s.stream_display_name AS nome_canal,
			u.username AS usuario,
			CASE cl.client_status
				WHEN 'COUNTRY_DISALLOW' THEN 'Erro: país não permitido'
				WHEN 'USER_EXPIRED' THEN 'Erro: usuário expirado'
				WHEN 'USER_DISABLED' THEN 'Erro: usuário desativado'
				WHEN 'USER_AGENT_BAN' THEN 'Erro: acesso bloqueado pelo agente do usuário'
				WHEN 'USER_BAN' THEN 'Erro: usuário banido'
				ELSE 'Erro desconhecido'
			END AS mensagem_erro,
			cl.user_agent,
			cl.ip,
			DATE_FORMAT(FROM_UNIXTIME(cl.date), '%d/%m/%Y %H:%i:%s') AS data_formatada
		FROM
			streamcreed_db.client_logs AS cl
		INNER JOIN streams AS s ON s.id = cl.stream_id
		INNER JOIN users AS u ON u.id = cl.user_id
		WHERE
			cl.user_id = ?
			AND cl.client_status IS NOT NULL
		ORDER BY cl.id DESC
		LIMIT ? OFFSET ?
	`

	// 📌 Executar consulta no banco
	rows, err := config.DB.Query(query, userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar erros do usuário"})
		return
	}
	defer rows.Close()

	// 📌 Processar resultados
	var erros []UserError
	for rows.Next() {
		var userError UserError
		if err := rows.Scan(&userError.NomeCanal, &userError.Usuario, &userError.MensagemErro,
			&userError.UserAgent, &userError.IP, &userError.DataFormatada); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao processar resultados"})
			return
		}
		erros = append(erros, userError)
	}

	response := ErrorResponse{
		TotalPaginas: totalPages,
		PaginaAtual:  page,
		TotalErros:   totalErros,
		Erros:        erros, // Agora garantimos a ordem correta!
	}

	c.JSON(http.StatusOK, response)
}
