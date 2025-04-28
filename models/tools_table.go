package models

// EditUserResponse representa a resposta após edição de usuário
type EditUserResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
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
