package models

// ClientTableData estrutura os dados limitados retornados na rota clients-table
type ClientTableData struct {
	ID             int    `json:"id"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	ExpDate        string `json:"exp_date"`
	Enabled        int    `json:"enabled"`
	AdminEnabled   int    `json:"admin_enabled"`
	MaxConnections int    `json:"max_connections"`
	CreatedAt      string `json:"created_at"`
	ResellerNotes  string `json:"reseller_notes"`
	IsTrial        int    `json:"is_trial"`
}
