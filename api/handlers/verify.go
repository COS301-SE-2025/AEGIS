package handlers

import (
	mfa "aegis-api/services_/auth/verification"
	"encoding/json"
	"fmt" // Add this import
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/pquerna/otp/totp"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type VerificationHandler struct {
	db         *sqlx.DB
	validator  *validator.Validate
	logger     *zap.Logger
	mfaService *mfa.MFAService
}

// VerifyAdminRequest represents the request to verify admin credentials
type VerifyAdminRequest struct {
	Password string `json:"password" validate:"required"`
}

// VerifyAdminResponse represents the response from admin verification
type VerifyAdminResponse struct {
	Valid       bool   `json:"valid"`
	Message     string `json:"message,omitempty"`
	RequiresMFA bool   `json:"requiresMFA,omitempty"`
}

// DeleteUserRequest represents the request to delete a user with authentication
type DeleteUserRequest struct {
	Password string `json:"password" validate:"required"`
	MFACode  string `json:"mfaCode,omitempty"`
}

// NewVerificationHandler is the constructor for VerificationHandler
func NewVerificationHandler(db *sqlx.DB, validator *validator.Validate, logger *zap.Logger, mfaService *mfa.MFAService) *VerificationHandler {
	return &VerificationHandler{
		db:         db,
		validator:  validator,
		logger:     logger,
		mfaService: mfaService,
	}
}

// VerifyAdminGin verifies admin password (simplified version)
func (h *VerificationHandler) VerifyAdminGin(c *gin.Context) {
	// Debug: Log all context values
	h.logger.Info("VerifyAdminGin: Starting verification",
		zap.Any("gin_keys", c.Keys))

	var req VerifyAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid JSON in request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		h.logger.Error("Validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}

	// Get current user from Gin context (set by AuthMiddleware)
	userIDInterface, exists := c.Get("userID")
	h.logger.Info("Getting userID from context",
		zap.Bool("exists", exists),
		zap.Any("value", userIDInterface))

	if !exists {
		h.logger.Error("userID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized - no user ID"})
		return
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		h.logger.Error("userID type assertion failed",
			zap.String("type", fmt.Sprintf("%T", userIDInterface)))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID format"})
		return
	}

	userRoleInterface, exists := c.Get("userRole")
	h.logger.Info("Getting role from context",
		zap.Bool("exists", exists),
		zap.Any("value", userRoleInterface))

	if !exists {
		h.logger.Error("role not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No role found"})
		return
	}

	userRole, ok := userRoleInterface.(string)
	if !ok {
		h.logger.Error("role type assertion failed",
			zap.String("type", fmt.Sprintf("%T", userRoleInterface)))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid role format"})
		return
	}

	if userRole != "DFIR Admin" {
		h.logger.Warn("Non-admin user attempted admin verification",
			zap.String("user_id", userID),
			zap.String("role", userRole))
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	h.logger.Info("Admin verification context OK",
		zap.String("user_id", userID),
		zap.String("role", userRole))

	// Get user from database
	var user struct {
		ID           string    `db:"id"`
		PasswordHash string    `db:"password_hash"`
		Role         string    `db:"role"`
		CreatedAt    time.Time `db:"created_at"`
	}

	query := `
        SELECT id, password_hash, role, created_at 
        FROM users 
        WHERE id = $1 AND role = 'DFIR Admin'
    `

	err := h.db.Get(&user, query, userID)
	if err != nil {
		h.logger.Error("Failed to get admin user", zap.Error(err), zap.String("user_id", userID))
		c.JSON(http.StatusNotFound, gin.H{"error": "Admin user not found"})
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		h.logger.Warn("Invalid admin password attempt", zap.String("user_id", userID))
		c.JSON(http.StatusOK, VerifyAdminResponse{
			Valid:   false,
			Message: "Invalid password",
		})
		return
	}

	// Log successful password verification
	h.logger.Info("Admin password verified", zap.String("user_id", userID))

	// Audit log
	_, err = h.db.Exec(`
        INSERT INTO audit_logs (user_id, action, resource_type, details, ip_address, user_agent)
        VALUES ($1, 'admin_password_verify', 'auth', $2, $3, $4)
    `, userID, `{"success": true}`,
		c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		h.logger.Error("Failed to log admin verification", zap.Error(err))
	}

	response := VerifyAdminResponse{
		Valid:   true,
		Message: "Password verified successfully",
	}

	c.JSON(http.StatusOK, response)
}

