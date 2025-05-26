package routes

import (
	"aegis-core/handlers"
	"fmt"
	"github.com/gin-gonic/gin"
)

// mock service structs
func SetUpRouter() *gin.Engine {
	router := gin.Default()
	h := handlers.NewHandler()

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

	//admin
	admin := api.Group("/admin")
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

	//cases
	cases := api.Group("/cases")
	{
		cases.GET("/", h.CaseService.GetCases) //support for pagination, filtering, etc.
		cases.POST("/", h.CaseService.CreateCase)

		//case-specific routes
		singleCase := cases.Group("/:id")
		{
			// ?case_id
			singleCase.GET("/", h.CaseService.GetCase)
			singleCase.PUT("/", h.CaseService.UpdateCase)
			//singleCase.DELETE("/", h.CaseService.DeleteCase)
			singleCase.POST("/assign", h.CaseService.AssignCase) //admin only? create collaborator

			//collaborators
			singleCase.GET("/collaborators", h.CaseService.GetCollaborators)

			singleCase.DELETE("/collaborators/:user", h.CaseService.RemoveCollaborator)

			//singleCase.GET("/timeline", h.CaseService.GetTimeline) later?

			evidence := singleCase.Group("/evidence")
			{
				evidence.GET("/", h.EvidenceService.GetEvidence) //evidence under a specific case
				evidence.POST("/", h.EvidenceService.UploadEvidence)
				//evidence.GET("/user") get evidence uploaded by a specific user

				//evidence specific to a single case
				evidenceItem := evidence.Group("/:e_id")
				{
					evidenceItem.GET("/", h.EvidenceService.GetEvidenceItem)
					evidenceItem.GET("/preview", h.EvidenceService.PreviewEvidence)
					//evidenceItem.POST("/annotations", h.EvidenceService.AddAnnotation)
					//evidenceItem.GET("/annotations", h.EvidenceService.GetAnnotations)
				}
			}
		}
	}

	user := api.Group("/user")
	{
		user.GET("/me", h.UserService.GetUserInfo)
		user.PUT("/me", h.UserService.UpdateUserInfo)
		user.GET("/me/cases", h.UserService.GetUserCases)
	}

	return router
}
