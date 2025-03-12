package handler

import (
	"errors"
	"net/http"

	"github.com/PeterM45/perfolio-api/internal/common/model"
	"github.com/PeterM45/perfolio-api/internal/user/interfaces"
	"github.com/PeterM45/perfolio-api/pkg/apperrors"
	"github.com/PeterM45/perfolio-api/pkg/logger"
	"github.com/gin-gonic/gin"
)

// WidgetHandler handles HTTP requests for widgets
type WidgetHandler struct {
	service interfaces.WidgetService
	logger  logger.Logger
}

// NewWidgetHandler creates a new WidgetHandler
func NewWidgetHandler(service interfaces.WidgetService, logger logger.Logger) *WidgetHandler {
	return &WidgetHandler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers all routes for backward compatibility
func (h *WidgetHandler) RegisterRoutes(router *gin.RouterGroup) {
	widgets := router.Group("/widgets")
	{
		// Public routes
		widgets.GET("/:id", h.GetWidget)
		widgets.GET("/user/:userId", h.GetUserWidgets)

		// Protected routes - will check auth internally
		widgets.POST("/", h.CreateWidget)
		widgets.PUT("/:id", h.UpdateWidget)
		widgets.DELETE("/:id", h.DeleteWidget)
		widgets.POST("/batch-update", h.BatchUpdateWidgets)
	}
}

// RegisterProtectedRoutes registers routes that require authentication
func (h *WidgetHandler) RegisterProtectedRoutes(router *gin.RouterGroup) {
	router.POST("/", h.CreateWidget)
	router.PUT("/:id", h.UpdateWidget)
	router.DELETE("/:id", h.DeleteWidget)
	router.POST("/batch-update", h.BatchUpdateWidgets)
}

// RegisterPublicRoutes registers routes that don't require authentication
func (h *WidgetHandler) RegisterPublicRoutes(router *gin.RouterGroup) {
	router.GET("/:id", h.GetWidget)
	router.GET("/user/:userId", h.GetUserWidgets)
	router.GET("/types", h.GetWidgetTypes) // Add this line
}

// GetWidget handles GET /widgets/:id
func (h *WidgetHandler) GetWidget(c *gin.Context) {
	id := c.Param("id")

	h.logger.Debug().Str("widget_id", id).Msg("Getting widget by ID")

	widget, err := h.service.GetWidgetByID(c, id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, widget)
}

// GetUserWidgets handles GET /widgets/user/:userId
func (h *WidgetHandler) GetUserWidgets(c *gin.Context) {
	userID := c.Param("userId")

	h.logger.Debug().Str("user_id", userID).Msg("Getting user widgets")

	widgets, err := h.service.GetUserWidgets(c, userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"widgets": widgets})
}

// CreateWidget handles POST /widgets
func (h *WidgetHandler) CreateWidget(c *gin.Context) {
	// Get authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req model.CreateWidgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Debug().
		Str("user_id", userID.(string)).
		Interface("request", req).
		Msg("Creating widget")

	widget, err := h.service.CreateWidget(c, userID.(string), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, widget)
}

// UpdateWidget handles PUT /widgets/:id
func (h *WidgetHandler) UpdateWidget(c *gin.Context) {
	id := c.Param("id")

	// Get authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req model.UpdateWidgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Debug().
		Str("widget_id", id).
		Str("user_id", userID.(string)).
		Interface("request", req).
		Msg("Updating widget")

	widget, err := h.service.UpdateWidget(c, id, userID.(string), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, widget)
}

// DeleteWidget handles DELETE /widgets/:id
func (h *WidgetHandler) DeleteWidget(c *gin.Context) {
	id := c.Param("id")

	// Get authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	h.logger.Debug().
		Str("widget_id", id).
		Str("user_id", userID.(string)).
		Msg("Deleting widget")

	err := h.service.DeleteWidget(c, id, userID.(string))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// BatchUpdateWidgets handles POST /widgets/batch-update
func (h *WidgetHandler) BatchUpdateWidgets(c *gin.Context) {
	// Get authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req model.BatchUpdateWidgetsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Debug().
		Str("user_id", userID.(string)).
		Int("update_count", len(req.Updates)).
		Msg("Batch updating widgets")

	err := h.service.BatchUpdateWidgets(c, userID.(string), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetWidgetTypes handles GET /widgets/types
func (h *WidgetHandler) GetWidgetTypes(c *gin.Context) {
	types, err := h.service.GetWidgetTypes(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"types": types})
}

// handleError handles errors and returns appropriate HTTP responses
func (h *WidgetHandler) handleError(c *gin.Context, err error) {
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
