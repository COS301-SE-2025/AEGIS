package x3dh

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type RotateSPKRequest struct {
	UserID    string  `json:"user_id"`
	NewSPK    string  `json:"new_spk"`
	Signature string  `json:"signature"`
	ExpiresAt *string `json:"expires_at,omitempty"` // optional ISO date
}

func registerSPKRotationHandler(rg *gin.RouterGroup, bundleService *BundleService) {
	rg.POST("/rotate-spk", func(c *gin.Context) {
		var req RotateSPKRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		// Convert expires_at
		var expiresAt *time.Time
		if req.ExpiresAt != nil {
			t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expires_at format"})
				return
			}
			expiresAt = &t
		}

		err := bundleService.RotateSPK(c.Request.Context(), req.UserID, req.NewSPK, req.Signature, expiresAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusOK)
	})
}

// RegisterX3DHHandlers registers X3DH-related routes
func RegisterX3DHHandlers(rg *gin.RouterGroup, bundleService *BundleService) {
	// GET bundle (already exists)
	rg.GET("/bundle/:user_id", func(c *gin.Context) {
		userID := c.Param("user_id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing user_id"})
			return
		}

		bundle, err := bundleService.GetBundle(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, bundle)
	})

	// POST /x3dh/register-bundle
	rg.POST("/register-bundle", func(c *gin.Context) {
		var req RegisterBundleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := bundleService.StoreBundle(c.Request.Context(), req); err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				c.JSON(http.StatusConflict, gin.H{"error": "bundle already exists"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"status": "bundle stored successfully"})
	})

	// ✅ GET /x3dh/opk-count/:user_id
	rg.GET("/opk-count/:user_id", func(c *gin.Context) {
		userID := c.Param("user_id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing user_id"})
			return
		}

		count, err := bundleService.CountAvailableOPKs(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"user_id": userID, "available_opks": count})
	})

	rg.POST("/refill-opks", func(c *gin.Context) {
		var req RefillOPKRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := bundleService.RefillOPKs(c.Request.Context(), req.UserID, req.OPKs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "refilled OPKs successfully"})
	})

}
