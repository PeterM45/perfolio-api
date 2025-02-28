# Perfolio API - Developer Guide

A high-performance Go backend for our professional networking platform with profiles, posts, feeds, and customizable widgets.

## Project Overview

Perfolio API powers our LinkedIn-like social networking platform, designed to:

- Handle high concurrent operations (profile edits, connections, post interactions)
- Deliver excellent performance for feeds and real-time features
- Support scalable customizable widget system for profiles
- Maintain low infrastructure costs with efficient resource utilization

## Architecture

The API follows Clean Architecture principles with clear domain boundaries:

```
┌───────────────────┐     ┌──────────────────┐     ┌───────────────────┐
│     Handler       │     │     Service      │     │    Repository     │
│  (HTTP Layer)     │ ──► │ (Business Logic) │ ──► │  (Data Access)    │
└───────────────────┘     └──────────────────┘     └───────────────────┘
```

### Key Design Principles

1. **Domain-Driven Design**: Organized by business domain (user, content) not technical concerns
2. **Interface Segregation**: Each layer communicates through interfaces for testing and flexibility
3. **Single Responsibility**: Each component has one reason to change
4. **Dependency Injection**: Dependencies provided via constructors, not created inside components
5. **Explicit Error Handling**: Structured errors with context and appropriate HTTP mappings

## Development Domains

The codebase is organized into two primary domains for parallel development:

### User Domain (Developer 1)

- User authentication and sessions
- Profile management
- Connection/follow system
- User discovery & search

### Content Domain (Developer 2)

- Posts creation and timeline
- Feed algorithms and delivery
- Widget system (creation, positioning, content)
- Reactions and engagement

## Getting Started

### Prerequisites

- Go 1.22+
- Docker and Docker Compose
- Make
- PostgreSQL client (optional)

### Setup Environment

1. Clone the repository:

```bash
git clone git@github.com:PeterM45/perfolio-api.git
cd perfolio-api
```

2. Set up local development environment:

```bash
make setup
```

3. Start development services:

```bash
make dev
```

4. Run migrations:

```bash
make migrate-up
```

5. Start the server with hot reload:

```bash
make run
```

## Development Workflow

### Adding New Features

1. **Define models first**:

   - Start with domain models in `internal/common/model`
   - Define clear request/response structs
   - Add proper validation tags

2. **Repository implementation**:

   - Implement data access in relevant repository
   - Use parameterized queries with proper context
   - Handle SQL errors with domain-specific errors

3. **Service implementation**:

   - Add business logic in service layer
   - Validate inputs before calling repositories
   - Transform data between layers as needed

4. **Handler implementation**:
   - Parse and validate requests
   - Call appropriate service methods
   - Map errors to appropriate HTTP responses

### Example: Creating a User Widget (Content Domain)

