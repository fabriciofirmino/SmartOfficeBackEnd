package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB é a instância global para conexão
var MongoDB *mongo.Client

// InitMongo inicializa a conexão com o MongoDB
func InitMongo() {
	// Lê a string de conexão do MongoDB a partir da variável de ambiente
	mongoURI := os.Getenv("MONGO_DB")
	if mongoURI == "" {
		log.Fatal("❌ ERRO: Variável de ambiente MONGO_DB não encontrada!")
	}

	// Configura a Stable API (ServerAPIOptions) – opcional, mas recomendado
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	clientOpts := options.Client().
		ApplyURI(mongoURI).
		SetServerAPIOptions(serverAPI)

	// Conecta ao MongoDB
	client, err := mongo.Connect(context.Background(), clientOpts)
	if err != nil {
		log.Fatalf("❌ ERRO ao criar cliente MongoDB: %v", err)
	}

	// Cria um contexto com timeout para o ping (evita travar se algo der errado)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Testa a conexão com um "ping"
	err = client.Database("admin").RunCommand(ctx, bson.D{{Key: "ping", Value: 1}}).Err()
	if err != nil {
		log.Fatalf("❌ ERRO ao fazer ping no MongoDB: %v", err)
	}

	// Conexão bem-sucedida – atribui ao global
	MongoDB = client
	fmt.Println("✅ Conectado ao MongoDB com sucesso!")
}
