package controllers

import (
	"apiBackEnd/config"
	"apiBackEnd/utils"
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Logout encerra a sessão do usuário, invalidando o token no Redis.
//
// @Summary Logout do Usuário
// @Description Remove o token ativo do usuário, invalidando sua sessão.
// @Tags Logout
// @Security BearerAuth
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]string "Logout realizado com sucesso"
// @Failure 401 {object} map[string]string "Token inválido ou não autorizado"
// @Router /logout [post]
func Logout(c *gin.Context) {
	ctx := context.Background()

	// 📌 **Recuperar o token do header**
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token não fornecido"})
		return
	}

	// 📌 **Validar token**
	claims, _, err := utils.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token inválido ou expirado"})
		return
	}

	// 📌 **Obter `member_id` do token**
	memberID, ok := claims["member_id"].(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "MemberID não encontrado no token"})
		return
	}

	redisKey := "token:" + strconv.Itoa(int(memberID))

	// 📌 **Remover o token do Redis**
	err = config.RedisClient.Del(ctx, redisKey).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao invalidar token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "Logout realizado com sucesso"})
}
