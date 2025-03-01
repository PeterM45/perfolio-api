package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/PeterM45/perfolio-api/internal/common/config"
	"github.com/PeterM45/perfolio-api/internal/common/middleware"
	contentHandler "github.com/PeterM45/perfolio-api/internal/content/handler"
	contentRepo "github.com/PeterM45/perfolio-api/internal/content/repository"
	contentService "github.com/PeterM45/perfolio-api/internal/content/service"
	"github.com/PeterM45/perfolio-api/internal/platform/cache"
	"github.com/PeterM45/perfolio-api/internal/platform/database"
	userHandler "github.com/PeterM45/perfolio-api/internal/user/handler"
	userRepo "github.com/PeterM45/perfolio-api/internal/user/repository"
	userService "github.com/PeterM45/perfolio-api/internal/user/service"
	"github.com/PeterM45/perfolio-api/pkg/logger"
)

// Application represents the API application
type Application struct {
	config *config.Config
	server *http.Server
	logger logger.Logger
	db     *database.DB
	cache  cache.Cache
}

// New creates a new application
func New(cfg *config.Config, log logger.Logger) (*Application, error) {
	// Initialize database
	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize cache
	var cacheClient cache.Cache
	if cfg.Cache.Type == "redis" {
		redisCache, err := cache.NewRedisCache(cfg.Cache.RedisURL)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to connect to Redis, falling back to in-memory cache")
			cacheClient = cache.NewInMemoryCache()
		} else {
			cacheClient = redisCache
		}
	} else {
		cacheClient = cache.NewInMemoryCache()
	}

	// Initialize repositories
	userRepository := userRepo.NewUserRepository(db)
	postRepository := contentRepo.NewPostRepository(db)
	widgetRepository := contentRepo.NewWidgetRepository(db)

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.Auth.ClerkSecretKey)

	// Initialize services
	userSvc := userService.NewUserService(userRepository, cacheClient, log)
	postSvc := contentService.NewPostService(postRepository, userSvc, cacheClient, log)
	widgetSvc := contentService.NewWidgetService(widgetRepository, userSvc, cacheClient, log)

	// Initialize handlers
	userHandler := userHandler.NewUserHandler(userSvc, log)
	postHandler := contentHandler.NewPostHandler(postSvc, log)
	widgetHandler := contentHandler.NewWidgetHandler(widgetSvc, log)

	// Initialize router
	router := NewRouter(userHandler, postHandler, widgetHandler, authMiddleware, log)

	// Create server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return &Application{
		config: cfg,
		server: server,
		logger: log,
		db:     db,
		cache:  cacheClient,
	}, nil
}

// Start starts the HTTP server
func (a *Application) Start() error {
	return a.server.ListenAndServe()
}

// Stop gracefully shuts down the server
func (a *Application) Stop(ctx context.Context) error {
	a.logger.Info().Msg("Closing database connections...")
	if err := a.db.Close(); err != nil {
		a.logger.Error().Err(err).Msg("Error closing database connections")
	}

	// If Redis cache, close it
	if redisCache, ok := a.cache.(*cache.RedisCache); ok {
		a.logger.Info().Msg("Closing Redis connection...")
		if err := redisCache.Close(); err != nil {
			a.logger.Error().Err(err).Msg("Error closing Redis connection")
		}
	}

	a.logger.Info().Msg("Shutting down HTTP server...")
	return a.server.Shutdown(ctx)
}
