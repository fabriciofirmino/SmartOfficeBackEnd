package config

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

// InitRedis inicializa a conexão com o Redis
func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),     // Exemplo: "localhost:6379"
		Username: os.Getenv("REDIS_USER"),     // Exemplo: "admin" (usuário configurado)
		Password: os.Getenv("REDIS_PASSWORD"), // Se não tiver senha, deixe ""
		DB:       0,
	})

	// Testar conexão
	ctx := context.Background()
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		fmt.Println("❌ Erro ao conectar ao Redis:", err)
	} else {
		fmt.Println("✅ Conexão com Redis estabelecida!")
	}
}
