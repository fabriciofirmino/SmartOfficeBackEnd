package models

type EditUserRequest struct {
	Username             string  `json:"username"`
	Password             string  `json:"password"`
	ResellerNotes        string  `json:"reseller_notes"`
	NumeroWhats          *string `json:"numero_whats"`
	NomeParaAviso        *string `json:"nome_para_aviso"`
	EnviarNotificacao    *bool   `json:"enviar_notificacao"`
	Bouquet              string  `json:"bouquet"`
	NomeDoAplicativo     string  `json:"nome_do_aplicativo"`
	MAC                  string  `json:"mac"`
	DeviceID             string  `json:"device_id"`
	VencimentoAplicativo string  `json:"vencimento_aplicativo"`
	Aplicativo           string  `json:"aplicativo"`
}
