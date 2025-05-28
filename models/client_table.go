package models

import (
	"database/sql"
	"encoding/json"
)

// ClientTableData representa os dados de um cliente na tabela
type ClientTableData struct {
	ID               int                    `json:"id"`
	Username         string                 `json:"username"`
	Password         string                 `json:"password"`
	ExpDate          sql.NullString         `json:"exp_date"` // ou string, conforme o banco
	Enabled          int                    `json:"enabled"`
	AdminEnabled     int                    `json:"admin_enabled"`
	MaxConnections   int                    `json:"max_connections"`
	CreatedAt        sql.NullString         `json:"created_at"`     // ou string
	ResellerNotes    sql.NullString         `json:"reseller_notes"` // ou string
	IsTrial          int                    `json:"is_trial"`
	Aplicativo       string                 `json:"aplicativo"`
	Online           map[string]interface{} `json:"online"`
	FranquiaMemberID sql.NullInt64          `json:"franquia_member_id"` // Novo campo
}

// MarshalJSON transforma NullString e NullInt64 em string/int normal no JSON
func (c ClientTableData) MarshalJSON() ([]byte, error) {
	type Alias ClientTableData
	return json.Marshal(&struct {
		ExpDate          string `json:"exp_date"`
		CreatedAt        string `json:"created_at"`
		ResellerNotes    string `json:"reseller_notes"`
		FranquiaMemberID *int64 `json:"franquia_member_id,omitempty"` // Alterado para ponteiro para omitempty funcionar corretamente
		Alias
	}{
		ExpDate:          nullStringToString(c.ExpDate),
		CreatedAt:        nullStringToString(c.CreatedAt),
		ResellerNotes:    nullStringToString(c.ResellerNotes),
		FranquiaMemberID: nullInt64ToInt64Ptr(c.FranquiaMemberID), // Nova função auxiliar
		Alias:            (Alias)(c),
	})
}

// Função auxiliar para converter sql.NullString → string normal
func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// Função auxiliar para converter sql.NullInt64 -> *int64
func nullInt64ToInt64Ptr(ni sql.NullInt64) *int64 {
	if ni.Valid {
		val := ni.Int64
		return &val
	}
	return nil
}
