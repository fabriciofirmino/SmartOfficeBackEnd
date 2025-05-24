package models

import (
	"apiBackEnd/config"
	"database/sql"
	"fmt"
	"os"
	"strings"
)

// User representa a estrutura do usuário
type User struct {
	Username      string
	PasswordHash  string
	MemberGroupID int
	Credits       float64
	Status        int
	MemberID      int
	// Adicione outros campos que você possa ter
}

// GetUserByUsername busca um usuário pelo nome, considerando ALLOWED_USERS.
func GetUserByUsername(username string) (*User, error) {
	if config.DB == nil {
		return nil, fmt.Errorf("conexão com banco de dados não inicializada")
	}

	allowedUsersEnv := os.Getenv("ALLOWED_USERS")
	if allowedUsersEnv != "" {
		isAllowed := false
		allowedUsersList := strings.Split(allowedUsersEnv, ",")
		for _, allowedUser := range allowedUsersList {
			if strings.TrimSpace(allowedUser) == username {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			// Usuário não está na lista de permitidos, retorna erro similar a "não encontrado"
			// para não vazar informação de que o usuário existe mas não está permitido.
			return nil, sql.ErrNoRows
		}
	}

	// Se chegou aqui, ou ALLOWED_USERS está vazio, ou o username está na lista.
	// Prossegue com a busca normal.
	query := `
		SELECT username, password, member_group_id, credits, status, id as member_id
		FROM streamcreed_db.reg_users
		WHERE username = ?
	`
	// Nota: No seu exemplo anterior, a tabela era streamcreed_db.reg_users AS ru.
	// Se você usa alias, certifique-se de que os nomes das colunas estão corretos (ex: ru.member_group_id).
	// Para simplificar, removi o alias 'ru' aqui, ajuste se necessário.

	user := &User{}
	err := config.DB.QueryRow(query, username).Scan(
		&user.Username,
		&user.PasswordHash,
		&user.MemberGroupID,
		&user.Credits,
		&user.Status,
		&user.MemberID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows // Usuário não encontrado
		}
		// Logar o erro real para depuração interna
		// log.Printf("Erro ao buscar usuário %s: %v", username, err)
		return nil, fmt.Errorf("erro ao buscar usuário: %w", err)
	}

	return user, nil
}

// Outras funções do model user.go ...
