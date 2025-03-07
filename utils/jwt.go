package utils

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("chave-secreta-super-segura")

// Função para gerar um token JWT
func GenerateToken(username string, memberID int) (string, error) {
	claims := jwt.MapClaims{
		"username":  username,
		"member_id": memberID, // Agora inclui o member_id correto
		"exp":       time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// Função para validar o token JWT sem quebrar GenerateToken
func ValidateToken(tokenString string) (jwt.MapClaims, error) {
	// Remover "Bearer " do início do token, se existir
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// Parse do token JWT
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verificar se o método de assinatura é válido
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de assinatura inválido")
		}
		return secretKey, nil
	})

	// Se o token não for válido, retorna erro
	if err != nil {
		return nil, err
	}

	// Extrair claims corretamente
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("token inválido")
	}

	return claims, nil
}
