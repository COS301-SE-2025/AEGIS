package routes

import (
	//_ "aegis-api/docs"
	//"aegis-api/handlers"
	//"aegis-api/middleware"
	//"fmt"
	"github.com/gin-gonic/gin"
	//swaggerFiles "github.com/swaggo/files"
		"aegis-api/handlers"
	//ginSwagger "github.com/swaggo/gin-swagger"
	"aegis-api/middleware"
	"aegis-api/services/case_assign"
	database "aegis-api/db"
	"aegis-api/services/case_creation"
)







// mock service structs
func SetUpRouter(h *handlers.Handler) *gin.Engine {
	  r := gin.Default()


    api := r.Group("/api/v1")
    admin := api.Group("/admin")
    {
        admin.POST("/users", h.AdminService.RegisterUser)
        admin.GET("/users", h.AdminService.ListUsers)
        admin.PUT("/users/:user_id", h.AdminService.UpdateUserRole)
        admin.DELETE("/users/:user_id", h.AdminService.DeleteUser)
    }

		// 	admin.POST("/users", middleware.RequireRole("Admin"), h.AdminService.RegisterUser)
		// admin.GET("/users", middleware.RequireRole("Admin"), h.AdminService.ListUsers)
		// admin.PUT("/users/:user_id", middleware.RequireRole("Admin"), h.AdminService.UpdateUserRole)
		// admin.DELETE("/users/:user_id", middleware.RequireRole("Admin"), h.AdminService.DeleteUser)
	auth := api.Group("/auth")
	{
		auth.POST("/login", h.AuthService.LoginHandler)

		//auth.POST("/logout", middleware.AuthMiddleware(), h.AuthService.LogoutHandler)
		auth.POST("/password-reset", h.AuthService.ResetPasswordHandler)
		// add logout, reset-password etc.
	}
caseRepo := case_creation.NewGormCaseRepository(database.DB)
caseService := case_creation.NewCaseService(caseRepo)

caseAssignmentRepo := case_assign.NewGormCaseAssignmentRepo(database.DB) // implement repo for case assignment
caseAssignmentService := case_assign.NewCaseAssignmentService(caseAssignmentRepo)

caseHandler := handlers.NewCaseHandler(caseService, caseAssignmentService)


	  cases := api.Group("/cases", middleware.AuthMiddleware())
    {
        // POST /api/v1/cases → CreateCase handler
        cases.POST("", h.CaseService.CreateCase)
		cases.POST("/:case_id/collaborators", middleware.AuthMiddleware(), caseHandler.CreateCollaborator)

    }

		


	return r















	// router.Use(gin.Logger()) //middleware for logging requests
	// fmt.Println()
	// //router.Use(gin.Recovery())              //middleware for recovering from panics

	// h := handlers.NewHandler(
	// 	handlers.MockAdminService{},
	// 	handlers.MockAuthService{},
	// 	handlers.MockCaseService{},
	// 	handlers.MockEvidenceService{},
	// 	handlers.MockUserService{},
	// )

	// // Debug: Print all routes
	// router.Use(func(c *gin.Context) {
	// 	fmt.Printf("Requested URL: %s\n", c.Request.URL.Path)
	// 	c.Next()
	// })

	// router.GET("/ping", func(ctx *gin.Context) {
	// 	ctx.JSON(200, gin.H{
	// 		"message": "pong",
	// 	})
	// })

	// //routes group
	// api := router.Group("/api/v1")

	// //auth
	// auth := api.Group("/auth")
	// {
	// 	// activate-account
	// 	//auth.POST("/activate-account", authService.activateAccount)

	// 	// login
	// 	auth.POST("/login", h.AuthService.Login)

	// 	// logout
	// 	auth.POST("/logout", h.AuthService.Logout)

	// 	// password-reset
	// 	auth.POST("/password-reset", h.AuthService.ResetPassword)

	// 	// auth.POST("/refresh-token", authService.refreshToken)
	// }

	// protected := api.Group("")
	// protected.Use(middleware.AuthMiddleware())
	// {
	// 	//admin (protected)
	// 	admin := protected.Group("/admin")
	// 	admin.Use(middleware.RequireRole("Admin"))
	// 	{
	// 		admin.POST("/users", h.AdminService.RegisterUser)
	// 		admin.GET("/users", h.AdminService.ListUsers)
	// 		admin.GET("/users/:user_id", h.AdminService.GetUserActivity)
	// 		admin.PUT("/users/:user_id", h.AdminService.UpdateUserRole)
	// 		admin.DELETE("/users/:user_id", h.AdminService.DeleteUser)

	// 		// roles
	// 		admin.GET("/roles", h.AdminService.GetRoles)
	// 		//admin.GET("/audit-logs", h.AdminService.getAuditLogs)
	// 		//dashboard stuff
	// 	}

	// 	user := protected.Group("/user")
	// 	{
	// 		user.GET("/me", h.UserService.GetUserInfo)
	// 		user.PUT("/me", h.UserService.UpdateUserInfo)
	// 	}

	// 	//cases
	// 	cases := protected.Group("/cases")
	// 	{
	// 		cases.GET("", h.CaseService.GetCases)                                     //support for pagination, filtering, etc.
	// 		cases.POST("", middleware.RequireRole("Admin"), h.CaseService.CreateCase) //admin only --adjust***

	// 		//case-specific routes
	// 		singleCase := cases.Group("/:id")
	// 		{
	// 			// ?case_id
	// 			singleCase.GET("", h.CaseService.GetCase)                                     //get a specific case by id
	// 			singleCase.PUT("", middleware.RequireRole("Admin"), h.CaseService.UpdateCase) //admin only

	// 			//cases.DELETE("", h.CaseService.DeleteCase)

	// 			singleCase.POST("/collaborators", middleware.RequireRole("Admin"), h.CaseService.CreateCollaborator)
	// 			singleCase.DELETE("/collaborators/:user", middleware.RequireRole("Admin"), h.CaseService.RemoveCollaborator)

	// 			//collaborators
	// 			singleCase.GET("/collaborators", h.CaseService.GetCollaborators)

	// 			//singleCase.GET("/timeline", h.CaseService.GetTimeline) later

	// 			evidence := singleCase.Group("/evidence")
	// 			{
	// 				evidence.GET("", h.EvidenceService.GetEvidence) //evidence under a specific case
	// 				evidence.POST("", h.EvidenceService.UploadEvidence)

	// 				//evidence specific to a single case
	// 				evidenceItem := evidence.Group("/:e_id")
	// 				{
	// 					evidenceItem.GET("", h.EvidenceService.GetEvidenceItem)
	// 					evidenceItem.GET("/preview", h.EvidenceService.PreviewEvidence)
	// 					//evidenceItem.POST("/annotations", h.EvidenceService.AddAnnotation)
	// 					//evidenceItem.GET("/annotations", h.EvidenceService.GetAnnotations)
	// 				}
	// 			}
	// 		}
	// 	}

	// }

	// router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// return router
}
