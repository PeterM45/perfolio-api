// This file demonstrates the application structure with key files and implementation examples

// =====================================================================
// cmd/api/main.go
// =====================================================================
package main

import (
"context"
"log"
"os"
"os/signal"
"syscall"
"time"

    "github.com/PeterM45/perfolio-api/cmd/api/app"
    "github.com/PeterM45/perfolio-api/internal/common/config"
    "github.com/PeterM45/perfolio-api/pkg/logger"

)

func main() {
// Initialize logger first to catch early errors
log := logger.NewZapLogger("debug")
defer log.Sync()

    // Load configuration
    cfg, err := config.Load()
    if err != nil {
    	log.Fatal().Err(err).Msg("Failed to load configuration")
    }

    // Create application
    application, err := app.New(cfg, log)
    if err != nil {
    	log.Fatal().Err(err).Msg("Failed to create application")
    }

    // Start server in a goroutine
    go func() {
    	if err := application.Start(); err != nil {
    		log.Fatal().Err(err).Msg("Failed to start server")
    	}
    }()

    // Graceful shutdown handling
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    sig := <-quit

    log.Info().Str("signal", sig.String()).Msg("Shutting down server...")

    // Use timeout context for shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := application.Stop(ctx); err != nil {
    	log.Error().Err(err).Msg("Server forced to shutdown")
    	os.Exit(1)
    }

    log.Info().Msg("Server gracefully stopped")

}

// =====================================================================
// cmd/api/app.go
// =====================================================================
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
db \*database.DB
}

// New creates a new application instance
func New(cfg *config.Config, log logger.Logger) (*Application, error) {
// Initialize database with connection pooling
db, err := database.NewPostgresDB(cfg.Database)
if err != nil {
return nil, fmt.Errorf("failed to connect to database: %w", err)
}
// Initialize cache (in-memory initially, can be replaced with Redis later)
cacheClient := cache.NewInMemoryCache()
// Initialize repositories
userRepository := userRepo.NewUserRepository(db)
contentRepository := contentRepo.NewContentRepository(db)
// Initialize services
userSvc := userService.NewUserService(userRepository, cacheClient, log)
contentSvc := contentService.NewContentService(contentRepository, userSvc, cacheClient, log)
// Initialize middleware
authMiddleware := middleware.NewAuthMiddleware(cfg.Auth.ClerkSecretKey)
// Initialize handlers
userHandler := userHandler.NewUserHandler(userSvc, log)
contentHandler := contentHandler.NewContentHandler(contentSvc, log)
// Initialize router with middleware
router := NewRouter(userHandler, contentHandler, authMiddleware, log)
// Create server with timeouts
server := &http.Server{
Addr: fmt.Sprintf(":%d", cfg.Server.Port),
Handler: router,
ReadTimeout: cfg.Server.ReadTimeout,
WriteTimeout: cfg.Server.WriteTimeout,
IdleTimeout: cfg.Server.IdleTimeout,
}
return &Application{
config: cfg,
server: server,
logger: log,
db: db,
}, nil
}

// Start starts the HTTP server
func (a \*Application) Start() error {
a.logger.Info().Int("port", a.config.Server.Port).Msg("Starting server")
return a.server.ListenAndServe()
}

// Stop gracefully shuts down the server
func (a \*Application) Stop(ctx context.Context) error {
// Close DB connections
if err := a.db.Close(); err != nil {
a.logger.Error().Err(err).Msg("Error closing database connections")
}
// Shutdown HTTP server
return a.server.Shutdown(ctx)
}

// =====================================================================
// cmd/api/router.go
// =====================================================================
package app

import (
"time"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    "github.com/PeterM45/perfolio-api/internal/common/middleware"
    contentHandler "github.com/PeterM45/perfolio-api/internal/content/handler"
    userHandler "github.com/PeterM45/perfolio-api/internal/user/handler"
    "github.com/PeterM45/perfolio-api/pkg/logger"

)

