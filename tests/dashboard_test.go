package tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDashboardSuccess(t *testing.T) {
	testName := "TestDashboardSuccess" // Nome do teste

	req, _ := http.NewRequest("GET", "/api/dashboard", nil)
	fmt.Println("🚀 Testando GET /api/dashboard") // 🔹 Log para debug
	req.Header.Set("Authorization", "Bearer "+TestToken)

	w := httptest.NewRecorder()
	SetupServer().ServeHTTP(w, req) // 🔹 Agora usa a função correta

	if assert.Equal(t, http.StatusOK, w.Code) {
		recordTestResult(testName, true)
	} else {
		recordTestResult(testName, false)
	}
}
