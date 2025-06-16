package routes

import (
	//_ "aegis-api/docs"
	"aegis-api/handlers"
	"aegis-api/middleware"
	"fmt"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// mock service structs
func SetUpRouter(h *handlers.Handler) *gin.Engine {
	router := gin.Default()
	//h := handlers.NewHandler()

	router.Use(gin.Logger()) //middleware for logging requests
	fmt.Println()
	//router.Use(gin.Recovery())              //middleware for recovering from panics

	// Debug: Print all routes
	router.Use(func(c *gin.Context) {
		fmt.Printf("Requested URL: %s\n", c.Request.URL.Path)
		c.Next()
	})

	router.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "pong",
		})
	})

	//routes group
	api := router.Group("/api/v1")

	//auth
	auth := api.Group("/auth")
	{
		// activate-account
		//auth.POST("/activate-account", authService.activateAccount)

		// login
		auth.POST("/login", h.AuthHandler.Login)

		// logout
		auth.POST("/logout", h.AuthHandler.Logout)

		// password-reset
		auth.POST("/password-reset", h.AuthHandler.ResetPassword)

		// auth.POST("/refresh-token", authService.refreshToken)
	}

	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		//admin (protected)
		admin := protected.Group("/admin")
		admin.Use(middleware.RequireRole("Admin"))
		{
			admin.POST("/users", h.AdminHandler.RegisterUser)
			admin.GET("/users", h.AdminHandler.ListUsers)
			admin.GET("/users/:user_id", h.AdminHandler.GetUserActivity)
			admin.PUT("/users/:user_id", h.AdminHandler.UpdateUserRole)
			admin.DELETE("/users/:user_id", h.AdminHandler.DeleteUser)

			// roles
			admin.GET("/roles", h.AdminHandler.GetRoles)
			//admin.GET("/audit-logs", h.AdminHandler.getAuditLogs)
			//dashboard stuff
		}

		user := protected.Group("/user")
		{
			user.GET("/me", h.UserHandler.GetUserInfo)
			user.PUT("/me", h.UserHandler.UpdateUserInfo)
		}

		//cases
		cases := protected.Group("/cases")
		{
			cases.GET("", h.CaseHandler.GetCases)                                     //support for pagination, filtering, etc.
			cases.POST("", middleware.RequireRole("Admin"), h.CaseHandler.CreateCase) //admin only --adjust***

			//case-specific routes
			singleCase := cases.Group("/:id")
			{
				// ?case_id
				singleCase.GET("", h.CaseHandler.GetCase)                                     //get a specific case by id
				singleCase.PUT("", middleware.RequireRole("Admin"), h.CaseHandler.UpdateCase) //admin only

				//cases.DELETE("", h.CaseHandler.DeleteCase)

				singleCase.POST("/collaborators", middleware.RequireRole("Admin"), h.CaseHandler.CreateCollaborator)
				singleCase.DELETE("/collaborators/:user", middleware.RequireRole("Admin"), h.CaseHandler.RemoveCollaborator)

				//collaborators
				singleCase.GET("/collaborators", h.CaseHandler.GetCollaborators)

				//singleCase.GET("/timeline", h.CaseHandler.GetTimeline) later

				evidence := singleCase.Group("/evidence")
				{
					evidence.GET("", h.EvidenceHandler.GetEvidence) //evidence under a specific case
					evidence.POST("", h.EvidenceHandler.UploadEvidence)

					//evidence specific to a single case
					evidenceItem := evidence.Group("/:e_id")
					{
						evidenceItem.GET("", h.EvidenceHandler.GetEvidenceItem)
						evidenceItem.GET("/preview", h.EvidenceHandler.PreviewEvidence)
						//evidenceItem.POST("/annotations", h.EvidenceHandler.AddAnnotation)
						//evidenceItem.GET("/annotations", h.EvidenceHandler.GetAnnotations)
					}
				}
			}
		}

	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return router
}
