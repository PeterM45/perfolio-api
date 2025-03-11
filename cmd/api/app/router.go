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
	authHandler *userHandler.AuthHandler,
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
		// Register auth routes - must come first to handle auth endpoints
		authHandler.RegisterRoutes(v1)

		// Public routes - no authentication required
		public := v1.Group("/public")
		{
			// Add any public routes here
			public.GET("/config", func(c *gin.Context) {
				c.JSON(200, gin.H{"version": "1.0"})
			})
		}

		// Protected routes - require authentication
		protected := v1.Group("")
		protected.Use(authMiddleware.Authenticate())
		{
			// Register user routes that need authentication
			userGroup := protected.Group("/users")
			userHandler.RegisterProtectedRoutes(userGroup)

			// Register protected post routes
			postGroup := protected.Group("/posts")
			postHandler.RegisterProtectedRoutes(postGroup)

			// Register protected widget routes
			widgetGroup := protected.Group("/widgets")
			widgetHandler.RegisterProtectedRoutes(widgetGroup)
		}

		// Optional authentication routes
		optional := v1.Group("")
		optional.Use(authMiddleware.OptionalAuthenticate())
		{
			// Register user routes with optional authentication
			userHandler.RegisterPublicRoutes(v1.Group("/users"))

			// Register post routes with optional authentication
			postHandler.RegisterPublicRoutes(v1.Group("/posts"))

			// Register widget routes with optional authentication
			widgetHandler.RegisterPublicRoutes(v1.Group("/widgets"))
		}
	}

	return router
}
