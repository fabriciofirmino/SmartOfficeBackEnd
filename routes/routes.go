package routes

import (
	"apiBackEnd/controllers"
	_ "apiBackEnd/docs" // Importação para Swagger funcionar

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes configura todas as rotas da API
func SetupRoutes(r *gin.Engine) {
	// Rota do Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Rotas de autenticação e informações iniciais
	r.POST("/login", controllers.Login)
	r.POST("/logout", controllers.Logout)
	r.GET("/api/version", controllers.GetAPIVersion)
	r.GET("/health", controllers.HealthCheck)

	// Grupo de rotas protegidas centralizado sob "/api"
	protected := r.Group("/api")
	protected.Use(controllers.AuthMiddleware())
	{
		// Endpoints gerais
		protected.GET("/clients", controllers.GetClients)
		protected.GET("/clients-table", controllers.GetClientsTable)
		protected.POST("/create-test", controllers.CreateTest)
		protected.GET("/details-error/:id_usuario", controllers.GetUserErrors)
		protected.GET("/dashboard", controllers.DashboardHandler)
		protected.POST("/renew", controllers.RenewAccount)
		protected.GET("/credits", controllers.GetCredits)
		protected.POST("/tools-table/add-screen", controllers.AddScreen)
		protected.POST("/tools-table/remove-screen", controllers.RemoveScreen)
		protected.PUT("/tools-table/edit/:id", controllers.EditUser)
		protected.POST("/trust-bonus", controllers.TrustBonusHandler)
		protected.POST("/renew-rollback", controllers.RenewRollbackHandler)
		protected.POST("/change-due-date", controllers.ChangeDueDateHandler)

		// Primeiro definir rotas fixas, depois rotas com parâmetros
		protected.GET("/users/deleted", controllers.ListDeletedUsersHandler)
		protected.GET("/regions/allowed", controllers.GetAllowedRegionsHandler)

		// Depois as rotas com parâmetros
		protected.PATCH("/users/:user_id/status", controllers.UpdateUserStatusHandler)
		protected.PATCH("/users/:user_id/region", controllers.ForceUserRegionHandler)
		protected.DELETE("/users/:user_id/session", controllers.KickUserSessionHandler)
		protected.PATCH("/users/:user_id/restore", controllers.RestoreUserHandler)
		protected.DELETE("/users/:user_id", controllers.SoftDeleteUserHandler)
	}
}
