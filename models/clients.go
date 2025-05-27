package models

import (
	"apiBackEnd/config"
	"fmt"
	"log"
	"strings"
)

// Estrutura para armazenar os clientes retornados
// ClientData representa os dados do cliente para operações de banco de dados
// swagger:model ClientData
type ClientData struct {
	ID                int        `json:"id" example:"123"`
	MemberID          int        `json:"member_id" example:"456"`
	Username          string     `json:"username" example:"john_doe"`
	Password          string     `json:"password" example:"hashed_password"`
	ExpDate           NullString `json:"exp_date" swaggertype:"string" example:"2025-12-31T23:59:59Z"`
	AdminEnabled      NullInt64  `json:"admin_enabled" swaggertype:"integer" example:"1"`
	Enabled           NullInt64  `json:"enabled" swaggertype:"integer" example:"1"`
	AdminNotes        NullString `json:"admin_notes" swaggertype:"string" example:"Notas administrativas"`
	ResellerNotes     NullString `json:"reseller_notes" swaggertype:"string" example:"Notas do revendedor"`
	Bouquet           NullString `json:"bouquet" swaggertype:"string" example:"Pacote Premium"`
	MaxConnections    NullInt64  `json:"max_connections" swaggertype:"integer" example:"5"`
	IsRestreamer      NullInt64  `json:"is_restreamer" swaggertype:"integer" example:"0"`
	AllowedIPs        NullString `json:"allowed_ips" swaggertype:"string" example:"192.168.1.1, 10.0.0.1"`
	AllowedUA         NullString `json:"allowed_ua" swaggertype:"string" example:"Mozilla/5.0"`
	IsTrial           NullInt64  `json:"is_trial" swaggertype:"integer" example:"1"`
	CreatedAt         NullString `json:"created_at" swaggertype:"string" example:"2024-03-15T14:30:00Z"`
	CreatedBy         NullString `json:"created_by" swaggertype:"string" example:"admin"`
	PairID            NullInt64  `json:"pair_id" swaggertype:"integer" example:"789"`
	IsMag             NullInt64  `json:"is_mag" swaggertype:"integer" example:"0"`
	IsE2              NullInt64  `json:"is_e2" swaggertype:"integer" example:"1"`
	ForceServerID     NullInt64  `json:"force_server_id" swaggertype:"integer" example:"3"`
	IsIspLock         NullInt64  `json:"is_isplock" swaggertype:"integer" example:"0"`
	IspDesc           NullString `json:"isp_desc" swaggertype:"string" example:"Provedor XYZ"`
	ForcedCountry     NullString `json:"forced_country" swaggertype:"string" example:"BR"`
	IsStalker         NullInt64  `json:"is_stalker" swaggertype:"integer" example:"0"`
	BypassUA          NullString `json:"bypass_ua" swaggertype:"string" example:"CustomUserAgent"`
	AsNumber          NullString `json:"as_number" swaggertype:"string" example:"AS12345"`
	PlayToken         NullString `json:"play_token" swaggertype:"string" example:"xyz123token"`
	PackageID         NullInt64  `json:"package_id" swaggertype:"integer" example:"11"`
	UsrMac            NullString `json:"usr_mac" swaggertype:"string" example:"00:1A:2B:3C:4D:5E"`
	UsrDeviceKey      NullString `json:"usr_device_key" swaggertype:"string" example:"device-key-123"`
	Notes2            NullString `json:"notes2" swaggertype:"string" example:"Notas adicionais"`
	RootEnabled       NullInt64  `json:"root_enabled" swaggertype:"integer" example:"1"`
	NumeroWhats       NullString `json:"numero_whats" swaggertype:"string" example:"+5511999999999"`
	NomeParaAviso     NullString `json:"nome_para_aviso" swaggertype:"string" example:"João Silva"`
	Email             NullString `json:"email" swaggertype:"string" example:"joao@example.com"`
	EnviarNotificacao NullString `json:"enviar_notificacao" swaggertype:"string" example:"true"`
	SobrenomeAvisos   NullString `json:"sobrenome_avisos" swaggertype:"string" example:"Silva"`
	Deleted           NullInt64  `json:"deleted" swaggertype:"integer" example:"0"`
	DateDeleted       NullString `json:"date_deleted" swaggertype:"string" example:"2024-01-01T00:00:00Z"`
	AppID             NullString `json:"app_id" swaggertype:"string" example:"app-12345"`
	TrustRenew        NullInt64  `json:"trust_renew" swaggertype:"integer" example:"1"`
	Franquia          NullString `json:"franquia" swaggertype:"string" example:"Franquia ABC"`
	FranquiaMemberID  NullInt64  `json:"franquia_member_id" swaggertype:"integer" example:"999"`
	P2P               NullInt64  `json:"p2p" swaggertype:"integer" example:"0"`
	// Adiciona o campo Aplicativo
	Aplicativo NullString `json:"aplicativo" swaggertype:"string" example:"AppX"`
}

