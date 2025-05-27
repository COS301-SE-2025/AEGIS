package routes

import (
	"aegis-core/handlers"
	"aegis-core/middleware"
	"fmt"
	"github.com/gin-gonic/gin"
)

// mock service structs
func SetUpRouter() *gin.Engine {
	router := gin.Default()
	//h := handlers.NewHandler()

	router.Use(gin.Logger()) //middleware for logging requests
	fmt.Println()
	//router.Use(gin.Recovery())              //middleware for recovering from panics

	h := handlers.NewHandler(
		handlers.MockAdminService{},
		handlers.MockAuthService{},
		handlers.MockCaseService{},
		handlers.MockEvidenceService{},
		handlers.MockUserService{},
	)

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
		auth.POST("/login", h.AuthService.Login)

		// logout
		auth.POST("/logout", h.AuthService.Logout)

		// password-reset
		auth.POST("/password-reset", h.AuthService.ResetPassword)

		// auth.POST("/refresh-token", authService.refreshToken)
	}

	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		//admin (protected)
		admin := protected.Group("/admin")
		admin.Use(middleware.RequireRole("Admin"))
		{
			admin.POST("/users", h.AdminService.RegisterUser)
			admin.GET("/users", h.AdminService.ListUsers)
			admin.GET("/users/:user_id", h.AdminService.GetUserActivity)
			admin.PUT("/users/:user_id", h.AdminService.UpdateUserRole)
			admin.DELETE("/users/:user_id", h.AdminService.DeleteUser)

			// roles
			admin.GET("/roles", h.AdminService.GetRoles)
			//admin.GET("/audit-logs", h.AdminService.getAuditLogs)
			//dashboard stuff
		}

		user := protected.Group("/user")
		{
			user.GET("/me", h.UserService.GetUserInfo)
			user.PUT("/me", h.UserService.UpdateUserInfo)
		}

		//cases
		cases := protected.Group("/cases")
		{
			cases.GET("", h.CaseService.GetCases)                                     //support for pagination, filtering, etc.
			cases.POST("", middleware.RequireRole("Admin"), h.CaseService.CreateCase) //admin only --adjust***

			//case-specific routes
			singleCase := cases.Group("/:id")
			{
				// ?case_id
				singleCase.GET("", h.CaseService.GetCase)                                     //get a specific case by id
				singleCase.PUT("", middleware.RequireRole("Admin"), h.CaseService.UpdateCase) //admin only

				//cases.DELETE("", h.CaseService.DeleteCase)

				singleCase.POST("/collaborators", middleware.RequireRole("Admin"), h.CaseService.CreateCollaborator)
				singleCase.DELETE("/collaborators/:user", middleware.RequireRole("Admin"), h.CaseService.RemoveCollaborator)

				//collaborators
				singleCase.GET("/collaborators", h.CaseService.GetCollaborators)

				//singleCase.GET("/timeline", h.CaseService.GetTimeline) later

				evidence := singleCase.Group("/evidence")
				{
					evidence.GET("", h.EvidenceService.GetEvidence) //evidence under a specific case
					evidence.POST("", h.EvidenceService.UploadEvidence)

					//evidence specific to a single case
					evidenceItem := evidence.Group("/:e_id")
					{
						evidenceItem.GET("", h.EvidenceService.GetEvidenceItem)
						evidenceItem.GET("/preview", h.EvidenceService.PreviewEvidence)
						//evidenceItem.POST("/annotations", h.EvidenceService.AddAnnotation)
						//evidenceItem.GET("/annotations", h.EvidenceService.GetAnnotations)
					}
				}
			}
		}

	}
	return router
}
