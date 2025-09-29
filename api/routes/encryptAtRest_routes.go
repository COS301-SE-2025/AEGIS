package routes

import (
	"aegis-api/handlers"

	"github.com/gin-gonic/gin"
)

// RegisterEncryptionRoutes registers encryption/decryption endpoints
func RegisterEncryptionRoutes(rg *gin.RouterGroup, h *handlers.EncryptionHandler) {
	group := rg.Group("/encryption")
	{
		group.POST("/encrypt", h.Encrypt)
		group.POST("/decrypt", h.Decrypt)
		group.POST("/batch/encrypt", h.BatchEncrypt) // for multiple values
		group.POST("/batch/decrypt", h.BatchDecrypt) // for multiple values
	}
}