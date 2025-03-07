package routes

import (
	"apiBackEnd/controllers"
	"apiBackEnd/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	r.POST("/login", controllers.Login)

	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/verificar", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "Acesso permitido!"})
		})
	}
}