// NewRouter sets up the HTTP router
func NewRouter(
userHandler *userHandler.UserHandler,
contentHandler *contentHandler.ContentHandler,
authMiddleware *middleware.AuthMiddleware,
log logger.Logger,
) *gin.Engine {
// Set Gin mode based on environment
gin.SetMode(gin.ReleaseMode)
router := gin.New()
// Apply middleware
router.Use(middleware.RequestIDMiddleware())
router.Use(middleware.LoggerMiddleware(log))
router.Use(gin.Recovery())
// CORS configuration
router.Use(cors.New(cors.Config{
AllowOrigins: []string{"https://perfolio.com", "https://www.perfolio.com", "http://localhost:3000"},
AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
ExposeHeaders: []string{"Content-Length", "X-Request-ID"},
AllowCredentials: true,
MaxAge: 12 * time.Hour,
}))
// Health check endpoints
router.GET("/health", func(c *gin.Context) {
c.JSON(200, gin.H{"status": "ok"})
})
// API routes
v1 := router.Group("/api/v1")
{
// User routes
users := v1.Group("/users")
{
users.GET("", userHandler.ListUsers)
users.GET("/:id", userHandler.GetUserByID)
users.PUT("/:id", authMiddleware.Authenticate(), userHandler.UpdateUser)
users.GET("/:id/connections", userHandler.GetUserConnections)
users.POST("/:id/connections", authMiddleware.Authenticate(), userHandler.AddConnection)
users.DELETE("/:id/connections/:targetId", authMiddleware.Authenticate(), userHandler.RemoveConnection)
}
// Content routes
posts := v1.Group("/posts")
{
posts.GET("", contentHandler.ListPosts)
posts.POST("", authMiddleware.Authenticate(), contentHandler.CreatePost)
posts.GET("/:id", contentHandler.GetPostByID)
posts.PUT("/:id", authMiddleware.Authenticate(), contentHandler.UpdatePost)
posts.DELETE("/:id", authMiddleware.Authenticate(), contentHandler.DeletePost)
}
// Feed route
v1.GET("/feed", authMiddleware.Authenticate(), contentHandler.GetUserFeed)
// Widget routes
widgets := v1.Group("/widgets")
{
widgets.GET("", authMiddleware.Authenticate(), contentHandler.ListUserWidgets)
widgets.POST("", authMiddleware.Authenticate(), contentHandler.CreateWidget)
widgets.GET("/:id", contentHandler.GetWidgetByID)
widgets.PUT("/:id", authMiddleware.Authenticate(), contentHandler.UpdateWidget)
widgets.DELETE("/:id", authMiddleware.Authenticate(), contentHandler.DeleteWidget)
widgets.POST("/batch-update", authMiddleware.Authenticate(), contentHandler.BatchUpdateWidgets)
}
// Webhooks
webhooks := v1.Group("/webhooks")
{
webhooks.POST("/clerk", contentHandler.HandleClerkWebhook)
}
}
return router
}

// =====================================================================
// internal/common/config/config.go
// =====================================================================
package config

import (
"time"

    "github.com/spf13/viper"

)

// Config holds all configuration for the application
type Config struct {
Server struct {
Port int `mapstructure:"port"`
ReadTimeout time.Duration `mapstructure:"read_timeout"`
WriteTimeout time.Duration `mapstructure:"write_timeout"`
IdleTimeout time.Duration `mapstructure:"idle_timeout"`
} `mapstructure:"server"`
Database struct {
Host string `mapstructure:"host"`
Port int `mapstructure:"port"`
User string `mapstructure:"user"`
Password string `mapstructure:"password"`
Name string `mapstructure:"name"`
SSLMode string `mapstructure:"ssl_mode"`
MaxOpenConns int `mapstructure:"max_open_conns"`
MaxIdleConns int `mapstructure:"max_idle_conns"`
ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
} `mapstructure:"database"`
Auth struct {
JWTSecret string `mapstructure:"jwt_secret"`
TokenExpiry time.Duration `mapstructure:"token_expiry"`
ClerkSecretKey string `mapstructure:"clerk_secret_key"`
} `mapstructure:"auth"`
Cache struct {
Type string `mapstructure:"type"`
RedisURL string `mapstructure:"redis_url"`
DefaultTTL time.Duration `mapstructure:"default_ttl"`
} `mapstructure:"cache"`
LogLevel string `mapstructure:"log_level"`
}

