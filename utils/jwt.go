package utils

import (
	"apiBackEnd/config"
	"context"
	"errors" // 🔹 Importação para logs
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("Prs5bR2t%vWT>m+?syisEh0f2h+?/CDz=sA[:Y9CSWAjdZv&oF1x8g*TT_76<QSI")

// GenerateToken agora armazena o token com TTL exato no Redis
func GenerateToken(username string, memberID int, credits float64, status string) (string, error) {
	ctx := context.Background()
	redisKey := "token:" + strconv.Itoa(memberID)

	////log.Printf("Gerando token para username: %s, memberID: %d, credits: %.2f, status: %s", username, memberID, credits, status)

	// 🔹 **Verifica se já existe um token ativo no Redis**
	existingToken, err := config.RedisClient.Get(ctx, redisKey).Result()
	if err == nil && existingToken != "" {
		////log.Printf("Token existente encontrado no Redis para memberID: %d", memberID)
		return existingToken, nil
	}

	// 🔹 **Obtém tempo de expiração do .env**
	expirationMinutes, err := strconv.Atoi(os.Getenv("TOKEN_EXPIRATION_MINUTES"))
	if err != nil || expirationMinutes <= 0 {
		////log.Printf("Erro ao obter TOKEN_EXPIRATION_MINUTES, usando fallback de 60 minutos. Erro: %v", err)
		expirationMinutes = 60
	}
	expirationDuration := time.Duration(expirationMinutes) * time.Minute
	expirationTime := time.Now().Add(expirationDuration).Unix()

	////log.Printf("Tempo de expiração do token: %d minutos (%d segundos)", expirationMinutes, expirationTime)

	// 🔥 Claims com campos 100% visíveis no JWT.io
	claims := jwt.MapClaims{
		"username":  username,
		"member_id": memberID,
		"exp":       expirationTime,
		"credits":   credits, // 🔹 Certifique-se de que o valor correto está sendo atribuído
		"status":    status,  // 🔹 Certifique-se de que o valor correto está sendo atribuído
	}

	////log.Printf("Claims geradas: %+v", claims)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		////log.Printf("Erro ao assinar o token: %v", err)
		return "", err
	}

	////log.Printf("Token gerado com sucesso: %s", tokenString)

	// 🔥 **Armazena o token no Redis com TTL baseado na expiração**
	err = config.RedisClient.Set(ctx, redisKey, tokenString, expirationDuration).Err()
	if err != nil {
		////log.Printf("Erro ao armazenar o token no Redis: %v", err)
		return "", err
	}

	////log.Printf("Token armazenado no Redis com TTL de %d minutos", expirationMinutes)

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

	//log.Printf("Validando token: %s", tokenString)

	// 🔥 **Remover "Bearer " do token, se existir**
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	//log.Printf("Token após remoção do prefixo 'Bearer ': %s", tokenString)

	// 🔹 **Decodificar o token para obter `member_id`**
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		//log.Printf("Erro ao decodificar o token: %v", err)
		return nil, 0, errors.New("token inválido")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		//log.Printf("Erro ao converter claims do token")
		return nil, 0, errors.New("token inválido")
	}

	//log.Printf("Claims decodificadas: %+v", claims)

	memberID, ok := claims["member_id"].(float64)
	if !ok {
		//log.Printf("Erro: member_id não encontrado nas claims")
		return nil, 0, errors.New("member_id não encontrado no token")
	}

	redisKey := "token:" + strconv.Itoa(int(memberID))
	//log.Printf("Chave do Redis para validação: %s", redisKey)

	// 🔥 **Verifica no Redis se o token ainda é válido**
	storedToken, err := config.RedisClient.Get(ctx, redisKey).Result()
	if err != nil || storedToken != tokenString {
		//log.Printf("Token não encontrado ou não corresponde ao armazenado no Redis")
		return nil, 0, errors.New("token expirado ou não autorizado")
	}

	//log.Printf("Token encontrado no Redis e válido")

	// 🔹 **Calcular tempo restante**
	exp, ok := claims["exp"].(float64)
	if !ok {
		//log.Printf("Erro: expiração do token não encontrada nas claims")
		return nil, 0, errors.New("expiração do token não encontrada")
	}

	now := time.Now().Unix()
	timeRemaining := int64(exp) - now
	//log.Printf("Tempo restante do token: %d segundos", timeRemaining)

	if timeRemaining <= 0 {
		//log.Printf("Token expirado")
		return nil, 0, errors.New("token expirado")
	}

	return claims, timeRemaining, nil
}

// RenewSubscription realiza a renovação da assinatura com validação e débito de créditos
func RenewSubscription(memberID int, renewalCost float64) error {
	ctx := context.Background()
	redisKey := "credits:" + strconv.Itoa(memberID)

	// 🔹 **Obter créditos atuais do Redis**
	creditsStr, err := config.RedisClient.Get(ctx, redisKey).Result()
	if err != nil {
		//log.Printf("Erro ao obter créditos para memberID %d: %v", memberID, err)
		return errors.New("não foi possível obter os créditos do usuário")
	}

	credits, err := strconv.ParseFloat(creditsStr, 64)
	if err != nil {
		//log.Printf("Erro ao converter créditos para float: %v", err)
		return errors.New("créditos inválidos")
	}

	//log.Printf("Créditos atuais para memberID %d: %.2f", memberID, credits)

	// 🔹 **Verificar se há créditos suficientes**
	if credits < renewalCost {
		//log.Printf("Créditos insuficientes para memberID %d. Necessário: %.2f, Disponível: %.2f", memberID, renewalCost, credits)
		return errors.New("créditos insuficientes para renovação")
	}

	// 🔹 **Debitar créditos**
	newCredits := credits - renewalCost
	err = config.RedisClient.Set(ctx, redisKey, strconv.FormatFloat(newCredits, 'f', 2, 64), 0).Err()
	if err != nil {
		//log.Printf("Erro ao debitar créditos para memberID %d: %v", memberID, err)
		return errors.New("não foi possível debitar os créditos")
	}

	//log.Printf("Créditos debitados com sucesso. Créditos restantes para memberID %d: %.2f", memberID, newCredits)

	// 🔹 **Renovar assinatura**
	// Aqui você pode adicionar a lógica para renovar a assinatura, como atualizar o status no banco de dados ou outro sistema.

	//log.Printf("Assinatura renovada com sucesso para memberID %d", memberID)
	return nil
}
