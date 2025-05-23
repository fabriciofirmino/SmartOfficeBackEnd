package controllers

import (
	"apiBackEnd/models"
	"apiBackEnd/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware valida o token JWT nas rotas protegidas.
// Esta função agora reside em controllers/auth.go
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token não fornecido"})
			c.Abort()
			return
		}

		claims, _, err := utils.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token inválido ou expirado"})
			c.Abort()
			return
		}

		if claims != nil {
			c.Set("claims", claims)
			if memberID, ok := claims["member_id"].(float64); ok {
				c.Set("member_id", int(memberID))
			}
			if username, ok := claims["username"].(string); ok {
				c.Set("username", username)
			}
		}

		c.Next()
	}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Estrutura de resposta do login para o Swagger
type LoginResponse struct {
	Token         string  `json:"token"`
	MemberGroupID int     `json:"member_group_id"`
	Credits       float64 `json:"credits"`
	Status        int     `json:"status"`
	MemberID      int     `json:"member_id"`
}

// Login realiza a autenticação do usuário.
//
// @Summary Autenticação de Usuário
// @Description Autentica um usuário e retorna um token JWT se as credenciais forem válidas.
// @Tags Autenticação
// @Accept  json
// @Produce  json
// @Param login body LoginRequest true "Credenciais de login"
// @Success 200 {object} LoginResponse "Login realizado com sucesso"
// @Failure 400 {object} map[string]string "Erro na requisição"
// @Failure 401 {object} map[string]string "Credenciais inválidas"
// @Router /login [post]
func Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos"})
		return
	}

	user, err := models.GetUserByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Usuário ou senha incorretos"})
		return
	}

	// **📌 Bloquear se a conta não estiver ativa**
	if user.Status != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Conta bloqueada. Entre em contato com o suporte."})
		return
	}

	hashedInputPassword, err := utils.CryptPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao processar senha"})
		return
	}

	if hashedInputPassword != user.PasswordHash {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Usuário ou senha incorretos"})
		return
	}

	// 📌 Gerar token com tempo de expiração configurável
	token, err := utils.GenerateToken(user.Username, user.MemberID, user.Credits, strconv.Itoa(user.Status))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao gerar token"})
		return
	}

	// 📌 Retornar os dados do usuário no login
	c.JSON(http.StatusOK, LoginResponse{
		Token:         token,
		MemberGroupID: user.MemberGroupID,
		Credits:       user.Credits,
		Status:        user.Status,
		MemberID:      user.MemberID,
	})
}