// Load loads configuration from environment variables and config files
func Load() (*Config, error) {
viper.SetConfigName("config")
viper.SetConfigType("yaml")
viper.AddConfigPath(".")
viper.AddConfigPath("./configs")
viper.AutomaticEnv()
// Default values
viper.SetDefault("server.port", 8080)
viper.SetDefault("server.read_timeout", time.Second*10)
viper.SetDefault("server.write*timeout", time.Second*10)
viper.SetDefault("server.idle_timeout", time.Second*60)
viper.SetDefault("database.max_open_conns", 25)
viper.SetDefault("database.max_idle_conns", 10)
viper.SetDefault("database.conn_max_lifetime", time.Minute*5)
viper.SetDefault("cache.type", "memory")
viper.SetDefault("cache.default_ttl", time.Minute*5)
viper.SetDefault("log_level", "info")
// Read configuration
if err := viper.ReadInConfig(); err != nil {
// Config file is optional
if *, ok := err.(viper.ConfigFileNotFoundError); !ok {
return nil, err
}
}
var config Config
if err := viper.Unmarshal(&config); err != nil {
return nil, err
}
return &config, nil
}

// =====================================================================
// internal/user/repository/user_repo.go
// =====================================================================
package repository

import (
"context"
"database/sql"
"errors"
"fmt"
"time"

    "github.com/google/uuid"
    "github.com/PeterM45/perfolio-api/internal/common/model"
    "github.com/PeterM45/perfolio-api/internal/platform/database"
    "github.com/PeterM45/perfolio-api/pkg/apperrors"

)

type UserRepository interface {
GetByID(ctx context.Context, id string) (*model.User, error)
GetByEmail(ctx context.Context, email string) (*model.User, error)
GetByAuthID(ctx context.Context, authID string) (*model.User, error)
Create(ctx context.Context, user *model.User) error
Update(ctx context.Context, user *model.User) error
GetConnections(ctx context.Context, userID string, limit, offset int) ([]*model.User, error)
CountConnections(ctx context.Context, userID string) (int, error)
AddConnection(ctx context.Context, userID, targetID string) error
RemoveConnection(ctx context.Context, userID, targetID string) error
IsConnected(ctx context.Context, userID, targetID string) (bool, error)
}

type userRepository struct {
db \*database.DB
}

func NewUserRepository(db \*database.DB) UserRepository {
return &userRepository{
db: db,
}
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
query := `
		SELECT id, email, display_name, bio, avatar_url, auth_id, created_at, updated_at 
		FROM users 
		WHERE id = $1
	`
var user model.User
err := r.db.QueryRowContext(ctx, query, id).Scan(
&user.ID,
&user.Email,
&user.DisplayName,
&user.Bio,
&user.AvatarURL,
&user.AuthID,
&user.CreatedAt,
&user.UpdatedAt,
)
if err != nil {
if errors.Is(err, sql.ErrNoRows) {
return nil, apperrors.NotFound("user", id)
}
return nil, fmt.Errorf("query user: %w", err)
}
return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
query := `
		UPDATE users 
		SET display_name = $1, bio = $2, avatar_url = $3, updated_at = $4
		WHERE id = $5
		RETURNING id
	`
now := time.Now().UTC()
user.UpdatedAt = now
var id string
err := r.db.QueryRowContext(ctx, query,
user.DisplayName,
user.Bio,
user.AvatarURL,
now,
user.ID,
).Scan(&id)
if err != nil {
if errors.Is(err, sql.ErrNoRows) {
return apperrors.NotFound("user", user.ID)
}
return fmt.Errorf("update user: %w", err)
}
return nil
}

