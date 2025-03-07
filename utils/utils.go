package utils

import (
	"crypto/rand"
	"math/big"
	"strconv"
	"time"
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
