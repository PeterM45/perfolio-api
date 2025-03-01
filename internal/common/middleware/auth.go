package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware handles authentication with Clerk
type AuthMiddleware struct {
	clerkSecretKey string
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(clerkSecretKey string) *AuthMiddleware {
	return &AuthMiddleware{
		clerkSecretKey: clerkSecretKey,
	}
}

// Authenticate verifies the JWT token from Clerk
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
			// Validate the alg
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// Return the key for validation
			return []byte(m.clerkSecretKey), nil
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

		// Extract user ID from claims (adjust based on Clerk's JWT format)
		sub, ok := claims["sub"].(string)
		if !ok || sub == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID in token"})
			c.Abort()
			return
		}

		// Set the user ID in context
		c.Set("userID", sub)
		c.Next()
	}
}

// Optional middleware to check if user is authenticated
// but doesn't block if they're not
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
			return []byte(m.clerkSecretKey), nil
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
		sub, ok := claims["sub"].(string)
		if !ok || sub == "" {
			c.Next()
			return
		}

		// Set the user ID in context
		c.Set("userID", sub)
		c.Next()
	}
}

// ClerkWebhookHandler is a middleware to handle Clerk webhooks
// This verifies the webhook signature from Clerk
func (m *AuthMiddleware) ClerkWebhookHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		svix_id := c.GetHeader("svix-id")
		svix_timestamp := c.GetHeader("svix-timestamp")
		svix_signature := c.GetHeader("svix-signature")

		// Verify webhook
		if svix_id == "" || svix_timestamp == "" || svix_signature == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing svix headers"})
			c.Abort()
			return
		}

		// TODO: Implement actual signature verification if needed
		// This is simplified for now and should be expanded based on Clerk's docs
		// https://clerk.dev/docs/integrations/webhooks

		c.Next()
	}
}
