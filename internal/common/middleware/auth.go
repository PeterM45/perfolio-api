package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware handles authentication with custom JWT
type AuthMiddleware struct {
	jwtSecretKey string
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(jwtSecretKey string) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecretKey: jwtSecretKey,
	}
}

// Authenticate verifies the JWT token
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			c.Abort()
			return
		}

		// Check if it starts with "Bearer "
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token is required"})
			c.Abort()
			return
		}

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the alg is what we expect
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// Return the key for validation
			return []byte(m.jwtSecretKey), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token: " + err.Error()})
			c.Abort()
			return
		}

		// Check if the token is valid
		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}

		// Check expiration
		exp, ok := claims["exp"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token expiration"})
			c.Abort()
			return
		}

		if time.Now().Unix() > int64(exp) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
			c.Abort()
			return
		}

		// Extract user ID from claims
		userID, ok := claims["user_id"].(string)
		if !ok || userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID in token"})
			c.Abort()
			return
		}

		// Extract additional claims if needed
		// For example, user roles or permissions
		if roles, ok := claims["roles"].([]interface{}); ok {
			c.Set("userRoles", roles)
		}

		// Set the user ID in context
		c.Set("userID", userID)
		c.Next()
	}
}

// OptionalAuthenticate checks for authentication but doesn't block if not present
func (m *AuthMiddleware) OptionalAuthenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Check if it starts with "Bearer "
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			c.Next()
			return
		}

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the alg
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// Return the key for validation
			return []byte(m.jwtSecretKey), nil
		})

		if err != nil || !token.Valid {
			c.Next()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.Next()
			return
		}

		// Check expiration
		exp, ok := claims["exp"].(float64)
		if !ok || time.Now().Unix() > int64(exp) {
			c.Next()
			return
		}

		// Extract user ID from claims
		userID, ok := claims["user_id"].(string)
		if !ok || userID == "" {
			c.Next()
			return
		}

		// Extract additional claims if needed
		if roles, ok := claims["roles"].([]interface{}); ok {
			c.Set("userRoles", roles)
		}

		// Set the user ID in context
		c.Set("userID", userID)
		c.Next()
	}
}

// GenerateToken creates a new JWT token for a user
func (m *AuthMiddleware) GenerateToken(userID string, roles []string, expiration time.Duration) (string, error) {
	// Create the token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["roles"] = roles
	claims["exp"] = time.Now().Add(expiration).Unix()
	claims["iat"] = time.Now().Unix()

	// Generate encoded token
	tokenString, err := token.SignedString([]byte(m.jwtSecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
