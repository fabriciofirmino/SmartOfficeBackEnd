package controllers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// GetAPIVersion retorna a versão da API definida no .env
//
// @Summary Obter versão da API
// @Description Retorna a versão atual da API definida no arquivo .env
// @Tags Versão
// @Produce json
// @Success 200 {object} map[string]string "Versão da API"
// @Router /api/version [get]
func GetAPIVersion(c *gin.Context) {
	apiVersion := os.Getenv("API_VERSION")
	if apiVersion == "" {
		apiVersion = "v1.0.0" // Valor padrão caso a variável não esteja definida
	}

	c.JSON(http.StatusOK, gin.H{
		"api_version": apiVersion,
	})
}