func (r \*userRepository) AddConnection(ctx context.Context, userID, targetID string) error {
// Validate both users exist in a transaction
tx, err := r.db.BeginTx(ctx, nil)
if err != nil {
return fmt.Errorf("begin transaction: %w", err)
}
defer tx.Rollback()
// Check if users exist
for _, id := range []string{userID, targetID} {
var exists bool
err := tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", id).Scan(&exists)
if err != nil {
return fmt.Errorf("check user existence: %w", err)
}
if !exists {
return apperrors.NotFound("user", id)
}
}
// Check if connection already exists
var exists bool
err = tx.QueryRowContext(ctx,
"SELECT EXISTS(SELECT 1 FROM connections WHERE follower_id = $1 AND following_id = $2)",
userID, targetID,
).Scan(&exists)
if err != nil {
return fmt.Errorf("check connection existence: %w", err)
}
if exists {
return apperrors.AlreadyExists("connection", fmt.Sprintf("%s-%s", userID, targetID))
}
// Create connection
_, err = tx.ExecContext(ctx,
"INSERT INTO connections (id, follower_id, following_id, created_at) VALUES ($1, $2, $3, $4)",
uuid.New().String(),
userID,
targetID,
time.Now().UTC(),
)
if err != nil {
return fmt.Errorf("create connection: %w", err)
}
if err := tx.Commit(); err != nil {
return fmt.Errorf("commit transaction: %w", err)
}
return nil
}

// Other repository methods implementation...

// =====================================================================
// internal/user/service/user_service.go
// =====================================================================
package service

import (
"context"
"fmt"
"time"

    "github.com/PeterM45/perfolio-api/internal/common/model"
    "github.com/PeterM45/perfolio-api/internal/platform/cache"
    "github.com/PeterM45/perfolio-api/internal/user/repository"
    "github.com/PeterM45/perfolio-api/pkg/apperrors"
    "github.com/PeterM45/perfolio-api/pkg/logger"
    "github.com/PeterM45/perfolio-api/pkg/validator"

)

type UserService interface {
GetUserByID(ctx context.Context, id string) (*model.User, error)
GetUserByEmail(ctx context.Context, email string) (*model.User, error)
CreateUser(ctx context.Context, req *model.CreateUserRequest) (*model.User, error)
UpdateUser(ctx context.Context, id string, req *model.UpdateUserRequest) (*model.User, error)
GetUserConnections(ctx context.Context, userID string, pagination *model.Pagination) (*model.ConnectionsResponse, error)
AddConnection(ctx context.Context, userID, targetID string) error
RemoveConnection(ctx context.Context, userID, targetID string) error
IsUserConnected(ctx context.Context, userID, targetID string) (bool, error)
}

type userService struct {
repo repository.UserRepository
cache cache.Cache
validator validator.Validator
logger logger.Logger
}

func NewUserService(repo repository.UserRepository, cache cache.Cache, logger logger.Logger) UserService {
return &userService{
repo: repo,
cache: cache,
validator: validator.NewValidator(),
logger: logger,
}
}

func (s *userService) GetUserByID(ctx context.Context, id string) (*model.User, error) {
if id == "" {
return nil, apperrors.BadRequest("user id cannot be empty")
}
// Try cache first
cacheKey := fmt.Sprintf("user:%s", id)
if cachedUser, found := s.cache.Get(cacheKey); found {
s.logger.Debug().Str("user_id", id).Msg("User found in cache")
return cachedUser.(*model.User), nil
}
// If not in cache, get from database
user, err := s.repo.GetByID(ctx, id)
if err != nil {
return nil, err
}
// Store in cache for future requests
s.cache.Set(cacheKey, user, 5*time.Minute)
return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, id string, req *model.UpdateUserRequest) (*model.User, error) {
if id == "" {
return nil, apperrors.BadRequest("user id cannot be empty")
}
// Validate request
if err := s.validator.Validate(req); err != nil {
return nil, apperrors.BadRequest(err.Error())
}
// Get existing user
user, err := s.repo.GetByID(ctx, id)
if err != nil {
return nil, err
}
// Update fields if provided
if req.DisplayName != nil {
user.DisplayName = *req.DisplayName
}
if req.Bio != nil {
user.Bio = req.Bio
}
if req.AvatarURL != nil {
user.AvatarURL = req.AvatarURL
}
// Save updated user
if err := s.repo.Update(ctx, user); err != nil {
return nil, err
}
// Invalidate cache
s.cache.Delete(fmt.Sprintf("user:%s", id))
return user, nil
}

