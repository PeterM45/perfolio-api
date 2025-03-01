# Perfolio API - Developer Guide

A high-performance Go backend for the Perfolio professional networking platform.

## Project Overview

Perfolio API is a scalable backend service designed to power our LinkedIn-like platform with:

- User profiles and connections
- Post creation and feed generation
- Customizable profile widgets
- High concurrency support

## Repository Structure

```
perfolio-api/
├── cmd/               # Application entry points
│   └── api/           # API server
│       ├── app/       # Application setup
│       └── main.go    # Main entry point
├── configs/           # Configuration files
├── internal/          # Private application code
│   ├── user/          # User domain
│   │   ├── handler/   # HTTP handlers
│   │   ├── service/   # Business logic
│   │   └── repository/# Data access
│   ├── content/       # Content domain
│   │   ├── handler/   # HTTP handlers
│   │   ├── service/   # Business logic
│   │   └── repository/# Data access
│   ├── common/        # Shared code
│   │   ├── config/    # Configuration
│   │   ├── middleware/# HTTP middleware
│   │   └── model/     # Domain models
│   └── platform/      # Infrastructure
│       ├── database/  # Database connections
│       └── cache/     # Caching
├── pkg/               # Public libraries
│   ├── logger/        # Logging package
│   ├── apperrors/     # Error handling
│   └── validator/     # Request validation
├── scripts/           # Scripts and migrations
│   └── migrations/    # SQL migration files
└── test/              # Test files
```

## Domain Separation

The codebase is organized into two primary domains to enable parallel development:

### User Domain (Developer 1)

| Component  | Location                                | Responsibility       |
| ---------- | --------------------------------------- | -------------------- |
| Models     | `internal/common/model/user.go`         | User entities & DTOs |
| Repository | `internal/user/repository/user_repo.go` | Database operations  |
| Service    | `internal/user/service/user_service.go` | Business logic       |
| Handler    | `internal/user/handler/user_handler.go` | HTTP endpoints       |

Handles:

- User authentication and profiles
- Connection/follow relationships
- Profile data management

### Content Domain (Developer 2)

| Component  | Location                                                      | Responsibility          |
| ---------- | ------------------------------------------------------------- | ----------------------- |
| Models     | `internal/common/model/post.go, widget.go`                    | Content entities & DTOs |
| Repository | `internal/content/repository/post_repo.go, widget_repo.go`    | Database operations     |
| Service    | `internal/content/service/post_service.go, widget_service.go` | Business logic          |
| Handler    | `internal/content/handler/post_handler.go, widget_handler.go` | HTTP endpoints          |

Handles:

- Posts and feed generation
- Content reactions
- Profile widgets and layout

## Getting Started

### Prerequisites

- Go 1.22+
- Docker and Docker Compose
- Make

### Initial Setup

1. Clone the repository:

```bash
git clone github.com/PeterM45/perfolio-api.git
cd perfolio-api
```

2. Install dependencies and tools:

```bash
make setup
```

3. Configure the application:

```bash
# Edit configs/config.yaml with appropriate values
```

4. Start the database:

```bash
make start-db
```

5. Run database migrations:

```bash
make migrate-up
```

### Development Workflow

Start the API with hot reload:

```bash
make dev
```

The server will automatically reload when you make changes to Go files.

Run tests:

```bash
make test
```

Run linter:

```bash
make lint
```

## Adding New Functionality

### Development Flow

1. **Define the model** in `internal/common/model/`
2. **Implement repository methods** in the appropriate repository
3. **Add business logic** in the corresponding service
4. **Create HTTP handlers** for API endpoints
5. **Register routes** in the domain's handler registration

### Example: Implementing a New Feature

Let's walk through adding a "save post" feature:

1. **Add model definitions**:

```go
// internal/common/model/post.go
type SavedPost struct {
    UserID    string    `json:"userId"`
    PostID    string    `json:"postId"`
    CreatedAt time.Time `json:"createdAt"`
}

type SavePostRequest struct {
    PostID string `json:"postId" validate:"required"`
}
```

2. **Add repository method**:

```go
// internal/content/repository/post_repo.go
func (r *postRepository) SavePost(ctx context.Context, userID, postID string) error {
    query := `
        INSERT INTO saved_posts (user_id, post_id, created_at)
        VALUES ($1, $2, $3)
        ON CONFLICT (user_id, post_id) DO NOTHING
    `

    _, err := r.db.ExecContext(ctx, query, userID, postID, time.Now().UTC())
    if err != nil {
        return fmt.Errorf("save post: %w", err)
    }

    return nil
}
```

