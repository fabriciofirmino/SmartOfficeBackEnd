package models

// Estrutura para edição de usuário
type EditUserRequest struct {
	ID                int                    `json:"id"`
	Username          string                 `json:"username"`
	Password          string                 `json:"password"`
	ResellerNotes     string                 `json:"reseller_notes"`
	NumeroWhats       string                 `json:"NUMERO_WHATS"`
	NomeParaAviso     string                 `json:"NOME_PARA_AVISO"`
	EnviarNotificacao bool                   `json:"ENVIAR_NOTIFICACAO"`
	Bouquet           string                 `json:"bouquet"`
	AppData           map[string]interface{} `json:"Aplicativo"`
}

// Estrutura para adicionar/remover tela
type ScreenRequest struct {
	UserID int `json:"userID" binding:"required"`
}

// Estrutura de resposta ao adicionar/remover telas
type ScreenResponse struct {
	TotalTelas     int     `json:"total_telas"`
	ValorCobrado   float64 `json:"valor_cobrado"`
	CreditosAntes  float64 `json:"creditos_antes"`
	CreditosAtuais float64 `json:"creditos_atuais"`
}