func (s \*userService) AddConnection(ctx context.Context, userID, targetID string) error {
// Validate input
if userID == targetID {
return apperrors.BadRequest("cannot connect with yourself")
}
// Add connection in repository
if err := s.repo.AddConnection(ctx, userID, targetID); err != nil {
return err
}
// Invalidate cache for both users' connections
s.cache.Delete(fmt.Sprintf("user_connections:%s", userID))
s.cache.Delete(fmt.Sprintf("user_connections:%s", targetID))
s.cache.Delete(fmt.Sprintf("is_connected:%s:%s", userID, targetID))
return nil
}

// Other service methods implementation...

// =====================================================================
// internal/user/handler/user_handler.go
// =====================================================================
package handler

import (
"net/http"
"strconv"

    "github.com/gin-gonic/gin"
    "github.com/PeterM45/perfolio-api/internal/common/model"
    "github.com/PeterM45/perfolio-api/internal/user/service"
    "github.com/PeterM45/perfolio-api/pkg/apperrors"
    "github.com/PeterM45/perfolio-api/pkg/logger"

)

type UserHandler struct {
service service.UserService
logger logger.Logger
}

func NewUserHandler(service service.UserService, logger logger.Logger) \*UserHandler {
return &UserHandler{
service: service,
logger: logger,
}
}

// GetUserByID handles GET /api/v1/users/:id
func (h *UserHandler) GetUserByID(c *gin.Context) {
requestID, \_ := c.Get("requestID")
log := h.logger.With().Str("requestID", requestID.(string)).Str("handler", "GetUserByID").Logger()
userID := c.Param("id")
log.Debug().Str("user_id", userID).Msg("Getting user by ID")
user, err := h.service.GetUserByID(c, userID)
if err != nil {
h.handleError(c, err)
return
}
c.JSON(http.StatusOK, user)
}

// UpdateUser handles PUT /api/v1/users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
requestID, \_ := c.Get("requestID")
log := h.logger.With().Str("requestID", requestID.(string)).Str("handler", "UpdateUser").Logger()
userID := c.Param("id")
// Get current user from context
currentUserID, exists := c.Get("userID")
if !exists {
c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
return
}
// Check if user is updating their own profile
if currentUserID != userID {
log.Warn().Str("current_user_id", currentUserID.(string)).Str("target_user_id", userID).Msg("Attempted to update another user's profile")
c.JSON(http.StatusForbidden, gin.H{"error": "you can only update your own profile"})
return
}
var req model.UpdateUserRequest
if err := c.ShouldBindJSON(&req); err != nil {
log.Error().Err(err).Msg("Failed to bind request")
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
return
}
updatedUser, err := h.service.UpdateUser(c, userID, &req)
if err != nil {
h.handleError(c, err)
return
}
c.JSON(http.StatusOK, updatedUser)
}

// GetUserConnections handles GET /api/v1/users/:id/connections
func (h *UserHandler) GetUserConnections(c *gin.Context) {
requestID, _ := c.Get("requestID")
log := h.logger.With().Str("requestID", requestID.(string)).Str("handler", "GetUserConnections").Logger()
userID := c.Param("id")
// Parse pagination parameters
page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
if page < 1 {
page = 1
}
limit, \_ := strconv.Atoi(c.DefaultQuery("limit", "20"))
if limit < 1 || limit > 100 {
limit = 20
}
pagination := &model.Pagination{
Page: page,
Limit: limit,
}
connections, err := h.service.GetUserConnections(c, userID, pagination)
if err != nil {
h.handleError(c, err)
return
}
c.JSON(http.StatusOK, connections)
}

