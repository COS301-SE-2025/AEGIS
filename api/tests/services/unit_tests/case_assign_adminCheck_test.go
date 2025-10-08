package unit_tests

import (
	"aegis-api/services_/case/case_assign"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupGinContext creates a test Gin context
func setupGinContext() *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c
}

// TestContextAdminChecker_IsAdminFromContext tests the IsAdminFromContext method
func TestContextAdminChecker_IsAdminFromContext(t *testing.T) {
	t.Run("returns true for DFIR Admin role", func(t *testing.T) {
		checker := case_assign.NewContextAdminChecker()
		ctx := setupGinContext()

		ctx.Set("userRole", "DFIR Admin")

		isAdmin, err := checker.IsAdminFromContext(ctx)

		assert.NoError(t, err)
		assert.True(t, isAdmin, "Expected DFIR Admin to be identified as admin")
	})

	t.Run("returns false for non-admin roles", func(t *testing.T) {
		checker := case_assign.NewContextAdminChecker()

		testCases := []struct {
			name string
			role string
		}{
			{"Analyst role", "Analyst"},
			{"User role", "User"},
			{"Manager role", "Manager"},
			{"Investigator role", "Investigator"},
			{"empty string", ""},
			{"random role", "SomeOtherRole"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				ctx := setupGinContext()
				ctx.Set("userRole", tc.role)

				isAdmin, err := checker.IsAdminFromContext(ctx)

				assert.NoError(t, err)
				assert.False(t, isAdmin, "Expected %s to not be identified as admin", tc.role)
			})
		}
	})

	t.Run("returns error when userRole is missing from context", func(t *testing.T) {
		checker := case_assign.NewContextAdminChecker()
		ctx := setupGinContext()

		// Don't set userRole in context

		isAdmin, err := checker.IsAdminFromContext(ctx)

		assert.Error(t, err)
		assert.False(t, isAdmin)
		assert.Equal(t, "missing role in context", err.Error())
	})

	t.Run("returns error when userRole is not a string", func(t *testing.T) {
		checker := case_assign.NewContextAdminChecker()

		testCases := []struct {
			name  string
			value interface{}
		}{
			{"integer", 123},
			{"boolean", true},
			{"slice", []string{"DFIR Admin"}},
			{"map", map[string]string{"role": "DFIR Admin"}},
			{"nil", nil},
			{"struct", struct{ Role string }{Role: "DFIR Admin"}},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				ctx := setupGinContext()
				ctx.Set("userRole", tc.value)

				isAdmin, err := checker.IsAdminFromContext(ctx)

				assert.Error(t, err)
				assert.False(t, isAdmin)
				assert.Equal(t, "invalid role type in context", err.Error())
			})
		}
	})

	t.Run("is case sensitive for DFIR Admin", func(t *testing.T) {
		checker := case_assign.NewContextAdminChecker()

		testCases := []string{
			"dfir admin",
			"DFIR ADMIN",
			"dfir Admin",
			"Dfir Admin",
			"DfIr AdMiN",
			" DFIR Admin",
			"DFIR Admin ",
			" DFIR Admin ",
		}

		for _, role := range testCases {
			t.Run("role: "+role, func(t *testing.T) {
				ctx := setupGinContext()
				ctx.Set("userRole", role)

				isAdmin, err := checker.IsAdminFromContext(ctx)

				assert.NoError(t, err)
				assert.False(t, isAdmin, "Expected '%s' to not match exact 'DFIR Admin'", role)
			})
		}
	})

	t.Run("handles empty context gracefully", func(t *testing.T) {
		checker := case_assign.NewContextAdminChecker()
		ctx := setupGinContext()

		// Context is empty, no keys set

		isAdmin, err := checker.IsAdminFromContext(ctx)

		assert.Error(t, err)
		assert.False(t, isAdmin)
		assert.Equal(t, "missing role in context", err.Error())
	})

	t.Run("multiple calls with same context", func(t *testing.T) {
		checker := case_assign.NewContextAdminChecker()
		ctx := setupGinContext()
		ctx.Set("userRole", "DFIR Admin")

		// First call
		isAdmin1, err1 := checker.IsAdminFromContext(ctx)
		assert.NoError(t, err1)
		assert.True(t, isAdmin1)

		// Second call should return same result
		isAdmin2, err2 := checker.IsAdminFromContext(ctx)
		assert.NoError(t, err2)
		assert.True(t, isAdmin2)
	})

	t.Run("context with multiple keys", func(t *testing.T) {
		checker := case_assign.NewContextAdminChecker()
		ctx := setupGinContext()

		// Set multiple keys in context
		ctx.Set("userID", "123")
		ctx.Set("tenantID", "456")
		ctx.Set("userRole", "DFIR Admin")
		ctx.Set("userName", "John Doe")

		isAdmin, err := checker.IsAdminFromContext(ctx)

		assert.NoError(t, err)
		assert.True(t, isAdmin)
	})

	t.Run("overwriting role in context", func(t *testing.T) {
		checker := case_assign.NewContextAdminChecker()
		ctx := setupGinContext()

		// First set as admin
		ctx.Set("userRole", "DFIR Admin")
		isAdmin1, err1 := checker.IsAdminFromContext(ctx)
		assert.NoError(t, err1)
		assert.True(t, isAdmin1)

		// Then change to non-admin
		ctx.Set("userRole", "Analyst")
		isAdmin2, err2 := checker.IsAdminFromContext(ctx)
		assert.NoError(t, err2)
		assert.False(t, isAdmin2)
	})
}

// TestNewContextAdminChecker tests the constructor
func TestNewContextAdminChecker(t *testing.T) {
	t.Run("creates new instance", func(t *testing.T) {
		checker := case_assign.NewContextAdminChecker()

		assert.NotNil(t, checker)
		assert.IsType(t, &case_assign.ContextAdminChecker{}, checker)
	})

	t.Run("creates independent instances", func(t *testing.T) {
		checker1 := case_assign.NewContextAdminChecker()
		checker2 := case_assign.NewContextAdminChecker()

		assert.NotNil(t, checker1)
		assert.NotNil(t, checker2)
		// They should be different instances
		assert.NotSame(t, checker1, checker2)
	})
}
