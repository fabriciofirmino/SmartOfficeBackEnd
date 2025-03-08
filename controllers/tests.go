package controllers

import (
	"apiBackEnd/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Mapa de bloqueios por IP
var blockedIPs = make(map[string]time.Time)
var mu sync.Mutex // Para evitar concorrência no acesso ao mapa

// Estrutura para receber os dados do front-end
type TestRequest struct {
	Username      string `json:"username,omitempty"` // Pode ser opcional
	Password      string `json:"password,omitempty"` // Sempre será gerado
	NumeroWhats   string `json:"numero_whats"`
	NomeParaAviso string `json:"nome_para_aviso"`
}

// CreateTest cria um novo teste IPTV.
//
// @Summary Criar Teste IPTV
// @Description Gera um usuário e senha de teste para IPTV e retorna as credenciais.
// @Tags Testes IPTV
// @Security BearerAuth
// @Accept  json
// @Produce  json
// @Param test body controllers.TestRequest true "Dados para criação do teste"
// @Success 200 {object} map[string]interface{} "Teste criado com sucesso"
// @Failure 400 {object} map[string]string "Erro na requisição"
// @Failure 401 {object} map[string]string "Token inválido"
// @Router /api/create-test [post]
func CreateTest(c *gin.Context) {
	ip := c.ClientIP()

	// 🔥 **VERIFICAR SE O IP ESTÁ BLOQUEADO**
	mu.Lock()
	blockTime, blocked := blockedIPs[ip]
	mu.Unlock()

	if blocked && time.Now().Before(blockTime) {
		c.JSON(http.StatusTooManyRequests, gin.H{"erro": "Muitas tentativas falharam. Tente novamente mais tarde."})
		return
	}

	// **1️⃣ Autenticação obrigatória**
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token não fornecido"})
		return
	}

	// 📌 Ajuste para capturar corretamente os três valores retornados por `ValidateToken`
	claims, _, err := utils.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token inválido"})
		return
	}

	memberIDFloat, exists := claims["member_id"].(float64)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "MemberID não encontrado no token"})
		return
	}

	// **2️⃣ Pegando valores do .env**
	apiURL := os.Getenv("IPTV_API_URL")
	expHours := os.Getenv("EXP_DATE")
	bouquet := os.Getenv("BOUQUET")
	totalUserChars, _ := strconv.Atoi(os.Getenv("TOTAL_CARACTERES_USER"))
	totalPassChars, _ := strconv.Atoi(os.Getenv("TOTAL_CARACTERES_SENHA"))
	prefixUser := os.Getenv("PREFIXO_USR")
	prefixPass := os.Getenv("PREFIXO_SENHA")

	// **3️⃣ Validar configuração**
	if apiURL == "" || expHours == "" || bouquet == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Configuração inválida no .env"})
		return
	}

	// **4️⃣ Ler dados do corpo da requisição**
	var req TestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos"})
		return
	}

	// **5️⃣ Limite de tentativas**
	attempts := 0
	maxAttempts := 3

	for attempts < maxAttempts {
		// 🔥 Sempre gerar um novo username em cada tentativa
		req.Username = utils.GenerateUsername(totalUserChars, prefixUser)

		// 🔥 Gerar senha aleatória
		req.Password = utils.GeneratePassword(totalPassChars, prefixPass)

		// 🔥 Gerar timestamp de expiração
		expTimestamp := utils.GenerateExpirationTimestamp(expHours)

		// Criar os dados do formulário x-www-form-urlencoded
		form := url.Values{}
		form.Add("action", "user")
		form.Add("sub", "create")
		form.Add("user_data[username]", req.Username)
		form.Add("user_data[password]", req.Password)
		form.Add("user_data[max_connections]", "1")
		form.Add("user_data[is_restreamer]", "0")
		form.Add("user_data[exp_date]", fmt.Sprintf("%d", expTimestamp))
		form.Add("user_data[bouquet]", bouquet)
		form.Add("user_data[member_id]", fmt.Sprintf("%d", int(memberIDFloat)))
		form.Add("user_data[is_trial]", "1")
		form.Add("user_data[NUMERO_WHATS]", req.NumeroWhats)
		form.Add("user_data[NOME_PARA_AVISO]", req.NomeParaAviso)
		form.Add("user_data[reseller_notes]", "Criado Via BOT")

		// Criar a requisição HTTP
		reqBody := bytes.NewBufferString(form.Encode())
		resp, err := http.Post(apiURL, "application/x-www-form-urlencoded", reqBody)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao criar teste"})
			return
		}
		defer resp.Body.Close()

		// **6️⃣ Decodificar a resposta da API**
		var responseMap map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseMap); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao processar resposta da API"})
			return
		}

		// 🔥 **7️⃣ Tentar extrair `exp_date` do JSON da API**
		var expirationTime string
		if expDate, exists := responseMap["exp_date"]; exists {
			expirationTime = utils.FormatTimestamp(expDate)
		} else {
			// 🔥 **8️⃣ Se `exp_date` não veio, calcular baseado no tempo do teste**
			durationHours, _ := strconv.Atoi(expHours)
			calculatedTime := time.Now().Add(time.Duration(durationHours) * time.Hour)
			expirationTime = calculatedTime.Format("02/01/2006 15:04") // 🔥 Formato dd/mm/aaaa hh:mm
		}

		// 🔥 **9️⃣ Adicionar `vencimento` ao JSON de resposta**
		responseMap["vencimento"] = expirationTime

		// **10️⃣ Se a resposta NÃO for "EXISTS", finaliza a criação**
		if result, exists := responseMap["result"].(bool); exists && result {
			fmt.Printf("✅ Usuário criado com sucesso: %s\n", req.Username)
			c.JSON(http.StatusOK, responseMap)
			return
		} else if errorMsg, exists := responseMap["error"].(string); exists && errorMsg == "EXISTS" {
			attempts++
			fmt.Printf("⚠️ Tentativa %d falhou! Usuário rejeitado: %s\n", attempts, req.Username)
			fmt.Printf("📢 Resposta da API: %+v\n", responseMap)
			time.Sleep(2 * time.Second) // 🔥 Espera 2 segundos antes de tentar novamente
			continue
		}
	}

	// **11️⃣ Se atingir o limite de tentativas, bloqueia IP**
	mu.Lock()
	blockedIPs[ip] = time.Now().Add(2 * time.Minute)
	mu.Unlock()

	fmt.Println("❌ Muitas tentativas falharam! IP bloqueado por 2 minutos.")
	c.JSON(http.StatusTooManyRequests, gin.H{"erro": "Muitas tentativas falharam. IP bloqueado por 2 minutos."})
}
