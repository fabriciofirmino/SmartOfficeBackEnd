package models

import "database/sql"

// ClientTableData representa os dados de um cliente na tabela
type ClientTableData struct {
	ID             int            `json:"id"`
	Username       string         `json:"username"`
	Password       string         `json:"password"`
	ExpDate        sql.NullString `json:"exp_date"` // ✅ Agora aceita NULL
	Enabled        bool           `json:"enabled"`
	AdminEnabled   bool           `json:"admin_enabled"`
	MaxConnections int            `json:"max_connections"`
	CreatedAt      string         `json:"created_at"`
	ResellerNotes  sql.NullString `json:"reseller_notes"` // ✅ Também pode ser NULL
	IsTrial        bool           `json:"is_trial"`
}
