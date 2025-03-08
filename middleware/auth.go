package middleware

import (
	"apiBackEnd/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware verifica o token JWT e impede uso de tokens prestes a expirar
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token não fornecido"})
			c.Abort()
			return
		}

		claims, timeRemaining, err := utils.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token inválido ou expirado"})
			c.Abort()
			return
		}

		if timeRemaining < 5 {
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token expirando, faça login novamente"})
			c.Abort()
			return
		}

		c.Set("claims", claims)
		c.Next()
	}
}
