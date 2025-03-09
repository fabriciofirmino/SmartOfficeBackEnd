package routes

import (
	"apiBackEnd/controllers"
	_ "apiBackEnd/docs" // 🔥 Importação para Swagger funcionar

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"     // ✅ Correto
	ginSwagger "github.com/swaggo/gin-swagger" // ✅ Correto
)

// SetupRoutes configura todas as rotas da API
func SetupRoutes(r *gin.Engine) {
	// 📌 Rota do Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler)) // 🔥 Rota do Swagger

	// 📌 Rota de autenticação
	r.POST("/login", controllers.Login)
	r.POST("/logout", controllers.Logout)

	// 📌 Grupo de rotas protegidas
	protected := r.Group("/api")
	protected.Use(controllers.AuthMiddleware()) // ✅ Certifique-se que esta função existe

	{
		// 📌 Rota de listagem de clientes
		protected.GET("/clients", controllers.GetClients)

		// 📌 Rota de criação de testes IPTV
		protected.POST("/create-test", controllers.CreateTest)

		// 📌 Rota de renovação de conta
		protected.POST("/renew", controllers.RenewAccount)
		protected.GET("/credits", controllers.GetCredits)

	}
}
