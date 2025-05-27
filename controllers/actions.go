package controllers

import (
	"apiBackEnd/config"
	"apiBackEnd/models"
	"apiBackEnd/utils"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"log"

	"github.com/gin-gonic/gin"
)

// --- Structs para Swagger ---
type TrustBonusRequest struct {
	UserID          int    `json:"user_id"`
	DiasAdicionados int    `json:"dias_adicionados"`
	Motivo          string `json:"motivo"`
}

type RenewRollbackRequest struct {
	UserID int `json:"user_id"`
}

type ChangeDueDateRequest struct {
	UserID             int    `json:"user_id"`
	NovaDataVencimento int    `json:"nova_data_vencimento"`
	Motivo             string `json:"motivo"`
}

// Liberação por confiança
//
// @Summary Liberação por confiança (dias extras)
// @Description Adiciona dias extras à conta do usuário, conforme regras parametrizadas. Só é permitido para contas vencidas, quantidade de dias entre CONFIANCA_DIAS_MIN e CONFIANCA_DIAS_MAX, e apenas uma vez a cada CONFIANCA_FREQUENCIA_DIAS dias.
// @Tags Ações
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body TrustBonusRequest true "Exemplo: {\"user_id\": 123, \"dias_adicionados\": 2, \"motivo\": \"Cortesia técnica\"}"
// @Success 200 {object} map[string]interface{} "Exemplo de resposta: {\"sucesso\": true, \"novo_exp_date\": 1716403200}"
// @Failure 400 {object} map[string]string "Erro de validação ou regra de negócio"
// @Failure 401 {object} map[string]string "Token inválido ou não fornecido"
// @Failure 500 {object} map[string]string "Erro interno"
// @Router /api/trust-bonus [post]
func TrustBonusHandler(c *gin.Context) {
	tokenInfo, ok := utils.ValidateAndExtractToken(c)
	if !ok {
		return
	}

	var req TrustBonusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos"})
		return
	}

	var userMemberID int
	var expDate sql.NullInt64
	err := config.DB.QueryRow("SELECT member_id, exp_date FROM streamcreed_db.users WHERE id = ?", req.UserID).Scan(&userMemberID, &expDate)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "Usuário não encontrado"})
			return
		}
		log.Printf("❌ Erro ao buscar usuário: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar usuário"})
		return
	}
	if userMemberID != tokenInfo.MemberID {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Ação não permitida: você não é o responsável por esta conta"})
		return
	}

	// --- BLOQUEIO DE FREQUÊNCIA ---
	redisKey := "trust_bonus:" + strconv.Itoa(req.UserID)
	val, err := config.RedisClient.Get(c, redisKey).Result()
	if err == nil && val != "" {
		var lastBonus models.TrustBonus
		if err := json.Unmarshal([]byte(val), &lastBonus); err == nil {
			freqDias := utils.GetConfiancaFrequenciaDias()
			msg := "Não é possível liberar bônus pois já foi concedido em: " +
				lastBonus.DataLiberacao.Format("02/01/2006 15:04") +
				". Essa liberação só pode ser feita a cada " + strconv.Itoa(freqDias) + " dias."
			c.JSON(http.StatusBadRequest, gin.H{"erro": msg})
			return
		}
	}
	// --- FIM DO BLOQUEIO ---

	// Validação: só permite bônus se a conta estiver vencida
	now := time.Now().Unix()
	if !expDate.Valid || expDate.Int64 > now {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Só é possível liberar bônus para contas vencidas."})
		return
	}

	// Validação: dias dentro do limite do .env
	minDias := utils.GetConfiancaDiasMin()
	maxDias := utils.GetConfiancaDiasMax()
	if req.DiasAdicionados < minDias || req.DiasAdicionados > maxDias {
		c.JSON(http.StatusBadRequest, gin.H{
			"erro": "Quantidade de dias inválida. Permitido entre " + strconv.Itoa(minDias) + " e " + strconv.Itoa(maxDias) + ".",
		})
		return
	}

	// Atualiza exp_date e enabled no banco
	newExpDate := now + int64(req.DiasAdicionados*86400)
	_, err = config.DB.Exec("UPDATE streamcreed_db.users SET exp_date = ?, enabled = 1 WHERE id = ?", newExpDate, req.UserID)
	if err != nil {
		log.Printf("❌ Erro ao atualizar exp_date: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao atualizar exp_date"})
		return
	}

	// Salva sessão no Redis
	bonus := models.TrustBonus{
		DiasAdicionados: req.DiasAdicionados,
		DataLiberacao:   time.Now(),
		AdminID:         tokenInfo.Username,
		Motivo:          req.Motivo,
	}
	// Passa o contexto da requisição para SaveToRedisJSON
	if err := utils.SaveToRedisJSON(c, redisKey, bonus, utils.GetConfiancaFrequenciaDias()*86400); err != nil {
		log.Printf("❌ Erro ao salvar bônus de confiança no Redis para chave %s: %v", redisKey, err)
		// Considerar se deve retornar erro ao cliente ou apenas logar
	}

	// Log MongoDB
	_ = utils.SaveActionLog(req.UserID, "trust_bonus", bonus, tokenInfo.Username)

	// Resposta padronizada de sucesso
	log.Printf("✅ Trust bonus aplicado com sucesso para usuário %d", req.UserID)
	c.JSON(http.StatusOK, gin.H{
		"sucesso":          true,
		"message":          "Bônus de confiança aplicado com sucesso",
		"novo_exp_date":    newExpDate,
		"dias_adicionados": req.DiasAdicionados,
		"usuario_id":       req.UserID,
	})
}

