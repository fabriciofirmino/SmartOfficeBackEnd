package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Variáveis globais para armazenar dados do teste
var TestClientID int

func TestCreateClient(t *testing.T) {
	fmt.Println("🚀 Testando POST /api/create-test")

	testData := map[string]interface{}{
		"username":  "testuser_" + fmt.Sprintf("%d", time.Now().Unix()),
		"password":  "testpassword",
		"member_id": "17729", // Usa o member_id global extraído do token
	}
	// 🔹 Converte os dados para JSON
	jsonData, err := json.Marshal(testData)
	if err != nil {
		t.Fatalf("Erro ao converter dados para JSON: %v", err)
	}

	req, _ := http.NewRequest("POST", "/api/create-test", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+TestToken) // ✅ Garante que está usando o token correto
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	SetupServer().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 🔹 Verifica o ID do cliente criado
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if id, ok := response["created_id"].(float64); ok { // 🔹 Mudamos de "id_cliente" para "created_id"
		TestClientID = int(id)
		fmt.Printf("✅ Cliente de teste criado com ID: %d\n", TestClientID)
	} else {
		t.Fatalf("❌ Erro ao extrair o ID do cliente! Resposta recebida: %v", response)
	}

}
