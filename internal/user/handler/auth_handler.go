package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PeterM45/perfolio-api/internal/common/middleware"
	"github.com/PeterM45/perfolio-api/internal/common/model"
	"github.com/PeterM45/perfolio-api/internal/user/interfaces"
	"github.com/PeterM45/perfolio-api/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	userService    interfaces.UserService
	authMiddleware *middleware.AuthMiddleware
	logger         logger.Logger
	tokenExpiry    time.Duration
	refreshExpiry  time.Duration
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(
	userService interfaces.UserService,
	authMiddleware *middleware.AuthMiddleware,
	logger logger.Logger,
) *AuthHandler {
	return &AuthHandler{
		userService:    userService,
		authMiddleware: authMiddleware,
		logger:         logger,
		tokenExpiry:    time.Hour * 24,      // 24 hours
		refreshExpiry:  time.Hour * 24 * 30, // 30 days
	}
}

// RegisterRoutes registers routes for the auth handler
func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.RefreshToken)

		// Protected routes that require authentication
		authenticated := auth.Group("")
		authenticated.Use(h.authMiddleware.Authenticate())
		{
			authenticated.POST("/logout", h.Logout)
			authenticated.GET("/me", h.GetCurrentUser)
		}
	}
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,alphanum,min=3,max=30"`
	Password string `json:"password" binding:"required,min=6"`
}

// TokenResponse represents the authentication token response
type TokenResponse struct {
	AccessToken  string      `json:"accessToken"`
	RefreshToken string      `json:"refreshToken"`
	ExpiresAt    time.Time   `json:"expiresAt"`
	User         *model.User `json:"user"`
}

// RefreshTokenRequest represents the refresh token request body
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// Register handles POST /auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Debug().Str("email", req.Email).Str("username", req.Username).Msg("Registering new user")

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to hash password")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process registration"})
		return
	}

	// Generate ID for new user
	userID := uuid.New().String()

	// Create first and last name from name
	nameParts := splitName(req.Name)
	firstName := nameParts[0]
	var lastName *string
	if len(nameParts) > 1 {
		lastNameStr := nameParts[len(nameParts)-1]
		lastName = &lastNameStr
	}

	// Create user request
	createUserReq := &model.CreateUserRequest{
		ID:           userID,
		Email:        req.Email,
		Username:     req.Username,
		FirstName:    &firstName,
		LastName:     lastName,
		PasswordHash: string(hashedPassword),
		AuthProvider: model.AuthProviderCustom,
	}

	// Create the user
	user, err := h.userService.CreateUser(c, createUserReq)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + err.Error()})
		return
	}

	// Generate token
	accessToken, err := h.authMiddleware.GenerateToken(user.ID, []string{"user"}, h.tokenExpiry)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to generate token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Generate refresh token
	refreshToken, err := h.authMiddleware.GenerateToken(user.ID, []string{"refresh"}, h.refreshExpiry)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to generate refresh token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	// Return the tokens
	c.JSON(http.StatusCreated, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(h.tokenExpiry),
		User:         user,
	})
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Debug().Str("email", req.Email).Msg("User login attempt")

	// Get user by email
	user, err := h.userService.GetUserByEmail(c, req.Email)
	if err != nil {
		h.logger.Debug().Err(err).Msg("User not found or invalid credentials")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		h.logger.Debug().Err(err).Msg("Invalid password")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Generate token
	accessToken, err := h.authMiddleware.GenerateToken(user.ID, []string{"user"}, h.tokenExpiry)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to generate token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Generate refresh token
	refreshToken, err := h.authMiddleware.GenerateToken(user.ID, []string{"refresh"}, h.refreshExpiry)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to generate refresh token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	// Return the tokens
	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(h.tokenExpiry),
		User:         user,
	})
}

// RefreshToken handles POST /auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse the token
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.authMiddleware.GetSecretKey()), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}

	// Check token type (should be a refresh token)
	roles, ok := claims["roles"].([]interface{})
	if !ok || len(roles) == 0 || roles[0].(string) != "refresh" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token type"})
		return
	}

	// Extract user ID
	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
		return
	}

	// Get user
	user, err := h.userService.GetUserByID(c, userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Generate new access token
	accessToken, err := h.authMiddleware.GenerateToken(userID, []string{"user"}, h.tokenExpiry)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to generate token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Generate new refresh token
	refreshToken, err := h.authMiddleware.GenerateToken(userID, []string{"refresh"}, h.refreshExpiry)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to generate refresh token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	// Return the new tokens
	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(h.tokenExpiry),
		User:         user,
	})
}

// Logout handles POST /auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// In a stateless JWT setup, the client simply discards the tokens
	// If you want to implement token blacklisting, you'd need to add that logic here

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

// GetCurrentUser handles GET /auth/me
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get user
	user, err := h.userService.GetUserByID(c, userID.(string))
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get current user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user details"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Helper function to split a full name into parts
func splitName(name string) []string {
	// You could use a more sophisticated name parsing library if needed
	// This is a simple space-based split
	return strings.Fields(name)
}
