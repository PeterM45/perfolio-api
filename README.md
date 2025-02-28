# Perfolio API

Backend API service for Perfolio

## Overview

Perfolio API

## Features

- **User Management**: Authentication, profiles, connections
- **Content Management**: Posts, feeds, reactions
- **Widget System**: Customizable profile layout with draggable widgets
- **Concurrent Operations**: Optimized for high-volume user interactions

## Tech Stack

- **Language**: Go 1.22+
- **Database**: PostgreSQL
- **Authentication**: Clerk
- **Frontend**: Next.js (separate repo)

## Quick Start

### Prerequisites

- Go 1.22+
- Docker and Docker Compose
- Make

### Setup

1. Clone the repository:

```bash
git clone git@github.com:PeterM45/perfolio-api.git
cd perfolio-api
```

2. Set up the development environment:

```bash
make setup
```

3. Start the database:

```bash
docker-compose up -d
```

4. Run database migrations:

```bash
make migrate-up
```

5. Start the server with hot reload:

```bash
make dev
```

The API will be available at `http://localhost:8080`.

## Project Structure

```
perfolio-api/
├── cmd/               # Application entry points
├── internal/          # Private application code
│   ├── user/          # User domain (Developer 1)
│   ├── content/       # Content domain (Developer 2)
│   ├── common/        # Shared code
│   └── platform/      # Infrastructure
├── pkg/               # Public libraries
├── scripts/           # Migration scripts and utilities
├── docs/              # Detailed documentation
└── test/              # Integration tests
```

## Documentation

Detailed documentation is available in the `/docs` directory:

- [Developer Guide](./docs/DEV_GUIDE.MD) - Comprehensive implementation guide
- [Database Migrations](./docs/DB_MIGRATIONS.md) - Database schema and migrations
- [Example Code](./docs/EXAMPLE_CODE.md) - Implementation examples for key features
