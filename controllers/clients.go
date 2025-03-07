package controllers

import (
	"apiBackEnd/models"
	"apiBackEnd/utils"
	"database/sql"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// Estrutura completa para o JSON de resposta
type ClientResponse struct {
	ID                int    `json:"id"`
	MemberID          int    `json:"member_id"`
	Username          string `json:"username"`
	Password          string `json:"password"`
	ExpDate           string `json:"exp_date"`
	AdminEnabled      int    `json:"admin_enabled"`
	Enabled           int    `json:"enabled"`
	AdminNotes        string `json:"admin_notes"`
	ResellerNotes     string `json:"reseller_notes"`
	Bouquet           string `json:"bouquet"`
	MaxConnections    int    `json:"max_connections"`
	IsRestreamer      int    `json:"is_restreamer"`
	AllowedIPs        string `json:"allowed_ips"`
	AllowedUA         string `json:"allowed_ua"`
	IsTrial           int    `json:"is_trial"`
	CreatedAt         string `json:"created_at"`
	CreatedBy         string `json:"created_by"`
	PairID            int    `json:"pair_id"`
	IsMag             int    `json:"is_mag"`
	IsE2              int    `json:"is_e2"`
	ForceServerID     int    `json:"force_server_id"`
	IsIspLock         int    `json:"is_isplock"`
	IspDesc           string `json:"isp_desc"`
	ForcedCountry     string `json:"forced_country"`
	IsStalker         int    `json:"is_stalker"`
	BypassUA          string `json:"bypass_ua"`
	AsNumber          string `json:"as_number"`
	PlayToken         string `json:"play_token"`
	PackageID         int    `json:"package_id"`
	UsrMac            string `json:"usr_mac"`
	UsrDeviceKey      string `json:"usr_device_key"`
	Notes2            string `json:"notes2"`
	RootEnabled       int    `json:"root_enabled"`
	NumeroWhats       string `json:"numero_whats"`
	NomeParaAviso     string `json:"nome_para_aviso"`
	Email             string `json:"email"`
	EnviarNotificacao bool   `json:"enviar_notificacao"`
	SobrenomeAvisos   string `json:"sobrenome_avisos"`
	Deleted           int    `json:"deleted"`
	DateDeleted       string `json:"date_deleted"`
	AppID             string `json:"app_id"`
	TrustRenew        int    `json:"trust_renew"`
	Franquia          string `json:"franquia"`
	FranquiaMemberID  int    `json:"franquia_member_id"`
	P2P               int    `json:"p2p"`
}

// GetClients retorna a lista de clientes de um membro autenticado.
//
// @Summary Lista clientes
// @Description Retorna todos os clientes associados ao usuário autenticado
// @Tags Clientes
// @Security BearerAuth
// @Accept  json
// @Produce  json
// @Success 200 {object} []models.ClientDataSwagger "Lista de clientes"
// @Failure 401 {object} map[string]string "Token inválido ou não fornecido"
// @Failure 500 {object} map[string]string "Erro interno ao buscar clientes"
// @Router /api/clients [get]
func GetClients(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token não fornecido"})
		return
	}

	claims, err := utils.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token inválido"})
		return
	}

	memberIDFloat, exists := claims["member_id"].(float64)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "MemberID não encontrado no token"})
		return
	}

	memberID := int(memberIDFloat)

	// Capturar filtros opcionais da URL
	filters := map[string]interface{}{}
	queryParams := []string{"username", "numero_whats", "enviar_notificacao", "max_connections", "is_trial", "enabled", "admin_notes", "email", "exp_date"}
	for _, q := range queryParams {
		if value := c.Query(q); value != "" {
			filters[q] = value
		}
	}

	clients, err := models.GetClientsByFilters(memberID, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar clientes"})
		return
	}

	// Criar a lista de resposta formatada corretamente
	responseClients := make([]ClientResponse, len(clients))
	var wg sync.WaitGroup
	wg.Add(len(clients))

	for i, client := range clients {
		go func(i int, client models.ClientData) {
			defer wg.Done()
			responseClients[i] = ClientResponse{
				ID:                client.ID,
				MemberID:          client.MemberID,
				Username:          client.Username,
				Password:          client.Password,
				ExpDate:           NullStringToString(client.ExpDate),
				AdminEnabled:      NullIntToInt(client.AdminEnabled),
				Enabled:           NullIntToInt(client.Enabled),
				AdminNotes:        NullStringToString(client.AdminNotes),
				ResellerNotes:     NullStringToString(client.ResellerNotes),
				Bouquet:           NullStringToString(client.Bouquet),
				MaxConnections:    NullIntToInt(client.MaxConnections),
				IsRestreamer:      NullIntToInt(client.IsRestreamer),
				AllowedIPs:        NullStringToString(client.AllowedIPs),
				AllowedUA:         NullStringToString(client.AllowedUA),
				IsTrial:           NullIntToInt(client.IsTrial),
				CreatedAt:         NullStringToString(client.CreatedAt),
				CreatedBy:         NullStringToString(client.CreatedBy),
				PairID:            NullIntToInt(client.PairID),
				IsMag:             NullIntToInt(client.IsMag),
				IsE2:              NullIntToInt(client.IsE2),
				ForceServerID:     NullIntToInt(client.ForceServerID),
				IsIspLock:         NullIntToInt(client.IsIspLock),
				IspDesc:           NullStringToString(client.IspDesc),
				ForcedCountry:     NullStringToString(client.ForcedCountry),
				IsStalker:         NullIntToInt(client.IsStalker),
				BypassUA:          NullStringToString(client.BypassUA),
				AsNumber:          NullStringToString(client.AsNumber),
				PlayToken:         NullStringToString(client.PlayToken),
				PackageID:         NullIntToInt(client.PackageID),
				UsrMac:            NullStringToString(client.UsrMac),
				UsrDeviceKey:      NullStringToString(client.UsrDeviceKey),
				Notes2:            NullStringToString(client.Notes2),
				RootEnabled:       NullIntToInt(client.RootEnabled),
				NumeroWhats:       NullStringToString(client.NumeroWhats),
				NomeParaAviso:     NullStringToString(client.NomeParaAviso),
				Email:             NullStringToString(client.Email),
				EnviarNotificacao: NullStringToBool(client.EnviarNotificacao),
				SobrenomeAvisos:   NullStringToString(client.SobrenomeAvisos),
				Deleted:           NullIntToInt(client.Deleted),
				DateDeleted:       NullStringToString(client.DateDeleted),
				AppID:             NullStringToString(client.AppID),
				TrustRenew:        NullIntToInt(client.TrustRenew),
				Franquia:          NullStringToString(client.Franquia),
				FranquiaMemberID:  NullIntToInt(client.FranquiaMemberID),
				P2P:               NullIntToInt(client.P2P),
			}
		}(i, client)
	}

	wg.Wait() // Espera todas as goroutines terminarem

	c.JSON(http.StatusOK, gin.H{
		"total_registros": len(responseClients),
		"clientes":        responseClients,
	})
}

// Converter `sql.NullString` para string normal
func NullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// Converter `sql.NullInt64` para int normal
func NullIntToInt(ni sql.NullInt64) int {
	if ni.Valid {
		return int(ni.Int64)
	}
	return 0
}

// Converter `sql.NullString` para `bool`
func NullStringToBool(ns sql.NullString) bool {
	if ns.Valid {
		return ns.String == "true" // Se for "true", retorna `true`, senão `false`
	}
	return false // Se for NULL, retorna `false`
}
