// test/middleware/auth_middleware_test.go
package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/PeterM45/perfolio-api/internal/common/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware_Authenticate(t *testing.T) {
	// Setup
	testSecret := "test-jwt-secret"
	middleware := middleware.NewAuthMiddleware(testSecret)

	// Generate test token
	token, err := middleware.GenerateToken("test-user", []string{"user"}, 1*time.Hour)
	require.NoError(t, err)

	// Setup test router
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.Authenticate())
	r.GET("/protected", func(c *gin.Context) {
		userID, exists := c.Get("userID")
		assert.True(t, exists)
		assert.Equal(t, "test-user", userID)
		c.Status(http.StatusOK)
	})

	// Test valid token
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test missing token
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/protected", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
