package controllers

import (
	"apiBackEnd/models"
	"apiBackEnd/utils"
	"net/http"

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

		_, err := utils.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token inválido"})
			c.Abort()
			return
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
// @Success 200 {object} map[string]interface{} "Login realizado com sucesso"
// @Failure 400 {object} map[string]string "Erro na requisição"
// @Failure 401 {object} map[string]string "Credenciais inválidas"
// @Router /login [post]
func Login(c *gin.Context) {
	var req LoginRequest

	// 📌 Validação do JSON recebido
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos"})
		return
	}

	// 📌 Busca usuário no banco de dados
	user, err := models.GetUserByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Usuário ou senha incorretos"})
		return
	}

	// 📌 Validação da senha com hashing correto
	hashedInputPassword, err := utils.CryptPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao processar senha"})
		return
	}

	if hashedInputPassword != user.PasswordHash {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Usuário ou senha incorretos"})
		return
	}

	// 📌 Geração do token JWT
	token, err := utils.GenerateToken(user.Username, user.MemeberId) // Passando corretamente o MemberID
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao gerar token"})
		return
	}

	// 📌 Resposta JSON para o cliente
	c.JSON(http.StatusOK, LoginResponse{
		Token:         token,
		MemberGroupID: user.MemberGroupID,
		Credits:       user.Credits,
		Status:        user.Status,
		MemberID:      user.MemeberId,
	})
}