// swagger:model
// swagger:model ClientDataSwagger
type ClientDataSwagger struct {
	ID                int    `json:"id" example:"123"`
	MemberID          int    `json:"member_id" example:"456"`
	Username          string `json:"username" example:"john_doe"`
	Password          string `json:"password" example:"hashed_password"`
	ExpDate           string `json:"exp_date" example:"2025-12-31T23:59:59Z"` // Usando string para datas
	AdminEnabled      int    `json:"admin_enabled" example:"1"`
	Enabled           int    `json:"enabled" example:"1"`
	AdminNotes        string `json:"admin_notes" example:"Notas administrativas"`
	ResellerNotes     string `json:"reseller_notes" example:"Notas do revendedor"`
	Bouquet           string `json:"bouquet" example:"Pacote Premium"`
	MaxConnections    int    `json:"max_connections" example:"5"`
	IsRestreamer      int    `json:"is_restreamer" example:"0"`
	AllowedIPs        string `json:"allowed_ips" example:"192.168.1.1, 10.0.0.1"`
	AllowedUA         string `json:"allowed_ua" example:"Mozilla/5.0"`
	IsTrial           int    `json:"is_trial" example:"1"`
	CreatedAt         string `json:"created_at" example:"2024-03-15T14:30:00Z"`
	CreatedBy         string `json:"created_by" example:"admin"`
	PairID            int    `json:"pair_id" example:"789"`
	IsMag             int    `json:"is_mag" example:"0"`
	IsE2              int    `json:"is_e2" example:"1"`
	ForceServerID     int    `json:"force_server_id" example:"3"`
	IsIspLock         int    `json:"is_isplock" example:"0"`
	IspDesc           string `json:"isp_desc" example:"Provedor XYZ"`
	ForcedCountry     string `json:"forced_country" example:"BR"`
	IsStalker         int    `json:"is_stalker" example:"0"`
	BypassUA          string `json:"bypass_ua" example:"CustomUserAgent"`
	AsNumber          string `json:"as_number" example:"AS12345"`
	PlayToken         string `json:"play_token" example:"xyz123token"`
	PackageID         int    `json:"package_id" example:"11"`
	UsrMac            string `json:"usr_mac" example:"00:1A:2B:3C:4D:5E"`
	UsrDeviceKey      string `json:"usr_device_key" example:"device-key-123"`
	Notes2            string `json:"notes2" example:"Notas adicionais"`
	RootEnabled       int    `json:"root_enabled" example:"1"`
	NumeroWhats       string `json:"numero_whats" example:"+5511999999999"`
	NomeParaAviso     string `json:"nome_para_aviso" example:"João Silva"`
	Email             string `json:"email" example:"joao@example.com"`
	EnviarNotificacao bool   `json:"enviar_notificacao" example:"true"`
	SobrenomeAvisos   string `json:"sobrenome_avisos" example:"Silva"`
	Deleted           int    `json:"deleted" example:"0"`
	DateDeleted       string `json:"date_deleted" example:"2024-01-01T00:00:00Z"`
	AppID             string `json:"app_id" example:"app-12345"`
	TrustRenew        int    `json:"trust_renew" example:"1"`
	Franquia          string `json:"franquia" example:"Franquia ABC"`
	FranquiaMemberID  int    `json:"franquia_member_id" example:"999"`
	P2P               int    `json:"p2p" example:"0"`
}

