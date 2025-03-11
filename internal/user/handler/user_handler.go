package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/PeterM45/perfolio-api/internal/common/interfaces"
	"github.com/PeterM45/perfolio-api/internal/common/model"
	"github.com/PeterM45/perfolio-api/pkg/apperrors"
	"github.com/PeterM45/perfolio-api/pkg/logger"
	"github.com/gin-gonic/gin"
)

// UserHandler handles HTTP requests for users
type UserHandler struct {
	service interfaces.UserService
	logger  logger.Logger
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(service interfaces.UserService, logger logger.Logger) *UserHandler {
	return &UserHandler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers all routes for backward compatibility
// This will be deprecated in favor of RegisterProtectedRoutes and RegisterPublicRoutes
func (h *UserHandler) RegisterRoutes(router *gin.RouterGroup) {
	users := router.Group("/users")
	{
		// Public profile routes
		users.GET("/:id", h.GetUserByID)
		users.GET("/username/:username", h.GetUserByUsername)
		users.GET("/search", h.SearchUsers)
		users.GET("/:id/is-following/:targetId", h.IsFollowing)
		users.GET("/:id/stats", h.GetProfileStats)
		users.GET("/:id/followers", h.GetFollowers)
		users.GET("/:id/following", h.GetFollowing)

		// These routes should be protected but are kept for backward compatibility
		// They will check for authentication internally
		users.POST("/", h.CreateUser)
		users.PUT("/:id", h.UpdateUser)
		users.POST("/:id/follow", h.ToggleFollow)
	}
}

// RegisterProtectedRoutes registers routes that require authentication
func (h *UserHandler) RegisterProtectedRoutes(router *gin.RouterGroup) {
	// Routes that require authentication
	router.POST("/", h.CreateUser)
	router.PUT("/:id", h.UpdateUser)
	router.POST("/:id/follow", h.ToggleFollow)

	// Admin-only routes could be added here
}

// RegisterPublicRoutes registers routes that don't require authentication
func (h *UserHandler) RegisterPublicRoutes(router *gin.RouterGroup) {
	// Routes that don't require authentication
	router.GET("/:id", h.GetUserByID)
	router.GET("/username/:username", h.GetUserByUsername)
	router.GET("/search", h.SearchUsers)
	router.GET("/:id/is-following/:targetId", h.IsFollowing)
	router.GET("/:id/stats", h.GetProfileStats)
	router.GET("/:id/followers", h.GetFollowers)
	router.GET("/:id/following", h.GetFollowing)
}

// GetUserByID handles GET /users/:id
func (h *UserHandler) GetUserByID(c *gin.Context) {
	id := c.Param("id")

	h.logger.Debug().Str("user_id", id).Msg("Getting user by ID")

	user, err := h.service.GetUserByID(c, id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetUserByUsername handles GET /users/username/:username
func (h *UserHandler) GetUserByUsername(c *gin.Context) {
	username := c.Param("username")

	h.logger.Debug().Str("username", username).Msg("Getting user by username")

	user, err := h.service.GetUserByUsername(c, username)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// CreateUser handles POST /users
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Debug().Interface("request", req).Msg("Creating user")

	user, err := h.service.CreateUser(c, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, user)
}

// UpdateUser handles PUT /users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	// Check authentication
	userID, exists := c.Get("userID")
	if !exists || userID.(string) != id {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only update your own profile"})
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Debug().Str("user_id", id).Interface("request", req).Msg("Updating user")

	user, err := h.service.UpdateUser(c, id, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// SearchUsers handles GET /users/search
func (h *UserHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query is required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	h.logger.Debug().Str("query", query).Int("limit", limit).Msg("Searching users")

	users, err := h.service.SearchUsers(c, query, limit)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// ToggleFollow handles POST /users/:id/follow
func (h *UserHandler) ToggleFollow(c *gin.Context) {
	// Get authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req model.FollowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Debug().
		Str("follower_id", userID.(string)).
		Str("following_id", req.FollowingID).
		Str("action", req.Action).
		Msg("Toggle follow")

	err := h.service.ToggleFollow(c, &req, userID.(string))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// IsFollowing handles GET /users/:id/is-following/:targetId
func (h *UserHandler) IsFollowing(c *gin.Context) {
	userID := c.Param("id")
	targetID := c.Param("targetId")

	h.logger.Debug().
		Str("user_id", userID).
		Str("target_id", targetID).
		Msg("Checking if user is following target")

	isFollowing, err := h.service.IsFollowing(c, userID, targetID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"isFollowing": isFollowing})
}

// GetProfileStats handles GET /users/:id/stats
func (h *UserHandler) GetProfileStats(c *gin.Context) {
	userID := c.Param("id")

	h.logger.Debug().Str("user_id", userID).Msg("Getting profile stats")

	stats, err := h.service.GetProfileStats(c, userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetFollowers handles GET /users/:id/followers
func (h *UserHandler) GetFollowers(c *gin.Context) {
	userID := c.Param("id")

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	h.logger.Debug().
		Str("user_id", userID).
		Int("limit", limit).
		Int("offset", offset).
		Msg("Getting followers")

	followers, err := h.service.GetFollowers(c, userID, limit, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": followers})
}

// GetFollowing handles GET /users/:id/following
func (h *UserHandler) GetFollowing(c *gin.Context) {
	userID := c.Param("id")

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	h.logger.Debug().
		Str("user_id", userID).
		Int("limit", limit).
		Int("offset", offset).
		Msg("Getting following")

	following, err := h.service.GetFollowing(c, userID, limit, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": following})
}

// handleError handles errors and returns appropriate HTTP responses
func (h *UserHandler) handleError(c *gin.Context, err error) {
	var appErr *apperrors.Error
	if errors.As(err, &appErr) {
		switch appErr.Type() {
		case apperrors.ErrTypeNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": appErr.Error()})
		case apperrors.ErrTypeBadRequest:
			c.JSON(http.StatusBadRequest, gin.H{"error": appErr.Error()})
		case apperrors.ErrTypeUnauthorized:
			c.JSON(http.StatusUnauthorized, gin.H{"error": appErr.Error()})
		case apperrors.ErrTypeForbidden:
			c.JSON(http.StatusForbidden, gin.H{"error": appErr.Error()})
		case apperrors.ErrTypeConflict:
			c.JSON(http.StatusConflict, gin.H{"error": appErr.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	// If not an AppError, treat as internal server error
	h.logger.Error().Err(err).Msg("Internal server error")
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
}
