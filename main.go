package main

import (
	"apiBackEnd/config" // Certifique-se de que o nome do módulo no go.mod é "apiBackEnd"
	_ "apiBackEnd/docs" // 🔥 Importação do Swagger
	"apiBackEnd/middleware"
	"apiBackEnd/routes"
	"fmt"

	"github.com/gin-gonic/gin" // ✅ Correto
)

// @title API IPTV
// @version 1.0
// @description Documentação da API IPTV
// @termsOfService http://example.com/terms/

// @contact.name Suporte
// @contact.email suporte@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Conectar ao banco de dados
	config.ConnectDB()

	// Criar servidor
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())
	routes.SetupRoutes(r)

	fmt.Println("Servidor rodando em http://localhost:8080")
	r.Run(":8080")
}