// Buscar clientes aplicando filtros opcionais
func GetClientsByFilters(memberID int, filters map[string]interface{}) ([]ClientData, error) {
	var clients []ClientData
	var conditions []string
	var args []interface{}

	// Filtro obrigatório (sempre verifica o member_id do usuário autenticado)
	conditions = append(conditions, "member_id = ?")
	args = append(args, memberID)

	// Aplicando filtros opcionais
	if id, ok := filters["id"]; ok {
		conditions = append(conditions, "id = ?")
		args = append(args, id)
	}
	if username, ok := filters["username"]; ok {
		conditions = append(conditions, "username = ?")
		args = append(args, username)
	}
	if numeroWhats, ok := filters["numero_whats"]; ok {
		conditions = append(conditions, "numero_whats = ?")
		args = append(args, numeroWhats)
	}
	if enviarNotificacao, ok := filters["enviar_notificacao"]; ok {
		conditions = append(conditions, "enviar_notificacao = ?")
		args = append(args, enviarNotificacao)
	}
	if maxConnections, ok := filters["max_connections"]; ok {
		conditions = append(conditions, "max_connections = ?")
		args = append(args, maxConnections)
	}
	if isTrial, ok := filters["is_trial"]; ok {
		conditions = append(conditions, "is_trial = ?")
		args = append(args, isTrial)
	}
	if enabled, ok := filters["enabled"]; ok {
		conditions = append(conditions, "enabled = ?")
		args = append(args, enabled)
	}
	if adminNotes, ok := filters["admin_notes"]; ok {
		conditions = append(conditions, "admin_notes LIKE ?")
		args = append(args, "%"+adminNotes.(string)+"%")
	}
	if email, ok := filters["email"]; ok {
		conditions = append(conditions, "email = ?")
		args = append(args, email)
	}
	if expDate, ok := filters["exp_date"]; ok {
		conditions = append(conditions, "exp_date = FROM_UNIXTIME(?)") // Converte Epoch para data no MySQL/MariaDB
		args = append(args, expDate)
	}

	// Construção da query final usando strings.Join (CORRIGIDO)
	query := fmt.Sprintf(`
		SELECT id, member_id, username, password, exp_date, admin_enabled, enabled, admin_notes,
		reseller_notes, bouquet, max_connections, is_restreamer, allowed_ips, allowed_ua, is_trial,
		created_at, created_by, pair_id, is_mag, is_e2, force_server_id, is_isplock, isp_desc,
		forced_country, is_stalker, bypass_ua, as_number, play_token, package_id, USR_MAC,
		USR_DEVICE_KEY, notes2, root_enabled, NUMERO_WHATS, NOME_PARA_AVISO, EMAIL, ENVIAR_NOTIFICACAO,
		SOBRENOME_AVISOS, deleted, date_deleted, app_id, trust_renew, franquia, franquia_member_id, p2p,
		Aplicativo
		FROM users
		WHERE %s
	`, strings.Join(conditions, " AND ")) // 📌 Corrigido: inclui coluna Aplicativo

	log.Printf("Executando query: %s\n", query)
	log.Printf("Com argumentos: %v\n", args)

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		log.Printf("Erro ao executar query: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	// Processar os resultados
	for rows.Next() {
		var client ClientData
		if err := rows.Scan(
			&client.ID, &client.MemberID, &client.Username, &client.Password, &client.ExpDate, &client.AdminEnabled,
			&client.Enabled, &client.AdminNotes, &client.ResellerNotes, &client.Bouquet, &client.MaxConnections,
			&client.IsRestreamer, &client.AllowedIPs, &client.AllowedUA, &client.IsTrial, &client.CreatedAt,
			&client.CreatedBy, &client.PairID, &client.IsMag, &client.IsE2, &client.ForceServerID, &client.IsIspLock,
			&client.IspDesc, &client.ForcedCountry, &client.IsStalker, &client.BypassUA, &client.AsNumber,
			&client.PlayToken, &client.PackageID, &client.UsrMac, &client.UsrDeviceKey, &client.Notes2,
			&client.RootEnabled, &client.NumeroWhats, &client.NomeParaAviso, &client.Email,
			&client.EnviarNotificacao, &client.SobrenomeAvisos, &client.Deleted, &client.DateDeleted,
			&client.AppID, &client.TrustRenew, &client.Franquia, &client.FranquiaMemberID, &client.P2P,
			&client.Aplicativo, // Adiciona o campo Aplicativo ao Scan
		); err != nil {
			log.Printf("Erro ao escanear linha: %v\n", err)
			return nil, err
		}
		clients = append(clients, client)
	}

	log.Printf("Total de registros encontrados: %d\n", len(clients))
	return clients, nil
}
