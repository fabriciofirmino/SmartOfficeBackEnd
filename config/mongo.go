package config

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoDB *mongo.Client

func InitMongo() {
	mongoURI := os.Getenv("MONGO_DB")
	if mongoURI == "" {
		log.Fatal("❌ ERRO: Variável de ambiente MONGO_DB não encontrada!")
	}

	// Configuração TLS mais robusta
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12, // Força TLS 1.2 ou superior
	}

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	clientOpts := options.Client().
		ApplyURI(mongoURI).
		SetTLSConfig(tlsConfig).
		SetServerAPIOptions(serverAPI).
		SetConnectTimeout(30 * time.Second).        // Tempo maior para conexão
		SetSocketTimeout(60 * time.Second).         // Tempo maior para operações
		SetServerSelectionTimeout(30 * time.Second) // Tempo para selecionar servidor

	// Tentativa de conexão com retry
	var client *mongo.Client
	var err error
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		client, err = mongo.Connect(context.Background(), clientOpts)
		if err == nil {
			break
		}
		log.Printf("⚠️ Tentativa %d/%d falhou: %v", i+1, maxRetries, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("❌ ERRO ao criar cliente MongoDB após %d tentativas: %v", maxRetries, err)
	}

	// Contexto com timeout maior para o ping
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.Database("admin").RunCommand(ctx, bson.D{{Key: "ping", Value: 1}}).Err()
	if err != nil {
		log.Fatalf("❌ ERRO ao fazer ping no MongoDB: %v", err)
	}

	MongoDB = client
	fmt.Println("✅ Conectado ao MongoDB com sucesso!")
}
