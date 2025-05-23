package utils

import (
	"crypto/rand"
	"encoding/json"
	"math/big"
	"os"
	"strconv"
	"time"

	"apiBackEnd/config"
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// Gerar timestamp de expiração baseado em horas
func GenerateExpirationTimestamp(hours string) int64 {
	hoursInt, err := strconv.Atoi(hours)
	if err != nil {
		hoursInt = 24 // Padrão: 24h caso a conversão falhe
	}
	return time.Now().Add(time.Duration(hoursInt) * time.Hour).Unix()
}

// Gerar nome de usuário aleatório (prefixo + caracteres permitidos)
func GenerateUsername(length int, prefix string) string {
	charset := "abcdemuvwxyz0123456789" // 🔥 Apenas caracteres permitidos

	result := make([]byte, length)
	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[num.Int64()]
	}

	// 🔥 Retorna username formatado corretamente
	return prefix + string(result)
}

// Gerar senha aleatória (prefixo + números)
func GeneratePassword(length int, prefix string) string {
	charset := "0123456789" // 🔥 Apenas números

	result := make([]byte, length)
	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[num.Int64()]
	}

	// 🔥 Retorna senha formatada corretamente
	return prefix + string(result)
}

// Função para formatar `exp_date` corretamente (timestamp → dd/mm/aaaa hh:mm)
func FormatTimestamp(timestampStr interface{}) string {
	// Converter `interface{}` para `string`
	timestampStrVal, ok := timestampStr.(string)
	if !ok {
		return "" // Retorna vazio se não for string
	}

	// Converter string para int64 (timestamp UNIX)
	timestamp, err := strconv.ParseInt(timestampStrVal, 10, 64)
	if err != nil {
		return "" // Retorna vazio se a conversão falhar
	}

	// Converter timestamp para `time.Time`
	t := time.Unix(timestamp, 0)

	// Retornar data formatada no padrão `dd/mm/aaaa hh:mm`
	return t.Format("02/01/2006 15:04")
}

// Funções utilitárias para ler variáveis de configuração do .env

func GetConfiancaDiasMin() int {
	val, _ := strconv.Atoi(os.Getenv("CONFIANCA_DIAS_MIN"))
	return val
}

func GetConfiancaDiasMax() int {
	val, _ := strconv.Atoi(os.Getenv("CONFIANCA_DIAS_MAX"))
	return val
}

func GetConfiancaFrequenciaDias() int {
	val, _ := strconv.Atoi(os.Getenv("CONFIANCA_FREQUENCIA_DIAS"))
	return val
}

func GetRollbackPermitidoDias() int {
	val, _ := strconv.Atoi(os.Getenv("ROLLBACK_PERMITIDO_DIAS"))
	return val
}

func GetAlteracaoVencimentoFrequenciaDias() int {
	val, _ := strconv.Atoi(os.Getenv("ALTERACAO_VENCIMENTO_FREQUENCIA_DIAS"))
	return val
}

func GetRollbackPermitidoFrequencia() int {
	val, _ := strconv.Atoi(os.Getenv("ROLLBACK_PERMITIDO_FREQUENCIA"))
	return val
}

// Salva log de ação em uma collection MongoDB "actions_log"
func SaveActionLog(userID int, action string, details interface{}, adminID string) error {
	if config.MongoDB == nil {
		return fmt.Errorf("MongoDB não inicializado")
	}
	collection := config.MongoDB.Database("Logs").Collection("actions_log")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logEntry := bson.M{
		"user_id":   userID,
		"action":    action,
		"details":   details,
		"admin_id":  adminID,
		"timestamp": time.Now(),
	}

	_, err := collection.InsertOne(ctx, logEntry)
	return err
}

// TokenInfo contém dados extraídos do token JWT
type TokenInfo struct {
	MemberID int
	Username string
}

// ValidateAndExtractToken faz a validação do token e retorna informações essenciais
func ValidateAndExtractToken(c *gin.Context) (*TokenInfo, bool) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(401, gin.H{"erro": "Token não fornecido"})
		return nil, false
	}
	claims, _, err := ValidateToken(tokenString)
	if err != nil {
		c.JSON(401, gin.H{"erro": "Token inválido ou expirado"})
		return nil, false
	}
	memberIDFloat, ok := claims["member_id"].(float64)
	if !ok {
		c.JSON(401, gin.H{"erro": "MemberID não encontrado no token"})
		return nil, false
	}
	username, ok := claims["username"].(string)
	if !ok {
		c.JSON(401, gin.H{"erro": "Username não encontrado no token"})
		return nil, false
	}
	return &TokenInfo{
		MemberID: int(memberIDFloat),
		Username: username,
	}, true
}

// Salva um valor em formato JSON no Redis com um tempo de expiração em segundos
func SaveToRedisJSON(ctx context.Context, key string, value interface{}, ttlSeconds int) error {
	log.Printf("Redis: Tentando salvar chave: %s, TTL: %d segundos", key, ttlSeconds)
	data, err := json.Marshal(value)
	if err != nil {
		log.Printf("Redis: Erro ao serializar JSON para chave %s: %v", key, err)
		return err
	}

	if config.RedisClient == nil {
		log.Printf("Redis: Cliente Redis não inicializado para chave %s", key)
		return fmt.Errorf("cliente Redis não inicializado")
	}

	statusCmd := config.RedisClient.Set(ctx, key, data, time.Duration(ttlSeconds)*time.Second)
	if err := statusCmd.Err(); err != nil {
		log.Printf("Redis: Erro ao executar SET para chave %s no Redis: %v", key, err)
		return err
	}
	log.Printf("Redis: Chave %s salva com sucesso no Redis. Resultado: %s", key, statusCmd.Val())
	return nil
}