// Rollback de renovação
//
// @Summary Rollback de renovação
// @Description Desfaz a última renovação: restaura exp_date e créditos a partir do backup da última renovação. Só pode ser feito uma vez a cada ROLLBACK_PERMITIDO_FREQUENCIA dias e dentro do período de ROLLBACK_PERMITIDO_DIAS após a renovação.
// @Tags Ações
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body RenewRollbackRequest true "Exemplo: {\"user_id\": 123}"
// @Success 200 {object} map[string]interface{} "Exemplo de resposta: {\"sucesso\": true, \"exp_date_anterior\": 1716403200, \"exp_date_restaurado\": 1716403200, \"creditos_devolvidos\": 3}"
// @Failure 400 {object} map[string]string "Erro de validação ou regra de negócio"
// @Failure 401 {object} map[string]string "Token inválido ou não fornecido"
// @Failure 500 {object} map[string]string "Erro interno"
// @Router /api/renew-rollback [post]
func RenewRollbackHandler(c *gin.Context) {
	tokenInfo, ok := utils.ValidateAndExtractToken(c)
	if !ok {
		return
	}

	var req RenewRollbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos"})
		return
	}

	var userMemberID int
	var expDate sql.NullInt64
	err := config.DB.QueryRow("SELECT member_id, exp_date FROM streamcreed_db.users WHERE id = ?", req.UserID).Scan(&userMemberID, &expDate)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "Usuário não encontrado"})
			return
		}
		log.Printf("❌ Erro ao buscar usuário: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar usuário"})
		return
	}
	if userMemberID != tokenInfo.MemberID {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Ação não permitida: você não é o responsável por esta conta"})
		return
	}
	if !expDate.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "exp_date inválido"})
		return
	}

	// --- BLOQUEIO DE FREQUÊNCIA ---
	rollbackKey := "rollback_lock:" + strconv.Itoa(req.UserID)
	val, err := config.RedisClient.Get(c, rollbackKey).Result() // c (gin.Context) é usado aqui
	if err == nil && val != "" {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Reversão já realizado recentemente para esta conta. Só é permitido uma reversão a cada " + strconv.Itoa(utils.GetRollbackPermitidoFrequencia()) + " dias."})
		return
	}
	// --- FIM DO BLOQUEIO ---

	// --- BUSCA BACKUP DA ÚLTIMA RENOVAÇÃO ---
	backupKey := "renew_backup:" + strconv.Itoa(req.UserID)
	backupVal, err := config.RedisClient.Get(c, backupKey).Result() // c (gin.Context) é usado aqui
	if err != nil || backupVal == "" {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Não há backup de renovação disponível para reversão ou o período expirou."})
		return
	}
	var backup models.RenewBackup
	if err := json.Unmarshal([]byte(backupVal), &backup); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao ler backup de renovação"})
		return
	}

	// --- VALIDAÇÃO DE JANELA DE ROLLBACK ---
	rollbackDias := utils.GetRollbackPermitidoDias()
	if time.Since(backup.DataRenovacao) > time.Duration(rollbackDias)*24*time.Hour {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "O prazo para reversão já expirou. Permitido até " + strconv.Itoa(rollbackDias) + " dias após a renovação."})
		return
	}

	// --- EXECUTA ROLLBACK ---
	tx, err := config.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao iniciar transação"})
		return
	}

	// Buscar o username do usuário para o log
	var username string
	err = tx.QueryRow("SELECT username FROM streamcreed_db.users WHERE id = ?", req.UserID).Scan(&username)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar dados do usuário"})
		return
	}

	_, err = tx.Exec("UPDATE streamcreed_db.users SET exp_date = ? WHERE id = ?", backup.ExpDateAnterior, req.UserID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao restaurar exp_date"})
		return
	}
	_, err = tx.Exec("UPDATE streamcreed_db.reg_users SET credits = credits + ? WHERE id = ?", backup.CreditosGastos, tokenInfo.MemberID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao devolver créditos, tente novamente mais tarde."})
		return
	}

	// Inserir log de créditos
	logReason := "Reversão de renovação do " + username
	_, err = tx.Exec("INSERT INTO streamcreed_db.credits_log (target_id, admin_id, amount, `date`, reason) VALUES (?, -1, ?, ?, ?)",
		tokenInfo.MemberID, backup.CreditosGastos, time.Now().Unix(), logReason)
	if err != nil {
		tx.Rollback()
		log.Printf("❌ Erro ao inserir log de créditos: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao registrar log de créditos"})
		return
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao finalizar transação"})
		return
	}

	// --- BLOQUEIO DE NOVO ROLLBACK ---
	ttl := utils.GetRollbackPermitidoFrequencia() * 86400
	// Usar c (gin.Context) para o Set do Redis
	if err := config.RedisClient.Set(c, rollbackKey, "1", time.Duration(ttl)*time.Second).Err(); err != nil {
		log.Printf("❌ Erro ao bloquear novo reversão no Redis para chave %s: %v", rollbackKey, err)
	}

	// --- REMOVE O BACKUP DO REDIS APÓS ROLLBACK EXECUTADO ---
	if err := config.RedisClient.Del(c, backupKey).Err(); err != nil {
		log.Printf("❌ Erro ao remover backup de renovação no Redis para chave %s: %v", backupKey, err)
	} else {
		log.Printf("✅ Backup de renovação removido do Redis para chave %s", backupKey)
	}

	// Log MongoDB
	_ = utils.SaveActionLog(req.UserID, "renew_rollback", backup, tokenInfo.Username)

	c.JSON(http.StatusOK, gin.H{
		"sucesso":             true,
		"exp_date_anterior":   expDate.Int64,
		"exp_date_restaurado": backup.ExpDateAnterior,
		"creditos_devolvidos": backup.CreditosGastos,
	})
}

