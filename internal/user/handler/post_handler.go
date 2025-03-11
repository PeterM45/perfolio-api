package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/PeterM45/perfolio-api/internal/common/model"
	"github.com/PeterM45/perfolio-api/internal/user/interfaces"
	"github.com/PeterM45/perfolio-api/pkg/apperrors"
	"github.com/PeterM45/perfolio-api/pkg/logger"
	"github.com/gin-gonic/gin"
)

// PostHandler handles HTTP requests for posts
type PostHandler struct {
	service interfaces.PostService
	logger  logger.Logger
}

// NewPostHandler creates a new PostHandler
func NewPostHandler(service interfaces.PostService, logger logger.Logger) *PostHandler {
	return &PostHandler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers all routes for backward compatibility
func (h *PostHandler) RegisterRoutes(router *gin.RouterGroup) {
	posts := router.Group("/posts")
	{
		// Public routes
		posts.GET("/:id", h.GetPost)
		posts.GET("/user/:userId", h.GetUserPosts)

		// Protected routes - will check auth internally
		posts.POST("/", h.CreatePost)
		posts.PUT("/:id", h.UpdatePost)
		posts.DELETE("/:id", h.DeletePost)
		posts.GET("/feed", h.GetFeed)
	}
}

// RegisterProtectedRoutes registers routes that require authentication
func (h *PostHandler) RegisterProtectedRoutes(router *gin.RouterGroup) {
	router.POST("/", h.CreatePost)
	router.PUT("/:id", h.UpdatePost)
	router.DELETE("/:id", h.DeletePost)
	router.GET("/feed", h.GetFeed)
}

// RegisterPublicRoutes registers routes that don't require authentication
func (h *PostHandler) RegisterPublicRoutes(router *gin.RouterGroup) {
	router.GET("/:id", h.GetPost)
	router.GET("/user/:userId", h.GetUserPosts)
}

// GetPost handles GET /posts/:id
func (h *PostHandler) GetPost(c *gin.Context) {
	id := c.Param("id")

	h.logger.Debug().Str("post_id", id).Msg("Getting post by ID")

	post, err := h.service.GetPostByID(c, id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, post)
}

// CreatePost handles POST /posts
func (h *PostHandler) CreatePost(c *gin.Context) {
	// Get authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req model.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Debug().
		Str("user_id", userID.(string)).
		Interface("request", req).
		Msg("Creating post")

	post, err := h.service.CreatePost(c, userID.(string), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, post)
}

// UpdatePost handles PUT /posts/:id
func (h *PostHandler) UpdatePost(c *gin.Context) {
	id := c.Param("id")

	// Get authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req model.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Debug().
		Str("post_id", id).
		Str("user_id", userID.(string)).
		Interface("request", req).
		Msg("Updating post")

	post, err := h.service.UpdatePost(c, id, userID.(string), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, post)
}

// DeletePost handles DELETE /posts/:id
func (h *PostHandler) DeletePost(c *gin.Context) {
	id := c.Param("id")

	// Get authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	h.logger.Debug().
		Str("post_id", id).
		Str("user_id", userID.(string)).
		Msg("Deleting post")

	err := h.service.DeletePost(c, id, userID.(string))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetUserPosts handles GET /posts/user/:userId
func (h *PostHandler) GetUserPosts(c *gin.Context) {
	userID := c.Param("userId")

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	h.logger.Debug().
		Str("user_id", userID).
		Int("limit", limit).
		Int("offset", offset).
		Msg("Getting user posts")

	posts, err := h.service.GetUserPosts(c, userID, limit, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"posts": posts})
}

// GetFeed handles GET /posts/feed
func (h *PostHandler) GetFeed(c *gin.Context) {
	// Get authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	h.logger.Debug().
		Str("user_id", userID.(string)).
		Int("limit", limit).
		Int("offset", offset).
		Msg("Getting feed")

	posts, err := h.service.GetFeed(c, userID.(string), limit, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"posts": posts})
}

// handleError handles errors and returns appropriate HTTP responses
func (h *PostHandler) handleError(c *gin.Context, err error) {
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
