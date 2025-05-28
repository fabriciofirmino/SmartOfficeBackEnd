package controllers

import (
	"apiBackEnd/models"
	"apiBackEnd/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Estrutura de resposta do dashboard
type DashboardResponse struct {
	TotalClientesRevenda int    `json:"totalClientesRevenda"`
	TotalTestesAtivos    int    `json:"totalTestesAtivos"`
	TotalVencido         int    `json:"totalVencido"`
	TotalClientes        int    `json:"totalClientes"`
	TotalFilmes          int    `json:"totalFilmes"`
	TotalEpisodiosSeries int    `json:"totalEpisodiosSeries"`
	TotalCanais          int    `json:"totalCanais"`
	TotalSeries          int    `json:"totalSeries"`
	CanaisOff            int    `json:"canaisOff"`
	LastUpdated          string `json:"lastUpdated"` // Ou sql.NullString se puder ser nulo
}

// DashboardHandler retorna métricas do dashboard com base no member_id.
//
// @Summary Obtém os dados do dashboard
// @Description Retorna os totais de clientes e testes ativos
// @Tags Dashboard
// @Security BearerAuth
// @Produce json
// @Success 200 {object} DashboardResponse "Dados do dashboard"
// @example response.200.success
//
//	{
//	  "totalClientesRevenda": 150,
//	  "totalTestesAtivos": 25,
//	  "totalVencido": 10,
//	  "totalClientes": 200,
//	  "totalFilmes": 5000,
//	  "totalEpisodiosSeries": 15000,
//	  "totalCanais": 300,
//	  "totalSeries": 200,
//	  "canaisOff": 5,
//	  "lastUpdated": "2024-05-27T10:30:00Z"
//	}
//
// @Failure 401 {object} map[string]string "Token inválido"
// @Failure 500 {object} map[string]string "Erro interno"
// @Router /api/dashboard [get]
func DashboardHandler(c *gin.Context) {
	// 📌 Recuperar o token e validar
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

	// 📌 Extrair `member_id` do token
	memberIDFloat, exists := claims["member_id"].(float64)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "MemberID não encontrado no token"})
		return
	}
	memberID := int(memberIDFloat)

	// 📌 Executar as procedures e obter os counts
	counts, err := models.ObterDadosDashboard(memberID) // ✅ Corrigido
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar dados do dashboard"})
		return
	}

	// 📌 Retornar os dados no JSON
	c.JSON(http.StatusOK, counts)
}
