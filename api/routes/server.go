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

func SetUpRouter(h *handlers.Handler) *gin.Engine {
	router := gin.Default()
	//h := handlers.NewHandler()

	router.Use(gin.Logger()) //middleware for logging requests
	fmt.Println()
	//router.Use(gin.Recovery())              //middleware for recovering from panics

	// Debug: Print all routes --remove later
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
		// login
		auth.POST("/login", h.AuthHandler.Login)

		// logout
		auth.DELETE("/logout", middleware.AuthMiddleware(), h.AuthHandler.Logout)

		// password-reset
		//auth.POST("/reset-password", h.AuthHandler.ResetPassword)
		//auth.POST("/reset-password/request", h.AuthHandler.RequestPasswordReset)

		// auth.POST("/refresh-token", authService.refreshToken)
	}

	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		//admin (protected)
		admin := protected.Group("/admin")
		admin.Use(middleware.RequireRole("Admin"))
		{
			users := admin.Group("/users")
			{
				users.GET("", h.AdminHandler.ListUsers)
				users.POST("", h.AdminHandler.RegisterUser)

				singleUser := users.Group("/:user_id")
				{
					singleUser.GET("", h.UserHandler.GetProfile)            //get user profile
					singleUser.PUT("/profile", h.UserHandler.UpdateProfile) //update user profile (name,email)

					singleUser.GET("/roles", h.UserHandler.GetUserRoles)
					singleUser.PUT("", h.AdminHandler.UpdateUserRole) //update user role
					singleUser.DELETE("", h.AdminHandler.DeleteUser)  //delete user

					singleUser.GET("/cases", h.CaseHandler.ListCasesByUserID)                            //get cases by user id
					singleUser.GET("/evidence", h.EvidenceHandler.ListEvidenceByUserID)                  //get evidence uploaded by user id
					singleUser.GET("/evidence/:evidence_id", h.EvidenceHandler.DownloadEvidenceByUserID) //download evidence by user id
				}
			}
			// roles
			//admin.GET("/roles", h.AdminHandler.GetRoles)
			//admin.GET("/audit-logs", h.AdminHandler.getAuditLogs)
			//dashboard stuff
		}

		// user self-service routes
		user := protected.Group("/users")
		{
			//profile
			user.GET("", h.UserHandler.GetProfile)
			user.PUT("", h.UserHandler.UpdateProfile) //update name/email

			//cases
			user.GET("/cases", h.CaseHandler.ListCasesByUserID)
			user.GET("/evidence", h.EvidenceHandler.ListEvidenceByUserID)
			user.GET("/evidence/:evidence_id", h.EvidenceHandler.DownloadEvidenceByUserID)

			//
			user.GET("/roles", h.UserHandler.GetUserRoles)
		}

		//cases
		cases := protected.Group("/cases")
		{
			cases.POST("", middleware.RequireRole("Admin"), h.CaseHandler.CreateCase)
			cases.GET("", middleware.RequireRole("Admin"), h.CaseHandler.ListAllCases)             //no filtering
			cases.GET("/filter", middleware.RequireRole("Admin"), h.CaseHandler.ListFilteredCases) //filter cases by status, type, etc.

			//case-specific routes
			singleCase := cases.Group("/:case_id")
			{
				// ?case_id
				//singleCase.GET("", h.CaseHandler.GetCaseByID) //get a specific case by id
				singleCase.PUT("/status", middleware.RequireRole("Admin"), h.CaseHandler.UpdateCaseStatus) //admin only
				//cases.DELETE("", h.CaseHandler.DeleteCase)

				//collaborators
				singleCase.POST("/collaborators", middleware.RequireRole("Admin"), h.CaseHandler.CreateCollaborator)
				singleCase.GET("/collaborators", h.CaseHandler.ListCollaborators)
				singleCase.DELETE("/collaborators/:user_id", middleware.RequireRole("Admin"), h.CaseHandler.RemoveCollaborator)

				//singleCase.GET("/timeline", h.CaseHandler.GetTimeline) later

				evidence := singleCase.Group("/evidence")
				{
					//evidence.POST("", h.EvidenceHandler.UploadEvidence) //UNDER REVIEW
					evidence.GET("", h.EvidenceHandler.ListEvidenceByCaseID)

					//evidence specific to a single case
					evidenceItem := evidence.Group("/:evidence_id")
					{
						evidenceItem.GET("", h.EvidenceHandler.GetEvidenceByID)
						evidenceItem.GET("/metadata", h.EvidenceHandler.GetEvidenceMetadata)
						evidenceItem.DELETE("", middleware.RequireRole("Admin"), h.EvidenceHandler.DeleteEvidenceByID)

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
