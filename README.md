Backend API service for Perfolio - a professional networking platform with customizable profiles, posts, feeds, and connections.

## Overview

Perfolio API is a high-performance Go backend designed to power a LinkedIn-like social network with a personal touch. It features a customizable profile system with draggable widgets, allowing users to create unique professional profiles. Built with scalability in mind, it provides excellent cost-to-performance ratio for handling concurrent user operations.

## Project Structure

The codebase is organized using Domain-Driven Design principles:

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
│   │   ├── interfaces/# Domain-specific interfaces
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

## Features

- **User Management**: Authentication, profiles, connections, follow relationships
- **Content Management**: Posts, feeds, reactions, hashtags
- **Widget System**: Customizable profile layout with draggable widgets
- **High Performance**: Caching strategies for frequently accessed data
- **Secure Authentication**: JWT-based authentication with refresh tokens

## Getting Started

### Prerequisites

- Go 1.22+
- PostgreSQL
- Docker & Docker Compose (optional)
- Redis (optional, for distributed caching)

### Setup

1. Clone the repository:

```bash
git clone https://github.com/PeterM45/perfolio-api.git
cd perfolio-api
```

2. Set up the development environment:

```bash
make setup
```

3. Configure the application by editing `configs/config.yaml`

4. Start the database:

```bash
make start-db
```

5. Create the database and run migrations:

```bash
make create-db
make migrate-up
```

6. Start the server with hot reload:

```bash
make dev
```

The API will be available at `http://localhost:8080`.

## Development

### Common Commands

```bash
# Go into PSQL
psql -h localhost -p 5432 -U postgres -d perfolio

# Run tests
make test

# Lint code
make lint

# Create a new migration
make migrate-create name=my_migration

# Build for production
make build
```

### API Documentation

The API follows RESTful principles with these main endpoints:

#### Authentication

- `POST /api/v1/auth/register` - Register a new user
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Refresh authentication token
- `POST /api/v1/auth/logout` - User logout
- `GET /api/v1/auth/me` - Get current authenticated user

#### Users

- `GET /api/v1/users/:id` - Get user profile
- `PUT /api/v1/users/:id` - Update user profile
- `GET /api/v1/users/search` - Search users
- `POST /api/v1/users/:id/follow` - Follow/unfollow a user
- `GET /api/v1/users/:id/followers` - Get user's followers
- `GET /api/v1/users/:id/following` - Get users being followed
- `GET /api/v1/users/:id/stats` - Get profile statistics

#### Posts

- `GET /api/v1/posts/:id` - Get a post
- `POST /api/v1/posts` - Create a post
- `PUT /api/v1/posts/:id` - Update a post
- `DELETE /api/v1/posts/:id` - Delete a post
- `GET /api/v1/posts/feed` - Get user feed
- `GET /api/v1/posts/user/:userId` - Get user's posts

#### Widgets

- `GET /api/v1/widgets/:id` - Get a specific widget
- `GET /api/v1/widgets/user/:userId` - Get user widgets
- `POST /api/v1/widgets` - Create a widget
- `PUT /api/v1/widgets/:id` - Update a widget
- `DELETE /api/v1/widgets/:id` - Delete a widget
- `POST /api/v1/widgets/batch-update` - Update multiple widgets positions

## Deployment

### Docker

Build and run the Docker container:

```bash
# Build the Docker image
make docker-build

# Run the Docker container
make docker-run
```

### Environment Variables

The application can be configured using environment variables instead of the config file:

- `SERVER_PORT` - HTTP server port
- `DATABASE_HOST` - PostgreSQL host
- `DATABASE_PORT` - PostgreSQL port
- `DATABASE_USER` - PostgreSQL user
- `DATABASE_PASSWORD` - PostgreSQL password
- `DATABASE_NAME` - PostgreSQL database name
- `DATABASE_SSL_MODE` - PostgreSQL SSL mode
- `AUTH_JWT_SECRET` - Secret key for JWT tokens
- `CACHE_TYPE` - Cache type (memory or redis)
- `CACHE_REDIS_URL` - Redis URL (if using Redis cache)
- `LOG_LEVEL` - Logging level (debug, info, warn, error)

## Architecture

The application follows domain-driven design principles with clear separation of concerns:

- **Handlers** - Handle HTTP requests and responses
- **Services** - Implement business logic and orchestrate repository calls
- **Repositories** - Manage data access and persistence
- **Models** - Define domain entities and DTOs
- **Interfaces** - Define contracts between layers within each domain

Each domain (user, post, widget) has its own set of interfaces, implementations, and handlers to maintain separation and enable independent development.

For more detailed implementation guidelines, refer to the [Developer Guide](docs/DEV_GUIDE.MD).

```

```
