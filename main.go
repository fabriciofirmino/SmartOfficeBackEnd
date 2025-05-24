package main

import (
	"apiBackEnd/config"
	"apiBackEnd/middleware"
	"apiBackEnd/routes"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// @title API IPTV
// @version 1.0.5
// @description Documentação da API IPTV

// @contact.name Suporte
// @contact.email suporte@example.com

// @host localhost:443
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// SetupServer inicializa e retorna o router do Gin (para uso nos testes)
func SetupServer() *gin.Engine {
	// Carregar variáveis de ambiente
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ Erro ao carregar .env (seguindo com valores padrão)")
	}

	// Conectar ao banco de dados
	config.ConnectDB()
	config.InitRedis()
	config.InitMongo()

	// Criar servidor
	r := gin.Default()

	// Aplicar Middlewares Globais
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.IPFilterMiddleware()) // Aplicar o filtro de IP

	routes.SetupRoutes(r)

	return r
}

func main() {
	// Inicializar o servidor real
	r := SetupServer()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Porta padrão
	}
	log.Printf("🚀 Servidor rodando na porta %s", port)
	r.Run("0.0.0.0:" + port)
}
