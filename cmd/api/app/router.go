package app

import (
	"time"

	"github.com/PeterM45/perfolio-api/internal/common/middleware"
	contentHandler "github.com/PeterM45/perfolio-api/internal/user/handler"
	userHandler "github.com/PeterM45/perfolio-api/internal/user/handler"
	"github.com/PeterM45/perfolio-api/pkg/logger"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// NewRouter sets up the HTTP router with all routes
func NewRouter(
	userHandler *userHandler.UserHandler,
	postHandler *contentHandler.PostHandler,
	widgetHandler *contentHandler.WidgetHandler,
	authMiddleware *middleware.AuthMiddleware,
	log logger.Logger,
) *gin.Engine {
	// Set Gin mode based on environment
	gin.SetMode(gin.ReleaseMode)

	// Create router
	router := gin.New()

	// Apply middleware
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggerMiddleware(log))
	router.Use(gin.Recovery())

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://*.perfolio.com", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check route
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Register user routes
		userHandler.RegisterRoutes(v1)

		// Register post routes
		postHandler.RegisterRoutes(v1)

		// Register widget routes
		widgetHandler.RegisterRoutes(v1)

		// Webhook routes
		webhooks := v1.Group("/webhooks")
		{
			// Clerk webhook route
			clerk := webhooks.Group("/clerk").Use(authMiddleware.ClerkWebhookHandler())
			{
				clerk.POST("/", func(c *gin.Context) {
					// Process Clerk webhook event
					c.JSON(200, gin.H{"status": "webhook received"})
				})
			}
		}
	}

	return router
}
