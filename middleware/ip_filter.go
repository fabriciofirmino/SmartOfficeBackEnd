package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// IPFilterMiddleware verifica se o IP do cliente está na lista de IPs permitidos.
func IPFilterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		allowedIPsEnv := os.Getenv("ALLOWED_IPS")
		if allowedIPsEnv == "" {
			// Se ALLOWED_IPS estiver vazio, permite todos os IPs
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		allowedIPs := strings.Split(allowedIPsEnv, ",")

		isAllowed := false
		for _, ip := range allowedIPs {
			trimmedIP := strings.TrimSpace(ip)
			if trimmedIP == clientIP {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			log.Printf("IP Bloqueado: %s (Não está na lista ALLOWED_IPS: %s)", clientIP, allowedIPsEnv)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Acesso negado: IP não permitido."})
			return
		}

		log.Printf("IP Permitido: %s", clientIP)
		c.Next()
	}
}
