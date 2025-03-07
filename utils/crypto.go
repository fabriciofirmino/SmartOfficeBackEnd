package utils

import (
	"github.com/tredoe/osutil/user/crypt/sha512_crypt"
)

// Função para gerar o hash da senha com SHA-512 e salt
func CryptPassword(password string) (string, error) {
	salt := "$6$rounds=20000$xtreamcodes$" // Salt fixo usado na criptografia

	crypt := sha512_crypt.New()
	hashedPassword, err := crypt.Generate([]byte(password), []byte(salt))
	if err != nil {
		return "", err
	}

	return hashedPassword, nil
}
