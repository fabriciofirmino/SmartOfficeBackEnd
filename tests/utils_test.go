package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// 🔹 TestToken e TestMemberID são globais para reutilização nos testes
var TestToken string
var TestMemberID int

var totalTests int
var passedTests int
var failedTests int

// EnsureAuthToken verifica se já existe um token ou gera um novo
func EnsureAuthToken(t *testing.T) {
	if TestToken != "" {
		fmt.Println("🔄 Reutilizando TestToken existente")
		return
	}
	if TestToken == "" {
		TestLoginSuccess(t) // 🔹 Garante que o token será gerado antes de outros testes
	}

	fmt.Println("🚀 Gerando novo TestToken")

	loginData := map[string]string{
		"username": os.Getenv("TEST_USER"),
		"password": os.Getenv("TEST_PASSWORD"),
	}

	jsonData, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	SetupServer().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("❌ Falha no login! Código HTTP: %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// 🔹 Salva o token globalmente
	if token, ok := response["token"].(string); ok {
		TestToken = token
		fmt.Println("✅ TestToken gerado:", TestToken)
	} else {
		t.Fatalf("❌ Erro ao extrair token do login!")
	}

	// 🔹 Extrai o `member_id` do token para testes futuros
	if id, ok := response["member_id"].(float64); ok {
		TestMemberID = int(id)
		fmt.Println("✅ TestMemberID extraído:", TestMemberID)
	} else {
		t.Fatalf("❌ Erro ao extrair MemberID!")
	}
}

// 🔹 Armazena os resultados individuais dos testes e atualiza a contagem
var testResults []string

func recordTestResult(testName string, success bool) {
	totalTests++

	if success {
		passedTests++
		testResults = append(testResults, fmt.Sprintf("✅ %s passou!", testName))
	} else {
		failedTests++
		testResults = append(testResults, fmt.Sprintf("❌ %s falhou!", testName))
	}
}

func printTestSummary() {
	fmt.Println("\n🔹🔹🔹 RESUMO DOS TESTES 🔹🔹🔹")
	fmt.Printf("🔹 Total: %d | ✅ Sucesso: %d | ❌ Falhas: %d\n", totalTests, passedTests, failedTests)

	if passedTests > 0 {
		fmt.Println("\n✅ Testes bem-sucedidos:")
		for _, result := range testResults {
			if result[:2] == "✅" {
				fmt.Println("   -", result)
			}
		}
	}

	if failedTests > 0 {
		fmt.Println("\n❌ Testes com falha:")
		for _, result := range testResults {
			if result[:2] == "❌" {
				fmt.Println("   -", result)
			}
		}
	}
}
