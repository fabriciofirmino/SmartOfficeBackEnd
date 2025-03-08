package utils

import (
	"apiBackEnd/config"
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("Prs5bR2t%vWT>m+?syisEh0f2h+?/CDz=sA[:Y9CSWAjdZv&oF1x8g*TT_76<QSI")

// GenerateToken agora armazena o token com TTL exato no Redis
func GenerateToken(username string, memberID int) (string, error) {
	ctx := context.Background()
	redisKey := "token:" + strconv.Itoa(memberID)

	// 🔹 **Verifica se já existe um token ativo no Redis**
	existingToken, err := config.RedisClient.Get(ctx, redisKey).Result()
	if err == nil && existingToken != "" {
		// 🔥 **Retorna o token existente sem gerar um novo**
		return existingToken, nil
	}

	// 🔹 **Obtém tempo de expiração do .env**
	expirationMinutes, err := strconv.Atoi(os.Getenv("TOKEN_EXPIRATION_MINUTES"))
	if err != nil || expirationMinutes <= 0 {
		expirationMinutes = 60 // 🔥 **Fallback para 60 minutos**
	}
	expirationDuration := time.Duration(expirationMinutes) * time.Minute
	expirationTime := time.Now().Add(expirationDuration).Unix()

	// 🔥 **Criação do token com expiração correta**
	claims := jwt.MapClaims{
		"username":  username,
		"member_id": memberID,
		"exp":       expirationTime,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	// 🔥 **Armazena o token no Redis com TTL baseado na expiração**
	err = config.RedisClient.Set(ctx, redisKey, tokenString, expirationDuration).Err()
	if err != nil {
		return "", err
	}

	// 🔥 **Define manualmente o TTL do Redis (caso necessário)**
	err = config.RedisClient.Expire(ctx, redisKey, expirationDuration).Err()
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken agora verifica TTL no Redis
func ValidateToken(tokenString string) (jwt.MapClaims, int64, error) {
	ctx := context.Background()

	// 🔥 **Remover "Bearer " do token, se existir**
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// 🔹 **Decodificar o token para obter `member_id`**
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, 0, errors.New("token inválido")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, 0, errors.New("token inválido")
	}

	memberID, ok := claims["member_id"].(float64)
	if !ok {
		return nil, 0, errors.New("member_id não encontrado no token")
	}

	redisKey := "token:" + strconv.Itoa(int(memberID))

	// 🔥 **Verifica no Redis se o token ainda é válido**
	storedToken, err := config.RedisClient.Get(ctx, redisKey).Result()
	if err != nil || storedToken != tokenString {
		return nil, 0, errors.New("token expirado ou não autorizado")
	}

	// 🔥 **Verifica o tempo restante do token no Redis**
	ttl, err := config.RedisClient.TTL(ctx, redisKey).Result()
	if err != nil || ttl <= 0 {
		return nil, 0, errors.New("token expirado")
	}

	// 🔹 **Calcular tempo restante**
	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, 0, errors.New("expiração do token não encontrada")
	}

	now := time.Now().Unix()
	timeRemaining := int64(exp) - now
	if timeRemaining <= 0 {
		return nil, 0, errors.New("token expirado")
	}

	return claims, timeRemaining, nil
}
