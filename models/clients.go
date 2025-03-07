package models

import (
	"apiBackEnd/config"
	"database/sql"
	"fmt"
	"log"
	"strings"
)

// Estrutura para armazenar os clientes retornados
type ClientData struct {
	ID                int            `json:"id"`
	MemberID          int            `json:"member_id"`
	Username          string         `json:"username"`
	Password          string         `json:"password"`
	ExpDate           sql.NullString `json:"exp_date"`
	AdminEnabled      sql.NullInt64  `json:"admin_enabled"`
	Enabled           sql.NullInt64  `json:"enabled"`
	AdminNotes        sql.NullString `json:"admin_notes"`
	ResellerNotes     sql.NullString `json:"reseller_notes"`
	Bouquet           sql.NullString `json:"bouquet"`
	MaxConnections    sql.NullInt64  `json:"max_connections"`
	IsRestreamer      sql.NullInt64  `json:"is_restreamer"`
	AllowedIPs        sql.NullString `json:"allowed_ips"`
	AllowedUA         sql.NullString `json:"allowed_ua"`
	IsTrial           sql.NullInt64  `json:"is_trial"`
	CreatedAt         sql.NullString `json:"created_at"`
	CreatedBy         sql.NullString `json:"created_by"`
	PairID            sql.NullInt64  `json:"pair_id"`
	IsMag             sql.NullInt64  `json:"is_mag"`
	IsE2              sql.NullInt64  `json:"is_e2"`
	ForceServerID     sql.NullInt64  `json:"force_server_id"`
	IsIspLock         sql.NullInt64  `json:"is_isplock"`
	IspDesc           sql.NullString `json:"isp_desc"`
	ForcedCountry     sql.NullString `json:"forced_country"`
	IsStalker         sql.NullInt64  `json:"is_stalker"`
	BypassUA          sql.NullString `json:"bypass_ua"`
	AsNumber          sql.NullString `json:"as_number"`
	PlayToken         sql.NullString `json:"play_token"`
	PackageID         sql.NullInt64  `json:"package_id"`
	UsrMac            sql.NullString `json:"usr_mac"`
	UsrDeviceKey      sql.NullString `json:"usr_device_key"`
	Notes2            sql.NullString `json:"notes2"`
	RootEnabled       sql.NullInt64  `json:"root_enabled"`
	NumeroWhats       sql.NullString `json:"numero_whats"`
	NomeParaAviso     sql.NullString `json:"nome_para_aviso"`
	Email             sql.NullString `json:"email"`
	EnviarNotificacao sql.NullString `json:"enviar_notificacao"` // Alterado para `sql.NullString`
	SobrenomeAvisos   sql.NullString `json:"sobrenome_avisos"`
	Deleted           sql.NullInt64  `json:"deleted"`
	DateDeleted       sql.NullString `json:"date_deleted"`
	AppID             sql.NullString `json:"app_id"`
	TrustRenew        sql.NullInt64  `json:"trust_renew"`
	Franquia          sql.NullString `json:"franquia"`
	FranquiaMemberID  sql.NullInt64  `json:"franquia_member_id"`
	P2P               sql.NullInt64  `json:"p2p"`
}

// ClientDataSwagger representa o modelo para documentação no Swagger
//
// swagger:model
type ClientDataSwagger struct {
	ID                int    `json:"id"`                 // ID do cliente
	MemberID          int    `json:"member_id"`          // ID do membro associado
	Username          string `json:"username"`           // Nome de usuário
	Password          string `json:"password"`           // Senha (hash)
	ExpDate           string `json:"exp_date"`           // Data de expiração (formato ISO 8601)
	AdminEnabled      int    `json:"admin_enabled"`      // Conta ativa para admin (0 ou 1)
	Enabled           int    `json:"enabled"`            // Conta ativa para o usuário (0 ou 1)
	AdminNotes        string `json:"admin_notes"`        // Notas administrativas
	ResellerNotes     string `json:"reseller_notes"`     // Notas do revendedor
	Bouquet           string `json:"bouquet"`            // Pacotes assinados
	MaxConnections    int    `json:"max_connections"`    // Máximo de conexões permitidas
	IsRestreamer      int    `json:"is_restreamer"`      // Indica se é restreamer (0 ou 1)
	AllowedIPs        string `json:"allowed_ips"`        // IPs permitidos
	AllowedUA         string `json:"allowed_ua"`         // User-agents permitidos
	IsTrial           int    `json:"is_trial"`           // Indica se é conta de teste (0 ou 1)
	CreatedAt         string `json:"created_at"`         // Data de criação da conta
	CreatedBy         string `json:"created_by"`         // Criado por (usuário)
	PairID            int    `json:"pair_id"`            // ID do par
	IsMag             int    `json:"is_mag"`             // Indica se é dispositivo MAG (0 ou 1)
	IsE2              int    `json:"is_e2"`              // Indica se é Enigma2 (0 ou 1)
	ForceServerID     int    `json:"force_server_id"`    // ID do servidor forçado
	IsIspLock         int    `json:"is_isplock"`         // Indica se é bloqueado por ISP (0 ou 1)
	IspDesc           string `json:"isp_desc"`           // Descrição do ISP
	ForcedCountry     string `json:"forced_country"`     // País forçado
	IsStalker         int    `json:"is_stalker"`         // Indica se é Stalker (0 ou 1)
	BypassUA          string `json:"bypass_ua"`          // User-agent ignorado
	AsNumber          string `json:"as_number"`          // Número do AS
	PlayToken         string `json:"play_token"`         // Token de reprodução
	PackageID         int    `json:"package_id"`         // ID do pacote
	UsrMac            string `json:"usr_mac"`            // MAC do usuário
	UsrDeviceKey      string `json:"usr_device_key"`     // Chave do dispositivo do usuário
	Notes2            string `json:"notes2"`             // Notas adicionais
	RootEnabled       int    `json:"root_enabled"`       // Indica se é root (0 ou 1)
	NumeroWhats       string `json:"numero_whats"`       // Número do WhatsApp
	NomeParaAviso     string `json:"nome_para_aviso"`    // Nome para aviso
	Email             string `json:"email"`              // Email do usuário
	EnviarNotificacao string `json:"enviar_notificacao"` // Indica se recebe notificações (0 ou 1)
	SobrenomeAvisos   string `json:"sobrenome_avisos"`   // Sobrenome para avisos
	Deleted           int    `json:"deleted"`            // Indica se foi deletado (0 ou 1)
	DateDeleted       string `json:"date_deleted"`       // Data de exclusão
	AppID             string `json:"app_id"`             // ID do aplicativo
	TrustRenew        int    `json:"trust_renew"`        // Indica se a renovação é confiável
	Franquia          string `json:"franquia"`           // Nome da franquia
	FranquiaMemberID  int    `json:"franquia_member_id"` // ID do membro da franquia
	P2P               int    `json:"p2p"`                // Indica se é P2P (0 ou 1)
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
		SOBRENOME_AVISOS, deleted, date_deleted, app_id, trust_renew, franquia, franquia_member_id, p2p
		FROM users
		WHERE %s
	`, strings.Join(conditions, " AND ")) // 📌 Correção aplicada aqui

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
		); err != nil {
			log.Printf("Erro ao escanear linha: %v\n", err)
			return nil, err
		}
		clients = append(clients, client)
	}

	log.Printf("Total de registros encontrados: %d\n", len(clients))
	return clients, nil
}
