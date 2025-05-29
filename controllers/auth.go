package controllers

import (
	"apiBackEnd/models"
	"apiBackEnd/utils"
	"log"
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
	log.Println("INFO: Iniciando requisição de login") // Log adicionado

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("ERRO: Dados de login inválidos: %v", err) // Log adicionado
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos"})
		return
	}
	log.Printf("INFO: Tentativa de login para o usuário: %s", req.Username) // Log adicionado

	user, err := models.GetUserByUsername(req.Username)
	if err != nil {
		log.Printf("ERRO: Usuário '%s' não encontrado ou erro ao buscar: %v", req.Username, err) // Log adicionado
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Usuário ou senha incorretos"})
		return
	}
	log.Printf("INFO: Usuário '%s' encontrado. Status: %d, MemberID: %d", user.Username, user.Status, user.MemberID) // Log adicionado

	// **📌 Bloquear se a conta não estiver ativa**
	if user.Status != 1 {
		log.Printf("AVISO: Tentativa de login para usuário '%s' com conta bloqueada (Status: %d)", user.Username, user.Status) // Log adicionado
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Conta bloqueada. Entre em contato com o suporte."})
		return
	}

	hashedInputPassword, err := utils.CryptPassword(req.Password)
	if err != nil {
		log.Printf("ERRO: Falha ao gerar hash da senha para o usuário '%s': %v", req.Username, err) // Log adicionado
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao processar senha"})
		return
	}

	if hashedInputPassword != user.PasswordHash {
		log.Printf("AVISO: Senha incorreta para o usuário '%s'", req.Username) // Log adicionado
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Usuário ou senha incorretos"})
		return
	}

	log.Printf("INFO: Credenciais válidas para o usuário '%s'. Gerando token...", user.Username) // Log adicionado
	// 📌 Gerar token com tempo de expiração configurável
	token, err := utils.GenerateToken(user.Username, user.MemberID, user.Credits, strconv.Itoa(user.Status))
	if err != nil {
		log.Printf("ERRO: Falha ao gerar token para o usuário '%s': %v", user.Username, err) // Log adicionado
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao gerar token"})
		return
	}
	log.Printf("INFO: Token gerado com sucesso para o usuário '%s': %s", user.Username, token) // Log adicionado

	// 📌 Retornar os dados do usuário no login
	c.JSON(http.StatusOK, LoginResponse{
		Token:         token,
		MemberGroupID: user.MemberGroupID,
		Credits:       user.Credits,
		Status:        user.Status,
		MemberID:      user.MemberID,
	})
}
