package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 🔹 **Remova** a declaração duplicada das variáveis `TestToken` e `TestMemberID`!

func TestLoginSuccess(t *testing.T) {
	fmt.Println("🚀 Testando POST /login")

	loginData := map[string]string{
		"username": os.Getenv("TEST_USER"),
		"password": os.Getenv("TEST_PASSWORD"),
	}

	jsonData, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	SetupServer().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// 🔹 Salva o token globalmente (já definido em utils_test.go)
	if token, ok := response["token"].(string); ok {
		TestToken = token
		fmt.Println("✅ TestToken gerado:", TestToken)
	} else {
		t.Fatalf("❌ Erro ao extrair token do login!")
	}

	// 🔹 Extrai o `member_id` do token (já definido em utils_test.go)
	if id, ok := response["member_id"].(float64); ok {
		TestMemberID = int(id)
		fmt.Println("✅ TestMemberID extraído:", TestMemberID)
	} else {
		t.Fatalf("❌ Erro ao extrair MemberID!")
	}
}