3. **Implement service method**:

```go
// internal/content/service/post_service.go
func (s *postService) SavePost(ctx context.Context, userID string, req *model.SavePostRequest) error {
    // Validate request
    if err := s.validator.Validate(req); err != nil {
        return apperrors.BadRequest(err.Error())
    }

    // Verify post exists
    if _, err := s.repo.GetByID(ctx, req.PostID); err != nil {
        return err
    }

    // Save post
    return s.repo.SavePost(ctx, userID, req.PostID)
}
```

4. **Create handler method**:

```go
// internal/content/handler/post_handler.go
func (h *PostHandler) SavePost(c *gin.Context) {
    // Get authenticated user
    userID, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    var req model.SavePostRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    h.logger.Debug().
        Str("user_id", userID.(string)).
        Str("post_id", req.PostID).
        Msg("Saving post")

    err := h.service.SavePost(c, userID.(string), &req)
    if err != nil {
        h.handleError(c, err)
        return
    }

    c.JSON(http.StatusOK, gin.H{"success": true})
}
```

5. **Register the route**:

```go
// internal/content/handler/post_handler.go (in RegisterRoutes method)
func (h *PostHandler) RegisterRoutes(router *gin.RouterGroup) {
    posts := router.Group("/posts")
    {
        // Existing routes...
        posts.POST("/save", h.SavePost)
    }
}
```

## Error Handling

Always use the custom error types from `pkg/apperrors` for consistent error responses:

```go
// Return application-specific errors
if user == nil {
    return nil, apperrors.NotFound("user", id)
}

if !authorized {
    return nil, apperrors.Forbidden("not authorized to update this post")
}

if !valid {
    return nil, apperrors.BadRequest("invalid request: " + details)
}
```

In handlers, use the `handleError` helper to convert application errors to appropriate HTTP responses.

## Database Operations

### Connection Management

The database connection is established in `internal/platform/database/database.go`:

```go
db, err := sql.Open("pgx", connStr)
if err != nil {
    return nil, fmt.Errorf("open database connection: %w", err)
}

// Configure connection pool
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(10)
db.SetConnMaxLifetime(5 * time.Minute)
```

### Migrations

Database migrations are versioned and located in `scripts/migrations/`:

```bash
# Create a new migration
make migrate-create name=add_saved_posts_table

# Apply all migrations
make migrate-up

# Roll back the latest migration
make migrate-down
```

## Caching Strategy

The application uses a tiered caching approach:

1. **In-Memory Cache**: Default for development and small deployments

   - Implementation: `internal/platform/cache/cache.go`
   - Automatic cleanup of expired items

2. **Redis Cache**: For production or larger deployments
   - Implementation: `internal/platform/cache/redis.go`
   - Configured via `configs/config.yaml`

Services use the cache interface to store frequently accessed data:

```go
// Check cache first
cacheKey := fmt.Sprintf("user:%s", id)
if cachedUser, found := s.cache.Get(cacheKey); found {
    return cachedUser.(*model.User), nil
}

// Get from database
user, err := s.repo.GetByID(ctx, id)
if err != nil {
    return nil, err
}

// Store in cache for future requests
s.cache.Set(cacheKey, user, 5*time.Minute)
```

## Authentication

Authentication is handled via Clerk integration:

- JWT token validation in `internal/common/middleware/auth.go`
- Webhook processing for Clerk events
- User ID extraction and context setting

Protected endpoints should use the authentication middleware:

```go
// In route registration
router.POST("/api/v1/posts", authMiddleware.Authenticate(), postHandler.CreatePost)
```

## Testing

### Unit Tests

Use mock interfaces for unit testing services:

```go
func TestUserService_GetUserByID(t *testing.T) {
    // Setup mock repository
    mockRepo := mocks.NewMockUserRepository(t)
    mockRepo.On("GetByID", mock.Anything, "user123").Return(&model.User{ID: "user123"}, nil)

    mockCache := mocks.NewMockCache(t)
    mockCache.On("Get", "user:user123").Return(nil, false)
    mockCache.On("Set", "user:user123", mock.Anything, mock.Anything).Return()

    // Create service with mocks
    service := NewUserService(mockRepo, mockCache, logger.NewZapLogger("test"))

    // Call method and assert
    user, err := service.GetUserByID(context.Background(), "user123")
    require.NoError(t, err)
    require.Equal(t, "user123", user.ID)

    // Verify expectations
    mockRepo.AssertExpectations(t)
    mockCache.AssertExpectations(t)
}
```

