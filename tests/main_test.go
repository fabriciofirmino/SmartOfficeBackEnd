package tests

import (
	"apiBackEnd/config"
	"apiBackEnd/middleware"
	"apiBackEnd/routes"
	"log"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func SetupServer() *gin.Engine {
	// Carregar variáveis de ambiente
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ Erro ao carregar .env (seguindo com valores padrão)")
	}

	// Conectar ao banco de dados
	config.ConnectDB()
	config.InitRedis()

	// Criar servidor
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())
	routes.SetupRoutes(r)

	return r
}

// TestMain inicializa as configurações antes dos testes
func TestMain(m *testing.M) {
	log.Println("🔧 Inicializando ambiente de testes...")

	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ Erro ao carregar .env (seguindo com valores padrão)")
	}

	log.Println("✅ Testes prontos para rodar...")
	exitCode := m.Run()

	// 🔹 Exibir resumo final
	printTestSummary()

	log.Println("✅ Finalizando os testes...")
	os.Exit(exitCode)
}