// Helper method to handle errors
func (h *UserHandler) handleError(c *gin.Context, err error) {
requestID, \_ := c.Get("requestID")
log := h.logger.With().Str("requestID", requestID.(string)).Logger()
var appError \*apperrors.Error
if errors.As(err, &appError) {
switch appError.Type() {
case apperrors.NotFound:
log.Debug().Err(err).Msg("Resource not found")
c.JSON(http.StatusNotFound, gin.H{"error": appError.Error()})
case apperrors.BadRequest:
log.Debug().Err(err).Msg("Bad request")
c.JSON(http.StatusBadRequest, gin.H{"error": appError.Error()})
case apperrors.Unauthorized:
log.Debug().Err(err).Msg("Unauthorized")
c.JSON(http.StatusUnauthorized, gin.H{"error": appError.Error()})
case apperrors.Forbidden:
log.Debug().Err(err).Msg("Forbidden")
c.JSON(http.StatusForbidden, gin.H{"error": appError.Error()})
case apperrors.AlreadyExists:
log.Debug().Err(err).Msg("Resource already exists")
c.JSON(http.StatusConflict, gin.H{"error": appError.Error()})
default:
log.Error().Err(err).Msg("Internal server error")
c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
return
}
log.Error().Err(err).Msg("Internal server error")
c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}

// =====================================================================
// internal/content/repository/content_repo.go - Example for Developer 2
// =====================================================================
package repository

import (
"context"
"database/sql"
"encoding/json"
"errors"
"fmt"
"time"

    "github.com/google/uuid"
    "github.com/PeterM45/perfolio-api/internal/common/model"
    "github.com/PeterM45/perfolio-api/internal/platform/database"
    "github.com/PeterM45/perfolio-api/pkg/apperrors"

)

type ContentRepository interface {
// Post methods
CreatePost(ctx context.Context, post *model.Post) error
GetPostByID(ctx context.Context, id string) (*model.Post, error)
UpdatePost(ctx context.Context, post *model.Post) error
DeletePost(ctx context.Context, id string) error
GetUserPosts(ctx context.Context, userID string, limit, offset int) ([]*model.Post, error)
GetPostsByUsers(ctx context.Context, userIDs []string, limit, offset int) ([]*model.Post, error)
// Widget methods
CreateWidget(ctx context.Context, widget *model.Widget) error
GetWidgetByID(ctx context.Context, id string) (*model.Widget, error)
UpdateWidget(ctx context.Context, widget *model.Widget) error
DeleteWidget(ctx context.Context, id string) error
GetUserWidgets(ctx context.Context, userID string) ([]*model.Widget, error)
BatchUpdateWidgetPositions(ctx context.Context, updates []*model.WidgetPositionUpdate) error
}

type contentRepository struct {
db \*database.DB
}

func NewContentRepository(db \*database.DB) ContentRepository {
return &contentRepository{
db: db,
}
}

func (r *contentRepository) CreateWidget(ctx context.Context, widget *model.Widget) error {
query := `
		INSERT INTO widgets (id, user_id, widget_type, title, content, position, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
	`
// Generate UUID if not provided
if widget.ID == "" {
widget.ID = uuid.New().String()
}
now := time.Now().UTC()
widget.CreatedAt = now
widget.UpdatedAt = now
// Marshal position and content to JSON
positionJSON, err := json.Marshal(widget.Position)
if err != nil {
return fmt.Errorf("marshal widget position: %w", err)
}
var contentJSON []byte
if widget.Content != nil {
contentJSON, err = json.Marshal(widget.Content)
if err != nil {
return fmt.Errorf("marshal widget content: %w", err)
}
}
\_, err = r.db.ExecContext(ctx, query,
widget.ID,
widget.UserID,
widget.WidgetType,
widget.Title,
contentJSON,
positionJSON,
now,
)
if err != nil {
return fmt.Errorf("create widget: %w", err)
}
return nil
}

func (r *contentRepository) BatchUpdateWidgetPositions(ctx context.Context, updates []*model.WidgetPositionUpdate) error {
// Use a transaction to ensure atomic updates
tx, err := r.db.BeginTx(ctx, nil)
if err != nil {
return fmt.Errorf("begin transaction: %w", err)
}
defer tx.Rollback()
// Prepare the update statement
stmt, err := tx.PrepareContext(ctx, `
		UPDATE widgets
		SET position = $1, updated_at = $2
		WHERE id = $3 AND user_id = $4
		RETURNING id
	`)
if err != nil {
return fmt.Errorf("prepare statement: %w", err)
}
defer stmt.Close()
now := time.Now().UTC()
// Apply each update
for \_, update := range updates {
positionJSON, err := json.Marshal(update.Position)
if err != nil {
return fmt.Errorf("marshal position JSON: %w", err)
}
var id string
err = stmt.QueryRowContext(ctx,
positionJSON,
now,
update.ID,
update.UserID,
).Scan(&id)
if err != nil {
if errors.Is(err, sql.ErrNoRows) {
return apperrors.NotFound("widget", update.ID)
}
return fmt.Errorf("update widget position: %w", err)
}
}
// Commit the transaction
if err := tx.Commit(); err != nil {
return fmt.Errorf("commit transaction: %w", err)
}
return nil
}

// Other repository methods implementation...

// =====================================================================
// pkg/apperrors/errors.go
// =====================================================================
package apperrors

import "fmt"

// ErrorType is the type of an error
type ErrorType string

const (
// Error types
BadRequest ErrorType = "BAD_REQUEST"
NotFound ErrorType = "NOT_FOUND"
Internal ErrorType = "INTERNAL"
Unauthorized ErrorType = "UNAUTHORIZED"
Forbidden ErrorType = "FORBIDDEN"
AlreadyExists ErrorType = "ALREADY_EXISTS"
)

// Error is a custom error implementation
type Error struct {
errorType ErrorType
message string
}

// Error returns the error message
func (e \*Error) Error() string {
return e.message
}

// Type returns the error type
func (e \*Error) Type() ErrorType {
return e.errorType
}

// New creates a new Error
func New(errorType ErrorType, message string) \*Error {
return &Error{
errorType: errorType,
message: message,
}
}

// BadRequest creates a new bad request error
func BadRequest(message string) \*Error {
return New(BadRequest, message)
}

// NotFound creates a new not found error
func NotFound(resource, id string) \*Error {
return New(NotFound, fmt.Sprintf("%s with ID %s not found", resource, id))
}

// Internal creates a new internal error
func Internal(message string) \*Error {
return New(Internal, message)
}

// Unauthorized creates a new unauthorized error
func Unauthorized(message string) \*Error {
return New(Unauthorized, message)
}

// Forbidden creates a new forbidden error
func Forbidden(message string) \*Error {
return New(Forbidden, message)
}

// AlreadyExists creates a new already exists error
func AlreadyExists(resource, key string) \*Error {
return New(AlreadyExists, fmt.Sprintf("%s with key %s already exists", resource, key))
}

// =====================================================================
// Makefile - Modern and Efficient Build System
// =====================================================================

# Makefile

.PHONY: setup dev build test lint migrate-up migrate-down migrate-create run clean

# Environment variables

ENV ?= development
GO*FILES = $(shell find . -name "*.go" -not -path "./vendor/\_" -not -path "./test/\*")

# Development setup

setup:
go mod tidy
cp configs/config.example.yaml configs/config.yaml
docker-compose up -d postgres

# Run with hot reload using Air

dev:
air -c .air.toml

# Build the application with proper flags

build:
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/perfolio-api cmd/api/main.go

# Run the application

run:
go run cmd/api/main.go

# Run tests with coverage

test:
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run tests for specific domain

test-user:
go test -v ./internal/user/...

test-content:
go test -v ./internal/content/...

# Run integration tests

test-integration:
go test -v ./test/integration/...

# Run linter

lint:
golangci-lint run

# Database migrations

migrate-up:
migrate -path ./scripts/migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" up

migrate-down:
migrate -path ./scripts/migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" down 1

migrate-create:
@if [ -z "$(name)" ]; then echo "Usage: make migrate-create name=migration_name"; exit 1; fi
migrate create -ext sql -dir ./scripts/migrations -seq $(name)

# Generate mocks for testing

generate-mocks:
mockery --dir=./internal --output=./test/mocks --all

# Clean build artifacts

clean:
rm -rf bin/ coverage.out coverage.html
