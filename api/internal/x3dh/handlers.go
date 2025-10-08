package x3dh

import (
	"context"
	"log"
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

// BundleServiceInterface defines the contract for bundle operations
type BundleServiceInterface interface {
	GetBundle(ctx context.Context, userID string) (*BundleResponse, error)
	StoreBundle(ctx context.Context, req RegisterBundleRequest) error
	RefillOPKs(ctx context.Context, userID string, opks []OneTimePreKeyUpload) error
	RotateSPK(ctx context.Context, userID, newSPK, signature string, expiresAt *time.Time) error
	CountAvailableOPKs(ctx context.Context, userID string) (int, error)
}

// Handler struct to hold dependencies
type Handler struct {
	BundleService BundleServiceInterface
}

// NewHandler creates a new handler instance
func NewHandler(bundleService BundleServiceInterface) *Handler {
	return &Handler{
		BundleService: bundleService,
	}
}

// GetBundleHandler - actual handler from your package
func (h *Handler) GetBundleHandler(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user_id"})
		return
	}

	bundle, err := h.BundleService.GetBundle(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bundle)
}

// RegisterBundleHandler - actual handler from your package
func (h *Handler) RegisterBundleHandler(c *gin.Context) {
	var req RegisterBundleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[x3dh] register-bundle userId=%s opks=%d", req.UserID, len(req.OneTimePreKeys))
	if err := h.BundleService.StoreBundle(c.Request.Context(), req); err != nil {
		log.Printf("[x3dh] StoreBundle failed for user %s: %v", req.UserID, err)
		if strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, gin.H{"error": "bundle already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "bundle stored successfully"})
}

// CountOPKsHandler - actual handler from your package
func (h *Handler) CountOPKsHandler(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user_id"})
		return
	}

	count, err := h.BundleService.CountAvailableOPKs(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_id": userID, "available_opks": count})
}

// RefillOPKsHandler - actual handler from your package
func (h *Handler) RefillOPKsHandler(c *gin.Context) {
	var req RefillOPKRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "detail": err.Error()})
		return
	}
	if err := h.BundleService.RefillOPKs(c.Request.Context(), req.UserID, req.OPKs); err != nil {
		log.Printf("[x3dh] refill-opks error user=%s err=%v", req.UserID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "refill_failed", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "refilled OPKs successfully"})
}

// RotateSPKHandler - actual handler from your package
func (h *Handler) RotateSPKHandler(c *gin.Context) {
	var req RotateSPKRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	var expTime *time.Time
	if req.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expires_at format"})
			return
		}
		expTime = &t
	}

	err := h.BundleService.RotateSPK(c.Request.Context(), req.UserID, req.NewSPK, req.Signature, expTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// RegisterX3DHHandlers registers X3DH-related routes
func RegisterX3DHHandlers(rg *gin.RouterGroup, bundleService BundleServiceInterface) {
	handler := NewHandler(bundleService)

	// log every request under /x3dh
	rg.Use(func(c *gin.Context) {
		log.Printf("[x3dh] %s %s", c.Request.Method, c.FullPath())
		c.Next()
	})

	// Register routes with actual handler methods
	rg.GET("/bundle/:user_id", handler.GetBundleHandler)
	rg.POST("/register-bundle", handler.RegisterBundleHandler)
	rg.GET("/opk-count/:user_id", handler.CountOPKsHandler)
	rg.POST("/refill-opks", handler.RefillOPKsHandler)
	rg.POST("/rotate-spk", handler.RotateSPKHandler)
}