```go
// 1. Define model
type CreateWidgetRequest struct {
    Title      string          `json:"title" validate:"required,min=3,max=50"`
    WidgetType string          `json:"widget_type" validate:"required,oneof=profile_info recent_posts connections calendar custom"`
    Content    json.RawMessage `json:"content,omitempty"`
    Position   WidgetPosition  `json:"position" validate:"required"`
}

type WidgetPosition struct {
    X      int `json:"x" validate:"required,min=0,max=11"`
    Y      int `json:"y" validate:"required,min=0"`
    Width  int `json:"width" validate:"required,min=1,max=12"`
    Height int `json:"height" validate:"required,min=1"`
}

// 2. Repository method
func (r *contentRepository) CreateWidget(ctx context.Context, widget *model.Widget) error {
    query := `
        INSERT INTO widgets (id, user_id, widget_type, title, content, position, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `

    // Implementation with proper error handling
}

// 3. Service method
func (s *contentService) CreateWidget(ctx context.Context, userID string, req *model.CreateWidgetRequest) (*model.Widget, error) {
    // Validation, business logic, repository call
}

// 4. Handler method
func (h *ContentHandler) CreateWidget(c *gin.Context) {
    // Request binding, validation, service call, response
}
```

## Widget System Implementation

The widget system requires special care for concurrent modifications:

### Widget Grid Management

1. **Optimistic Concurrency Control**:

   - Include version field in layout updates
   - Detect and handle concurrent modification conflicts

2. **Batch Updates**:

   - Use transactions for grid rearrangements
   - Implement specialized endpoint for bulk position updates

3. **Grid Collision Detection**:
   - Implement server-side validation for widget overlaps
   - Use Go's concurrency primitives for parallel validation

## Integration Between Domains

When the content domain needs user data:

```go
// In content service
type contentService struct {
    repo        repository.ContentRepository
    userService user.Service  // Interface from user domain
}

// Using the interface
func (s *contentService) GetUserFeed(ctx context.Context, userID string) ([]*model.Post, error) {
    // Verify user exists
    _, err := s.userService.GetUserByID(ctx, userID)
    if err != nil {
        return nil, err
    }

    // Get connections
    connections, err := s.userService.GetUserConnections(ctx, userID)
    if err != nil {
        return nil, err
    }

    // Get posts from connections
    // Implementation
}
```

## Performance Considerations

### Database Access

1. **Connection Pooling**:

   ```go
   db, err := sql.Open("postgres", connString)
   if err != nil {
      return nil, err
   }

   // Set connection pool parameters
   db.SetMaxOpenConns(25)
   db.SetMaxIdleConns(10)
   db.SetConnMaxLifetime(5 * time.Minute)
   ```

2. **Query Optimization**:
   - Use `EXPLAIN ANALYZE` to verify query performance
   - Create proper indexes (already in migrations)
   - Consider materialized views for complex queries

### Caching Strategy

Start with in-memory caching for simplicity:

```go
type cacheService struct {
    cache map[string]cacheItem
    mutex sync.RWMutex
}

type cacheItem struct {
    value      interface{}
    expiration time.Time
}
```

Evolve to Redis when needed:

```go
func NewRedisCache(redisClient *redis.Client) Cache {
    return &redisCache{
        client: redisClient,
    }
}
```

## Testing Guidelines

1. **Unit Tests**: Test each component in isolation

   ```go
   func TestUserService_GetUserByID(t *testing.T) {
       // Setup mock repository
       mockRepo := mocks.NewMockUserRepository(t)
       mockRepo.EXPECT().GetByID(mock.Anything, "user123").Return(&model.User{ID: "user123"}, nil)

       // Create service with mock
       service := NewUserService(mockRepo)

       // Call method and assert
       user, err := service.GetUserByID(context.Background(), "user123")
       require.NoError(t, err)
       require.Equal(t, "user123", user.ID)
   }
   ```

2. **Integration Tests**: Test interaction between components
3. **API Tests**: Test endpoints end-to-end

## Database Schema Evolution

All database changes should be versioned:

```bash
# Create a new migration
make migrate-create name=add_widget_permissions

# Apply migrations
make migrate-up
```

## Deployment Strategy

For cost-efficiency, deploy as a single binary initially:

```bash
# Build optimized binary
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/perfolio-api cmd/api/main.go

# Deploy to AWS with proper configurations
```

Consider containerized deployment when traffic increases:

```yaml
# Dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o perfolio-api cmd/api/main.go

FROM alpine:latest
COPY --from=builder /app/perfolio-api /perfolio-api
ENTRYPOINT ["/perfolio-api"]
```

## Collaboration Guidelines

1. Work in feature branches
2. Use clear, descriptive branch names (e.g., `feature/user-widgets`)
3. Write meaningful commit messages
4. Keep pull requests focused and manageable in size
5. Implement comprehensive test coverage
6. Document public APIs using godoc format
7. Update API documentation when adding endpoints

## Monitoring & Observability

Include request ID in all logs for traceability:

```go
// Middleware
func RequestIDMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := c.GetHeader("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }

        c.Set("requestID", requestID)
        c.Header("X-Request-ID", requestID)

        c.Next()
    }
}

// In handlers
func (h *UserHandler) GetUserByID(c *gin.Context) {
    requestID, _ := c.Get("requestID")
    logger := h.logger.With().Str("requestID", requestID.(string)).Logger()

    // Handler implementation with contextual logging
}
```
