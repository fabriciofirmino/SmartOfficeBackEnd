package controllers

import (
	"apiBackEnd/config"
	"apiBackEnd/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetCredits retorna o total de créditos atualizado e o tempo restante do token.
//
// @Summary Obtém créditos atualizados e tempo restante do token
// @Description Retorna o total de créditos do usuário autenticado e o tempo restante do token em segundos.
// @Tags Créditos
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "Dados de créditos e tempo restante"
// @Failure 401 {object} map[string]string "Token inválido ou expirado"
// @Failure 500 {object} map[string]string "Erro ao buscar créditos"
// @Router /api/credits [get]
func GetCredits(c *gin.Context) {
	// 📌 Recuperar token do header
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token não fornecido"})
		return
	}

	// 📌 Validar token e extrair claims + tempo restante
	claims, timeRemaining, err := utils.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token inválido ou expirado"})
		return
	}

	// 📌 Extrair `member_id` do token
	memberIDFloat, exists := claims["member_id"].(float64)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "MemberID não encontrado no token"})
		return
	}
	memberID := int(memberIDFloat)

	// 🔹 **Buscar créditos do banco em tempo real**
	var creditos int
	query := "SELECT credits FROM streamcreed_db.reg_users WHERE id = ?"
	err = config.DB.QueryRow(query, memberID).Scan(&creditos)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar créditos"})
		return
	}

	// 🔹 🔥 Retorno atualizado
	c.JSON(http.StatusOK, gin.H{
		"creditos":        creditos,      // Agora sempre atualizado ✅
		"token_expira_em": timeRemaining, // Tempo do token em segundos
	})
}
