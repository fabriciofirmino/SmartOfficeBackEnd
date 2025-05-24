package models

import (
	"apiBackEnd/config"
	"apiBackEnd/utils"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware valida o token JWT nas rotas protegidas
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token não fornecido"})
			c.Abort()
			return
		}

		// 📌 Captura os três valores retornados por `ValidateToken`
		_, _, err := utils.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token inválido"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Estrutura do usuário retornado do banco
type User struct {
	Username      string
	PasswordHash  string
	MemberGroupID int
	Credits       float64
	Status        int
	MemberID      int // 🔥 Corrigido: Agora "MemberID" com "D" maiúsculo
}

// Buscar usuário pelo username
func GetUserByUsername(username string) (*User, error) {
	var user User

	allowedUsers := os.Getenv("ALLOWED_USERS")
	baseQuery := `
		SELECT username, password, ru.member_group_id, credits, status, id as member_id
		FROM streamcreed_db.reg_users AS ru
		WHERE username = ?
	`
	args := []interface{}{username}

	if allowedUsers != "" {
		// Constrói placeholders "?" para cada usuário permitido
		list := strings.Split(allowedUsers, ",")
		placeholders := []string{}
		for _, usr := range list {
			placeholders = append(placeholders, "?")
			args = append(args, strings.TrimSpace(usr))
		}
		baseQuery += " AND ru.username IN (" + strings.Join(placeholders, ",") + ")"
	}

	err := config.DB.QueryRow(baseQuery, args...).Scan(
		&user.Username, &user.PasswordHash, &user.MemberGroupID, &user.Credits, &user.Status, &user.MemberID,
	)

	if err != nil {
		log.Println("Erro ao buscar usuário:", err)
		return nil, err
	}

	return &user, nil
}
