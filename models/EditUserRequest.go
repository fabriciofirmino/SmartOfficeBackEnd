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
	Username          string           `json:"username,omitempty" validate:"omitempty"`
	Password          string           `json:"password,omitempty" validate:"omitempty"`
	ResellerNotes     string           `json:"reseller_notes,omitempty"`
	NumeroWhats       *string          `json:"numero_whats,omitempty"`       // Ponteiro para string para aceitar null ou string vazia
	NomeParaAviso     *string          `json:"nome_para_aviso,omitempty"`    // Ponteiro para string
	EnviarNotificacao *bool            `json:"enviar_notificacao,omitempty"` // Ponteiro para bool
	Bouquet           string           `json:"bouquet,omitempty"`            // String JSON representando um array de IDs
	Aplicativos       []AplicativoInfo `json:"aplicativos,omitempty"`        // Slice de AplicativoInfo
	Notificacao_conta *bool            `json:"Notificacao_conta,omitempty"`  // Preferência de notificação para conta
	Notificacao_vods  *bool            `json:"Notificacao_vods,omitempty"`   // Preferência de notificação para VODs
	Notificacao_jogos *bool            `json:"Notificacao_jogos,omitempty"`  // Preferência de notificação para jogos
	FranquiaMemberID  *int             `json:"franquia_member_id,omitempty"` // ID da franquia (opcional)
	Valor_plano       *float64         `json:"Valor_plano,omitempty"`        // Novo campo para o valor do plano
}
