package models

import (
	"database/sql"
)

// EditUserRequest representa os dados que podem ser editados no usuário
type EditUserRequest struct {
	Username             string         `json:"username" binding:"required"`
	Password             string         `json:"password,omitempty"`
	ResellerNotes        string         `json:"reseller_notes,omitempty"`
	NumeroWhats          string         `json:"numero_whats,omitempty"`
	NomeParaAviso        string         `json:"nome_para_aviso,omitempty"`
	EnviarNotificacao    *bool          `json:"enviar_notificacao,omitempty"`
	Bouquet              string         `json:"bouquet,omitempty"`
	Aplicativo           sql.NullString `json:"aplicativo,omitempty"`
	DeviceID             int64          `json:"device_id,omitempty"`
	MAC                  string         `json:"mac,omitempty"`
	NomeDoAplicativo     string         `json:"nome_do_aplicativo,omitempty"`
	VencimentoAplicativo string         `json:"vencimento_aplicativo,omitempty"`
}

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
