package models

// Estrutura para cada aplicativo
type AplicativoInfo struct {
	NomeDoAplicativo     string `json:"nome_do_aplicativo"`
	MAC                  string `json:"mac"`
	DeviceID             string `json:"device_id"`
	VencimentoAplicativo string `json:"vencimento_aplicativo"`
}

// EditUserRequest atualizado para aceitar array de aplicativos
type EditUserRequest struct {
	Username          string           `json:"username,omitempty"`
	Password          string           `json:"password,omitempty"`
	ResellerNotes     string           `json:"reseller_notes,omitempty"`
	NumeroWhats       *string          `json:"numero_whats,omitempty"`
	NomeParaAviso     *string          `json:"nome_para_aviso,omitempty"`
	EnviarNotificacao *bool            `json:"enviar_notificacao,omitempty"`
	Bouquet           string           `json:"bouquet,omitempty"`
	Aplicativos       []AplicativoInfo `json:"aplicativos,omitempty"`
	Notificacao_conta *bool            `json:"Notificacao_conta,omitempty"`
	Notificacao_vods  *bool            `json:"Notificacao_vods,omitempty"`
	Notificacao_jogos *bool            `json:"Notificacao_jogos,omitempty"`
	FranquiaMemberID  *int             `json:"franquia_member_id,omitempty"`
}