// Alteração de vencimento
//
// @Summary Alteração da data mensal de vencimento
// @Description Altera o dia do vencimento da conta para o mês atual. Só pode ser feito uma vez a cada ALTERACAO_VENCIMENTO_FREQUENCIA_DIAS dias e o dia deve ser válido para o mês.
// @Tags Ações
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body ChangeDueDateRequest true "Exemplo: {\"user_id\": 123, \"nova_data_vencimento\": 15, \"motivo\": \"Pedido do cliente\"}"
// @Success 200 {object} map[string]interface{} "Exemplo de resposta: {\"sucesso\": true, \"novo_exp_date\": 1716403200}"
// @Failure 400 {object} map[string]string "Erro de validação ou regra de negócio"
// @Failure 401 {object} map[string]string "Token inválido ou não fornecido"
// @Failure 500 {object} map[string]string "Erro interno"
// @Router /api/change-due-date [post]
func ChangeDueDateHandler(c *gin.Context) {
	tokenInfo, ok := utils.ValidateAndExtractToken(c)
	if !ok {
		return
	}

	var req ChangeDueDateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos"})
		return
	}

	var userMemberID int
	var expDate sql.NullInt64
	var isTrial int
	var enabled int
	err := config.DB.QueryRow("SELECT member_id, exp_date, is_trial, enabled FROM streamcreed_db.users WHERE id = ?", req.UserID).Scan(&userMemberID, &expDate, &isTrial, &enabled)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "Usuário não encontrado"})
			return
		}
		log.Printf("❌ Erro ao buscar usuário: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar usuário"})
		return
	}
	if userMemberID != tokenInfo.MemberID {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Ação não permitida: você não é o responsável por esta conta"})
		return
	}
	if isTrial == 1 {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Não é permitido alterar o vencimento de contas de teste (trial)."})
		return
	}
	if enabled != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Não é permitido alterar o vencimento de contas bloqueadas/desativadas."})
		return
	}

	// --- NOVA VERIFICAÇÃO DE FREQUÊNCIA ---
	redisKey := "change_due_date:" + strconv.Itoa(req.UserID)
	val, err := config.RedisClient.Get(c, redisKey).Result()
	if err == nil && val != "" {
		var lastChange models.ChangeDueDate
		if err := json.Unmarshal([]byte(val), &lastChange); err == nil {
			freqDias := utils.GetAlteracaoVencimentoFrequenciaDias()
			msg := "Não é possível efetuar alteração pois já foi feita em: " +
				lastChange.DataAlteracao.Format("02/01/2006 15:04") +
				". Essa alteração só pode ser feita a cada " + strconv.Itoa(freqDias) + " dias."
			c.JSON(http.StatusBadRequest, gin.H{"erro": msg})
			return
		}
	}
	// --- FIM DA VERIFICAÇÃO ---

	if !expDate.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "exp_date inválido"})
		return
	}
	exp := time.Unix(expDate.Int64, 0)
	firstOfNextMonth := time.Date(exp.Year(), exp.Month()+1, 1, 0, 0, 0, 0, exp.Location())
	lastDayOfMonth := firstOfNextMonth.AddDate(0, 0, -1).Day()

	if req.NovaDataVencimento < 1 || req.NovaDataVencimento > lastDayOfMonth {
		c.JSON(http.StatusBadRequest, gin.H{
			"erro": "O dia informado é inválido para o mês do vencimento atual. Último dia permitido: " + strconv.Itoa(lastDayOfMonth),
		})
		return
	}

	// Corrigido: altera apenas o dia, mantendo mês, ano e horário do exp_date original
	novoVenc := time.Date(exp.Year(), exp.Month(), req.NovaDataVencimento, exp.Hour(), exp.Minute(), exp.Second(), exp.Nanosecond(), exp.Location())
	if novoVenc.Before(exp) {
		// Se a nova data ficou antes do exp_date original, ajusta para o próximo mês
		novoVenc = time.Date(exp.Year(), exp.Month()+1, req.NovaDataVencimento, exp.Hour(), exp.Minute(), exp.Second(), exp.Nanosecond(), exp.Location())
	}

	_, err = config.DB.Exec("UPDATE streamcreed_db.users SET exp_date = ? WHERE id = ?", novoVenc.Unix(), req.UserID)
	if err != nil {
		log.Printf("❌ Erro ao atualizar exp_date: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao atualizar exp_date"})
		return
	}

	change := models.ChangeDueDate{
		NovaDataVencimento: req.NovaDataVencimento,
		DataAlteracao:      time.Now(),
		AdminID:            tokenInfo.Username,
		Motivo:             req.Motivo,
	}
	ttl := utils.GetAlteracaoVencimentoFrequenciaDias() * 86400
	// Passa o contexto da requisição para SaveToRedisJSON
	if err := utils.SaveToRedisJSON(c, redisKey, change, ttl); err != nil {
		log.Printf("❌ Erro ao salvar alteração de vencimento no Redis para chave %s: %v", redisKey, err)
		// Considerar se deve retornar erro ao cliente ou apenas logar
	}
	_ = utils.SaveActionLog(req.UserID, "change_due_date", change, tokenInfo.Username)

	// Resposta padronizada de sucesso
	log.Printf("✅ Data de vencimento alterada com sucesso para usuário %d", req.UserID)
	c.JSON(http.StatusOK, gin.H{
		"sucesso":              true,
		"message":              "Data de vencimento alterada com sucesso",
		"novo_exp_date":        novoVenc.Unix(),
		"nova_data_vencimento": req.NovaDataVencimento,
		"usuario_id":           req.UserID,
	})
}
