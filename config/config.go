package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql" // Driver MySQL/MariaDB
	"github.com/joho/godotenv"
)

var DB *sql.DB

func init() {
	godotenv.Load(".env")
}

func ConnectDB() {
	// Carregar variáveis do .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Erro ao carregar .env:", err)
	}

	// Ler variáveis de ambiente
	user := strings.TrimSpace(os.Getenv("DB_USER"))
	password := strings.TrimSpace(os.Getenv("DB_PASSWORD"))
	host := strings.TrimSpace(os.Getenv("DB_HOST"))
	port := strings.TrimSpace(os.Getenv("DB_PORT"))
	database := strings.TrimSpace(os.Getenv("DB_NAME"))

	// DSN para MariaDB/MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, database)
	//log criação da string de conexão fmt.Println("DSN:", dsn)

	// Conectar ao banco
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Erro ao conectar ao banco:", err)
	}

	// Testar conexão
	err = DB.Ping()
	if err != nil {
		log.Fatal("Erro ao verificar conexão:", err)
	}

	fmt.Println("✅ Conexão com MariaDB estabelecida com sucesso!")
}
