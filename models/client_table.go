package models

import (
	"database/sql"
	"encoding/json"
)

// ClientTableData representa os dados de um cliente na tabela
type ClientTableData struct {
	ID             int            `json:"id"`
	Username       string         `json:"username"`
	Password       string         `json:"password"`
	ExpDate        sql.NullString `json:"exp_date"`
	Enabled        bool           `json:"enabled"`
	AdminEnabled   bool           `json:"admin_enabled"`
	MaxConnections int            `json:"max_connections"`
	CreatedAt      string         `json:"created_at"`
	ResellerNotes  sql.NullString `json:"reseller_notes"`
	IsTrial        bool           `json:"is_trial"`
}

// MarshalJSON transforma NullString em string normal no JSON
func (c ClientTableData) MarshalJSON() ([]byte, error) {
	type Alias ClientTableData
	return json.Marshal(&struct {
		ExpDate       string `json:"exp_date"`
		ResellerNotes string `json:"reseller_notes"`
		Alias
	}{
		ExpDate:       nullStringToString(c.ExpDate),
		ResellerNotes: nullStringToString(c.ResellerNotes),
		Alias:         (Alias)(c),
	})
}

// Função auxiliar para converter sql.NullString → string normal
func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}