// DeleteUserWithAuth deletes a user with admin authentication and optional MFA
func (h *VerificationHandler) DeleteUserWithAuth(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	targetUserID := vars["userId"]

	if targetUserID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	var req DeleteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		http.Error(w, "Password is required", http.StatusBadRequest)
		return
	}

	// Get current admin user from JWT token
	adminUserID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	adminRole, ok := r.Context().Value("role").(string)
	if !ok || adminRole != "DFIR Admin" {
		http.Error(w, "Admin access required", http.StatusForbidden)
		return
	}

	// Prevent self-deletion
	if adminUserID == targetUserID {
		http.Error(w, "Cannot delete your own account", http.StatusForbidden)
		return
	}

	// Get admin user details
	var adminUser struct {
		ID           string `db:"id"`
		PasswordHash string `db:"password_hash"`
		MFAEnabled   bool   `db:"mfa_enabled"`
		MFASecret    string `db:"mfa_secret"`
		TenantID     string `db:"tenant_id"`
	}

	query := `
        SELECT id, password_hash, COALESCE(mfa_enabled, false) as mfa_enabled, 
               COALESCE(mfa_secret, '') as mfa_secret, tenant_id
        FROM users 
        WHERE id = $1 AND role = 'DFIR Admin'
    `

	err := h.db.Get(&adminUser, query, adminUserID)
	if err != nil {
		h.logger.Error("Failed to get admin user for deletion", zap.Error(err), zap.String("admin_id", adminUserID))
		http.Error(w, "Admin user not found", http.StatusNotFound)
		return
	}

	// Verify admin password
	err = bcrypt.CompareHashAndPassword([]byte(adminUser.PasswordHash), []byte(req.Password))
	if err != nil {
		h.logger.Warn("Invalid admin password for user deletion", zap.String("admin_id", adminUserID), zap.String("target_user_id", targetUserID))

		// Audit log failed attempt
		_, _ = h.db.Exec(`
            INSERT INTO audit_logs (user_id, action, resource_type, resource_id, details, ip_address, user_agent)
            VALUES ($1, 'user_delete_attempt_failed', 'user', $2, $3, $4, $5)
        `, adminUserID, targetUserID, `{"reason": "invalid_password"}`,
			getClientIP(r), r.UserAgent())

		http.Error(w, "Invalid admin password", http.StatusUnauthorized)
		return
	}

	// Verify MFA if enabled
	if adminUser.MFAEnabled {
		if req.MFACode == "" {
			http.Error(w, "MFA code is required", http.StatusBadRequest)
			return
		}

		if !h.verifyTOTP(adminUser.MFASecret, req.MFACode) {
			h.logger.Warn("Invalid MFA code for user deletion", zap.String("admin_id", adminUserID), zap.String("target_user_id", targetUserID))

			// Audit log failed MFA attempt
			_, _ = h.db.Exec(`
                INSERT INTO audit_logs (user_id, action, resource_type, resource_id, details, ip_address, user_agent)
                VALUES ($1, 'user_delete_attempt_failed', 'user', $2, $3, $4, $5)
            `, adminUserID, targetUserID, `{"reason": "invalid_mfa"}`,
				getClientIP(r), r.UserAgent())

			http.Error(w, "Invalid MFA code", http.StatusUnauthorized)
			return
		}
	}

	// Get target user details for audit logging
	var targetUser struct {
		ID       string `db:"id"`
		FullName string `db:"full_name"`
		Email    string `db:"email"`
		Role     string `db:"role"`
		TenantID string `db:"tenant_id"`
	}

	query = `
        SELECT id, full_name, email, role, tenant_id
        FROM users 
        WHERE id = $1
    `

	err = h.db.Get(&targetUser, query, targetUserID)
	if err != nil {
		h.logger.Error("Target user not found for deletion", zap.Error(err), zap.String("target_user_id", targetUserID))
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Verify same tenant (additional security check)
	if targetUser.TenantID != adminUser.TenantID {
		h.logger.Warn("Cross-tenant user deletion attempt", zap.String("admin_id", adminUserID), zap.String("target_user_id", targetUserID))
		http.Error(w, "Cannot delete user from different tenant", http.StatusForbidden)
		return
	}

	// Prevent deletion of other admins
	if targetUser.Role == "DFIR Admin" || targetUser.Role == "Tenant Admin" {
		h.logger.Warn("Attempt to delete admin user", zap.String("admin_id", adminUserID), zap.String("target_user_id", targetUserID), zap.String("target_role", targetUser.Role))
		http.Error(w, "Cannot delete admin users", http.StatusForbidden)
		return
	}

	// Start transaction for user deletion
	tx, err := h.db.Beginx()
	if err != nil {
		h.logger.Error("Failed to start transaction for user deletion", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Delete user's associated data first (maintain referential integrity)

	// Delete from case_users
	_, err = tx.Exec(`DELETE FROM case_users WHERE user_id = $1`, targetUserID)
	if err != nil {
		h.logger.Error("Failed to delete user case associations", zap.Error(err), zap.String("user_id", targetUserID))
		http.Error(w, "Failed to delete user associations", http.StatusInternalServerError)
		return
	}

	// Delete from team_members
	_, err = tx.Exec(`DELETE FROM team_members WHERE user_id = $1`, targetUserID)
	if err != nil {
		h.logger.Error("Failed to delete user team memberships", zap.Error(err), zap.String("user_id", targetUserID))
		http.Error(w, "Failed to delete user team memberships", http.StatusInternalServerError)
		return
	}

	// Update any cases where the user was the examiner (set to null or another admin)
	_, err = tx.Exec(`UPDATE cases SET examiner_id = NULL WHERE examiner_id = $1`, targetUserID)
	if err != nil {
		h.logger.Error("Failed to update cases with deleted examiner", zap.Error(err), zap.String("user_id", targetUserID))
		http.Error(w, "Failed to update case associations", http.StatusInternalServerError)
		return
	}

	// Delete user sessions
	_, err = tx.Exec(`DELETE FROM user_sessions WHERE user_id = $1`, targetUserID)
	if err != nil {
		h.logger.Error("Failed to delete user sessions", zap.Error(err), zap.String("user_id", targetUserID))
		// Continue anyway - not critical
	}

	// Finally, delete the user
	result, err := tx.Exec(`DELETE FROM users WHERE id = $1`, targetUserID)
	if err != nil {
		h.logger.Error("Failed to delete user", zap.Error(err), zap.String("user_id", targetUserID))
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		h.logger.Error("User deletion affected no rows", zap.String("user_id", targetUserID), zap.Int64("rows_affected", rowsAffected))
		http.Error(w, "User not found or already deleted", http.StatusNotFound)
		return
	}

	// Audit log successful deletion
	auditDetails := map[string]interface{}{
		"deleted_user": map[string]interface{}{
			"id":        targetUser.ID,
			"full_name": targetUser.FullName,
			"email":     targetUser.Email,
			"role":      targetUser.Role,
		},
		"auth_method": func() string {
			if adminUser.MFAEnabled {
				return "password_and_mfa"
			}
			return "password_only"
		}(),
		"success": true,
	}

	auditDetailsJSON, _ := json.Marshal(auditDetails)
	_, err = tx.Exec(`
        INSERT INTO audit_logs (user_id, action, resource_type, resource_id, details, ip_address, user_agent)
        VALUES ($1, 'user_deleted', 'user', $2, $3, $4, $5)
    `, adminUserID, targetUserID, string(auditDetailsJSON), getClientIP(r), r.UserAgent())

	if err != nil {
		h.logger.Error("Failed to log user deletion", zap.Error(err))
		// Continue anyway - deletion was successful
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		h.logger.Error("Failed to commit user deletion transaction", zap.Error(err))
		http.Error(w, "Failed to complete user deletion", http.StatusInternalServerError)
		return
	}

	h.logger.Info("User successfully deleted",
		zap.String("admin_id", adminUserID),
		zap.String("deleted_user_id", targetUserID),
		zap.String("deleted_user_email", targetUser.Email),
		zap.Bool("mfa_used", adminUser.MFAEnabled))

	// Return success response
	response := map[string]interface{}{
		"message": "User deleted successfully",
		"deleted_user": map[string]interface{}{
			"id":        targetUser.ID,
			"full_name": targetUser.FullName,
			"email":     targetUser.Email,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// VerifyAdminGinWithMFA verifies admin password and checks if MFA is required (Gin version)
func (h *VerificationHandler) VerifyAdminGinWithMFA(c *gin.Context) {
	var req VerifyAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}

	// Get current user from Gin context (set by AuthMiddleware)
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized - no user ID"})
		return
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID format"})
		return
	}

	userRoleInterface, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No role found"})
		return
	}

	userRole, ok := userRoleInterface.(string)
	if !ok || userRole != "DFIR Admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	// Get user from database
	var user struct {
		ID           string    `db:"id"`
		PasswordHash string    `db:"password_hash"`
		MFAEnabled   bool      `db:"mfa_enabled"`
		Role         string    `db:"role"`
		CreatedAt    time.Time `db:"created_at"`
	}

	query := `
        SELECT id, password_hash, COALESCE(mfa_enabled, false) as mfa_enabled, role, created_at 
        FROM users 
        WHERE id = $1 AND role = 'DFIR Admin'
    `

	err := h.db.Get(&user, query, userID)
	if err != nil {
		h.logger.Error("Failed to get admin user", zap.Error(err), zap.String("user_id", userID))
		c.JSON(http.StatusNotFound, gin.H{"error": "Admin user not found"})
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		h.logger.Warn("Invalid admin password attempt", zap.String("user_id", userID))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}

	// Log successful password verification
	h.logger.Info("Admin password verified", zap.String("user_id", userID), zap.Bool("mfa_required", user.MFAEnabled))

	// Audit log
	_, err = h.db.Exec(`
        INSERT INTO audit_logs (user_id, action, resource_type, details, ip_address, user_agent)
        VALUES ($1, 'admin_password_verify', 'auth', $2, $3, $4)
    `, userID, `{"success": true, "mfa_required": `+strconv.FormatBool(user.MFAEnabled)+`}`,
		c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		h.logger.Error("Failed to log admin verification", zap.Error(err))
	}

	response := VerifyAdminResponse{
		RequiresMFA: user.MFAEnabled,
		Message:     "Password verified successfully",
	}

	c.JSON(http.StatusOK, response)
}

// Helper function to get client IP address
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (in case of proxy)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// Update the verifyTOTP method
func (h *VerificationHandler) verifyTOTP(secret, code string) bool {
	// Use the MFA service instead of placeholder logic
	return totp.Validate(code, secret)
}

// Better yet, use the MFA service for complete verification
func (h *VerificationHandler) verifyMFA(userID uuid.UUID, code string) bool {
	valid, err := h.mfaService.VerifyTOTP(userID, code)
	if err != nil {
		h.logger.Error("MFA verification failed", zap.Error(err))
		return false
	}
	return valid
}
