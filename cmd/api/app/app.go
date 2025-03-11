package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/PeterM45/perfolio-api/internal/common/config"
	"github.com/PeterM45/perfolio-api/internal/common/middleware"
	"github.com/PeterM45/perfolio-api/internal/platform/cache"
	"github.com/PeterM45/perfolio-api/internal/platform/database"

	"github.com/PeterM45/perfolio-api/internal/user/handler"
	"github.com/PeterM45/perfolio-api/internal/user/repository"
	"github.com/PeterM45/perfolio-api/internal/user/service"

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
	dbConfig := database.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		Name:            cfg.Database.Name,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	}
	db, err := database.NewPostgresDB(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize cache
	var cacheClient cache.Cache
	if cfg.Cache.Type == "redis" {
		redisCache, err := cache.NewRedisCache(cfg.Cache.RedisURL)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to connect to Redis, falling back to in-memory cache")
			cacheClient = cache.NewInMemoryCache(cfg.Cache.DefaultTTL)
		} else {
			cacheClient = redisCache
		}
	} else {
		cacheClient = cache.NewInMemoryCache(cfg.Cache.DefaultTTL)
	}

	// Initialize repositories
	userRepository := repository.NewUserRepository(db)
	postRepository := repository.NewPostRepository(db)
	widgetRepository := repository.NewWidgetRepository(db)

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.Auth.JWTSecret)

	// Initialize services
	userSvc := service.NewUserService(userRepository, cacheClient, log)
	postSvc := service.NewPostService(postRepository, userSvc, cacheClient, log)
	widgetSvc := service.NewWidgetService(widgetRepository, userSvc, cacheClient, log)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userSvc, log)
	postHandler := handler.NewPostHandler(postSvc, log)
	widgetHandler := handler.NewWidgetHandler(widgetSvc, log)
	authHandler := handler.NewAuthHandler(userSvc, authMiddleware, log)

	// Initialize router
	router := NewRouter(userHandler, postHandler, widgetHandler, authHandler, authMiddleware, log)

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
	if _, ok := a.cache.(*cache.RedisCache); ok {
		a.logger.Info().Msg("Redis cache is being used, but no close method is available.")
	}

	a.logger.Info().Msg("Shutting down HTTP server...")
	return a.server.Shutdown(ctx)
}
