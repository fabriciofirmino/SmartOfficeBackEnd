package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// BasicAuth é um middleware para proteger rotas com autenticação HTTP Basic.
// Ele compara o usuário e senha fornecidos com os valores configurados.
func BasicAuth(expectedUser, expectedPass string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, pass, hasAuth := c.Request.BasicAuth()

		// Verifica se as credenciais do .env foram carregadas
		// Se não foram, permite o acesso (para desenvolvimento local sem .env, por exemplo)
		// Ou pode optar por negar o acesso se as variáveis não estiverem setadas.
		// Neste caso, se não estiverem no .env, a rota não será protegida por este middleware
		// conforme a lógica em SetupRoutes.
		if expectedUser == "" || expectedPass == "" {
			c.Next()
			return
		}

		if !hasAuth || user != expectedUser || pass != expectedPass {
			c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		c.Next()
	}
}

// GinBasicAuth é um wrapper para BasicAuth que pode ser usado diretamente
// com as credenciais do ambiente.
func GinBasicAuth() gin.HandlerFunc {
	// Obtenha as credenciais do ambiente
	// Considere logar um aviso se não estiverem definidas, dependendo da sua política de segurança
	user := os.Getenv("SWAGGER_USER")
	pass := os.Getenv("SWAGGER_PASS")
	return BasicAuth(user, pass)
}
