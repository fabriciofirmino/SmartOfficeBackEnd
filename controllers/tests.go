package controllers

import (
	"apiBackEnd/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"log"

	"github.com/gin-gonic/gin"
)

// Mapa de bloqueios por IP
var blockedIPs = make(map[string]time.Time)
var mu sync.Mutex // Para evitar concorrência no acesso ao mapa

// Estrutura para receber os dados do front-end
type TestRequest struct {
	Username         string `json:"username,omitempty"` // Pode ser opcional
	Password         string `json:"password,omitempty"` // Sempre será gerado
	NumeroWhats      string `json:"numero_whats"`
	NomeParaAviso    string `json:"nome_para_aviso"`
	FranquiaMemberID *int   `json:"franquia_member_id,omitempty"` // Novo campo opcional
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
// @example request.body.random_user_pass
//
//	{
//	  "numero_whats": "+5511999998888",
//	  "nome_para_aviso": "Cliente Teste Geração Automática",
//	  "franquia_member_id": 123
//	}
//
// @example request.body.specific_user_pass
//
//	{
//	  "username": "usuario",
//	  "password": "senha123",
//	  "numero_whats": "+5511999997777",
//	  "nome_para_aviso": "Cliente Teste Específico",
//	  "franquia_member_id": 456
//	}
//
// @Success 200 {object} map[string]interface{} "Teste criado com sucesso"
// @Failure 400 {object} map[string]string "Erro na requisição ou usuário já existe (com credenciais fornecidas)"
// @Failure 401 {object} map[string]string "Token inválido"
// @Failure 429 {object} map[string]string "Muitas tentativas de geração aleatória falharam"
// @Failure 500 {object} map[string]string "Erro interno do servidor"
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

	// 🔥 Gerar timestamp de expiração uma vez, pois será usado em ambos os cenários
	expTimestamp := utils.GenerateExpirationTimestamp(expHours)

	// Cenário 1: Usuário e senha fornecidos na requisição
	if req.Username != "" && req.Password != "" {
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
		if req.FranquiaMemberID != nil {
			form.Add("user_data[franquia_member_id]", fmt.Sprintf("%d", *req.FranquiaMemberID))
		}

		reqBody := bytes.NewBufferString(form.Encode())
		log.Printf("ℹ️  [Usuário Fornecido] Enviando requisição para API IPTV. URL: %s, Corpo: %s", apiURL, form.Encode())
		resp, err := http.Post(apiURL, "application/x-www-form-urlencoded", reqBody)
		if err != nil {
			log.Printf("❌ [Usuário Fornecido] Erro ao fazer POST para API IPTV: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao criar teste com dados fornecidos"})
			return
		}
		defer resp.Body.Close()
		log.Printf("ℹ️  [Usuário Fornecido] Resposta da API IPTV recebida. Status: %s", resp.Status)

		responseBodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("❌ [Usuário Fornecido] Erro ao ler corpo da resposta: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao processar resposta da API (leitura)"})
			return
		}
		log.Printf("ℹ️  [Usuário Fornecido] Corpo da resposta (raw): %s", string(responseBodyBytes))
		resp.Body = io.NopCloser(bytes.NewBuffer(responseBodyBytes))

		var responseMap map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseMap); err != nil {
			log.Printf("❌ [Usuário Fornecido] Erro ao decodificar resposta: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao processar resposta da API (decodificação)"})
			return
		}

		if result, exists := responseMap["result"].(bool); exists && result {
			var expirationTime string
			if expDate, ok := responseMap["exp_date"]; ok {
				expirationTime = utils.FormatTimestamp(expDate)
			} else {
				durationHours, _ := strconv.Atoi(expHours)
				calculatedTime := time.Now().Add(time.Duration(durationHours) * time.Hour)
				expirationTime = calculatedTime.Format("02/01/2006 15:04")
			}
			responseMap["vencimento"] = expirationTime
			log.Printf("✅ [Usuário Fornecido] Usuário %s criado com sucesso.", req.Username)
			c.JSON(http.StatusOK, responseMap)
			return
		} else if errorMsg, exists := responseMap["error"].(string); exists && errorMsg == "EXISTS" {
			log.Printf("⚠️ [Usuário Fornecido] Usuário %s já existe.", req.Username)
			c.JSON(http.StatusBadRequest, gin.H{"erro": "Nome de usuário já em uso. Tente outro ou deixe em branco para geração automática."})
			return
		} else {
			log.Printf("❌ [Usuário Fornecido] Erro não esperado da API IPTV: %+v", responseMap)
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao criar teste, resposta inesperada da API."})
			return
		}
	} else {
		// Cenário 2: Usuário e/ou senha NÃO fornecidos - Lógica de geração aleatória e retentativas
		log.Printf("ℹ️  Username/Password não fornecidos. Iniciando geração aleatória.")
		attempts := 0
		maxAttempts := 3

		for attempts < maxAttempts {
			// 🔥 Sempre gerar um novo username em cada tentativa
			currentUsername := utils.GenerateUsername(totalUserChars, prefixUser)
			currentPassword := utils.GeneratePassword(totalPassChars, prefixPass)

			form := url.Values{}
			form.Add("action", "user")
			form.Add("sub", "create")
			form.Add("user_data[username]", currentUsername)
			form.Add("user_data[password]", currentPassword)
			form.Add("user_data[max_connections]", "1")
			form.Add("user_data[is_restreamer]", "0")
			form.Add("user_data[exp_date]", fmt.Sprintf("%d", expTimestamp))
			form.Add("user_data[bouquet]", bouquet)
			form.Add("user_data[member_id]", fmt.Sprintf("%d", int(memberIDFloat)))
			form.Add("user_data[is_trial]", "1")
			form.Add("user_data[NUMERO_WHATS]", req.NumeroWhats)
			form.Add("user_data[NOME_PARA_AVISO]", req.NomeParaAviso)
			form.Add("user_data[reseller_notes]", "Criado Via BOT")
			if req.FranquiaMemberID != nil {
				form.Add("user_data[franquia_member_id]", fmt.Sprintf("%d", *req.FranquiaMemberID))
			}

			reqBody := bytes.NewBufferString(form.Encode())
			log.Printf("ℹ️  [Geração Aleatória Attempt %d] Enviando requisição. URL: %s, Corpo: %s", attempts+1, apiURL, form.Encode())
			resp, err := http.Post(apiURL, "application/x-www-form-urlencoded", reqBody)
			if err != nil {
				log.Printf("❌ [Geração Aleatória Attempt %d] Erro ao fazer POST para API IPTV: %v", attempts+1, err)
				c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao criar teste (geração aleatória)"})
				return
			}
			defer resp.Body.Close()

			log.Printf("ℹ️  [Geração Aleatória Attempt %d] Resposta da API IPTV recebida. Status: %s", attempts+1, resp.Status)

			responseBodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("❌ [Geração Aleatória Attempt %d] Erro ao ler corpo da resposta: %v", attempts+1, err)
				c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao processar resposta da API (leitura)"})
				return
			}
			log.Printf("ℹ️  [Geração Aleatória Attempt %d] Corpo da resposta (raw): %s", attempts+1, string(responseBodyBytes))
			resp.Body = io.NopCloser(bytes.NewBuffer(responseBodyBytes))

			var responseMap map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&responseMap); err != nil {
				log.Printf("❌ [Geração Aleatória Attempt %d] Erro ao decodificar resposta: %v", attempts+1, err)
				c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao processar resposta da API (decodificação)"})
				return
			}

			if result, exists := responseMap["result"].(bool); exists && result {
				log.Printf("✅ [Geração Aleatória] Usuário %s criado com sucesso.", currentUsername)
				c.JSON(http.StatusOK, responseMap)
				return
			} else if errorMsg, exists := responseMap["error"].(string); exists && errorMsg == "EXISTS" {
				attempts++
				log.Printf("⚠️ [Geração Aleatória Attempt %d] Usuário %s rejeitado (EXISTS).", attempts, currentUsername)
				log.Printf("📢 Resposta da API: %+v", responseMap)
				if attempts >= maxAttempts {
					break // Sai do loop se exceder tentativas
				}
				time.Sleep(1 * time.Second) // Reduzido para 1 segundo para testes mais rápidos
				continue
			} else {
				log.Printf("❌ [Geração Aleatória Attempt %d] Erro não esperado da API IPTV para usuário %s: %+v", attempts+1, currentUsername, responseMap)
				c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao criar teste (geração aleatória), resposta inesperada da API."})
				return // Sai em caso de erro inesperado da API
			}
		}

		// **11️⃣ Se atingir o limite de tentativas, bloqueia IP**
		if attempts >= maxAttempts {
			mu.Lock()
			blockedIPs[ip] = time.Now().Add(2 * time.Minute)
			mu.Unlock()
			log.Println("❌ Muitas tentativas de geração aleatória falharam! IP bloqueado por 2 minutos.")
			c.JSON(http.StatusTooManyRequests, gin.H{"erro": "Muitas tentativas de geração aleatória falharam. IP bloqueado por 2 minutos."})
		}
	}
}
