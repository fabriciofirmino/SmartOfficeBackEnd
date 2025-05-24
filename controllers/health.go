package controllers

import (
	"apiBackEnd/config"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthCheck godoc
// @Summary Health Check
// @Description Check the health of the service, including database and Redis connections
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Exemplo: {\"status\": \"OK\", \"database\": \"OK\", \"redis\": \"OK\"}"
// @Failure 503 {object} map[string]string "Exemplo: {\"status\": \"Error\", \"database\": \"Error: connection failed\", \"redis\": \"OK\"}"
// @Router /health [get]
func HealthCheck(c *gin.Context) {
	dbStatus := "OK"
	redisStatus := "OK"
	overallStatus := "OK"

	// Verificar conexão com o banco de dados
	if config.DB == nil {
		dbStatus = "Error: not initialized"
		overallStatus = "Error"
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := config.DB.PingContext(ctx); err != nil {
			dbStatus = "Error: " + err.Error()
			overallStatus = "Error"
		}
	}

	// Verificar conexão com o Redis
	if config.RedisClient == nil {
		redisStatus = "Error: not initialized"
		overallStatus = "Error"
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if _, err := config.RedisClient.Ping(ctx).Result(); err != nil {
			redisStatus = "Error: " + err.Error()
			overallStatus = "Error"
		}
	}

	response := gin.H{
		"status":   overallStatus,
		"database": dbStatus,
		"redis":    redisStatus,
	}

	if overallStatus == "OK" {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusServiceUnavailable, response)
	}
}
