package utils

import (
	"apiBackEnd/config"
	"database/sql"
)

// VerificaPermissaoUsuario checa se a revenda tem permissão para modificar o usuário
// Retorna: permitido (bool), responsibleMemberID (int), erro (error)
func VerificaPermissaoUsuario(userID, revendaResponsavel int) (bool, int, error) {
	// Verificação de segurança: apenas o membro responsável ou super admin pode alterar o usuário
	var responsibleMemberID int
	checkQuery := "SELECT member_id FROM streamcreed_db.users WHERE id = ?"
	err := config.DB.QueryRow(checkQuery, userID).Scan(&responsibleMemberID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, 0, err
		}
		return false, 0, err
	}

	// Super admin (ID 1) tem acesso a tudo, ou se for a própria revenda responsável
	isAdmin := revendaResponsavel == 1
	if isAdmin || revendaResponsavel == responsibleMemberID {
		return true, responsibleMemberID, nil
	}

	return false, responsibleMemberID, nil
}
