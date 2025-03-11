package tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetClientsSuccess(t *testing.T) {
	fmt.Println("🚀 Testando GET /api/clients")

	// 🔹 Garante que o TestToken esteja definido antes de continuar
	EnsureAuthToken(t)

	req, _ := http.NewRequest("GET", "/api/clients", nil)
	req.Header.Set("Authorization", "Bearer "+TestToken)

	w := httptest.NewRecorder()
	SetupServer().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
