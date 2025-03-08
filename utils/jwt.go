package utils

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("Prs5bR2t%vWT>m+?syisEh0f2h+?/CDz=sA[:Y9CSWAjdZv&oF1x8g*TT_76<QSI")

// GenerateToken gera um token JWT com tempo de expiração definido no .env
func GenerateToken(username string, memberID int) (string, error) {
	expirationMinutes, err := strconv.Atoi(os.Getenv("TOKEN_EXPIRATION_MINUTES"))
	if err != nil || expirationMinutes <= 0 {
		expirationMinutes = 1440 // 🔥 Padrão: 24 horas (caso não tenha no .env)
	}

	claims := jwt.MapClaims{
		"username":  username,
		"member_id": memberID,
		"exp":       time.Now().Add(time.Duration(expirationMinutes) * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// ValidateToken valida o JWT e retorna os claims + tempo restante
func ValidateToken(tokenString string) (jwt.MapClaims, int, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de assinatura inválido")
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, 0, errors.New("token inválido")
	}

	expTime, ok := claims["exp"].(float64)
	if !ok {
		return nil, 0, errors.New("token sem data de expiração")
	}

	expirationTime := time.Unix(int64(expTime), 0)
	timeRemaining := int(time.Until(expirationTime).Minutes())

	if timeRemaining <= 0 {
		return nil, 0, errors.New("token expirado")
	}

	return claims, timeRemaining, nil
}