### API Tests

Use `httptest` package for testing API endpoints:

```go
func TestUserHandler_GetUserByID(t *testing.T) {
    // Setup mock service
    mockService := mocks.NewMockUserService(t)
    mockService.On("GetUserByID", mock.Anything, "user123").Return(&model.User{ID: "user123"}, nil)

    // Create handler with mock
    handler := NewUserHandler(mockService, logger.NewZapLogger("test"))

    // Setup router
    router := gin.New()
    router.GET("/users/:id", handler.GetUserByID)

    // Create request
    req := httptest.NewRequest("GET", "/users/user123", nil)
    w := httptest.NewRecorder()

    // Serve request
    router.ServeHTTP(w, req)

    // Assert response
    require.Equal(t, http.StatusOK, w.Code)

    var response model.User
    err := json.Unmarshal(w.Body.Bytes(), &response)
    require.NoError(t, err)
    require.Equal(t, "user123", response.ID)

    // Verify expectations
    mockService.AssertExpectations(t)
}
```

## Deployment

### Building for Production

```bash
# Build optimized binary
make build

# The binary will be generated in bin/perfolio-api
```

### Docker Deployment

```bash
# Build Docker image
make docker-build

# Run Docker container
make docker-run
```

### Container Configuration

Environment variables for the Perfolio API container:

```
SERVER_PORT=8080
DATABASE_HOST=postgres
DATABASE_PORT=5432
DATABASE_USER=postgres
DATABASE_PASSWORD=postgres
DATABASE_NAME=perfolio
DATABASE_SSL_MODE=disable
CACHE_TYPE=redis
CACHE_REDIS_URL=redis://redis:6379
LOG_LEVEL=info
```

## Performance Considerations

### Database Optimization

- Use proper indexes for frequently queried columns
- Keep transactions short and focused
- Use EXPLAIN ANALYZE to verify query performance

### Caching Recommendations

- Cache items with appropriate TTL based on update frequency
- Invalidate cache when data changes
- Use batch operations for Redis when possible

### Request Processing

- Use appropriate timeout contexts for external calls
- Implement rate limiting for public endpoints
- Add pagination for list endpoints

## Collaboration Guidelines

1. **Branch naming**:

   - Feature branches: `feature/user-profiles`
   - Bugfix branches: `fix/post-creation-error`
   - Release branches: `release/v1.0.0`

2. **Pull requests**:

   - Keep PRs focused on a single feature or fix
   - Include relevant tests
   - Update documentation as needed
   - Reference issues in PR description

3. **Code style**:

   - Follow Go standard naming conventions
   - Run linter before committing: `make lint`
   - Add comments for exported functions and types
   - Keep functions small and focused

4. **Commit messages**:
   - Use present tense, imperative style
   - Start with a capital letter
   - Keep first line under 50 characters
   - Include issue number if applicable

## Common Issues and Solutions

### Database Connection Issues

If you encounter database connection problems:

1. Verify PostgreSQL is running: `docker ps`
2. Check connection details in config file
3. Ensure migrations have been applied: `make migrate-up`

### Authentication Problems

If authentication is failing:

1. Verify Clerk secret key in configuration
2. Check JWT token format in Authorization header
3. Ensure token hasn't expired

### Hot Reload Not Working

If changes aren't being detected:

1. Check if the file is excluded in `.air.toml`
2. Verify Air is running properly
3. Try restarting the Air process: `make dev`

## API Routes Reference

| Method | Path                         | Description             | Auth Required |
| ------ | ---------------------------- | ----------------------- | ------------- |
| GET    | /api/v1/users/:id            | Get user profile        | No            |
| PUT    | /api/v1/users/:id            | Update user profile     | Yes           |
| GET    | /api/v1/users/search         | Search users            | No            |
| POST   | /api/v1/users/:id/follow     | Follow/unfollow a user  | Yes           |
| GET    | /api/v1/posts/:id            | Get a post              | No            |
| POST   | /api/v1/posts                | Create a post           | Yes           |
| PUT    | /api/v1/posts/:id            | Update a post           | Yes           |
| DELETE | /api/v1/posts/:id            | Delete a post           | Yes           |
| GET    | /api/v1/posts/feed           | Get user feed           | Yes           |
| GET    | /api/v1/widgets/user/:userId | Get user widgets        | No            |
| POST   | /api/v1/widgets              | Create a widget         | Yes           |
| PUT    | /api/v1/widgets/:id          | Update a widget         | Yes           |
| POST   | /api/v1/widgets/batch-update | Update multiple widgets | Yes           |
