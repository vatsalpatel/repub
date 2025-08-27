# Claude Code Project Context

## Project Overview

**Repub** - A fully-tested Go implementation of a pub.dev compatible package repository server with server-side rendering using templ.guide.

## Architecture

- **Language**: Go 1.22+
- **Database**: PostgreSQL with sqlc for type-safe queries
- **Storage**: Pluggable filesystem interface (local storage implemented)
- **Frontend**: Server-side rendering with templ templates
- **Build**: Ko for optimized container builds with distroless images
- **Development**: Air for live reloading + Docker Compose

## Project Structure

```
├── cmd/server/           # HTTP server, handlers, and routes
├── internal/
│   ├── config/          # Environment configuration
│   ├── db/              # Generated sqlc code
│   ├── domain/          # Domain models and API types
│   ├── repository/      # Data access layer
│   │   ├── pkg/         # Package repository (PostgreSQL)
│   │   └── storage/     # File storage (local filesystem)
│   ├── service/         # Business logic layer
│   └── testutil/        # Shared test utilities
├── sql/                 # Database schema and queries for sqlc
├── test/                # Integration tests
└── web/templates/       # Templ templates for SSR
```

## Development Workflow

### Essential Commands

```bash
task deps    # Install all dependencies and tools
task gen     # Generate code (sqlc + templ)
task dev     # Start development environment (PostgreSQL + app with live reloading)
task test    # Run all tests
```

### Development Environment

- **`task dev`** starts a full Docker Compose stack:
  - PostgreSQL database with schema auto-loaded
  - Go app with Air live reloading
  - Watches Go files, templ templates, SQL files
  - Debug logging enabled

- **Air Configuration** (`.air.toml`):
  - Builds to `tmp/main`
  - Excludes test files and build artifacts
  - 1 second delay for stability

## Architecture Patterns

### Clean Architecture Layers

1. **Domain Layer** (`internal/domain/`): Core business models and interfaces
2. **Repository Layer** (`internal/repository/`): Data access abstractions
3. **Service Layer** (`internal/service/`): Business logic orchestration
4. **Handler Layer** (`cmd/server/`): HTTP request/response handling

### Key Design Decisions

- **Routes only use services, not repositories** - Clean separation of concerns
- **Interface-based design** - All dependencies are injected via interfaces
- **File system abstraction** - Uses `fs.FS` interface for testability
- **Structured logging** - slog throughout with configurable levels
- **Error handling** - All errors properly checked and handled (passes golangci-lint)

## Testing Strategy

### Test Coverage: 75% (core application code)

- **Unit Tests**: All components with comprehensive mocking
- **Integration Tests**: SQLite-based testing of PostgreSQL functionality
- **Table-driven Tests**: Consolidated test files per component
- **Error Path Coverage**: Comprehensive error handling testing
- **No External Dependencies**: All tests run without Docker/database

### Test Utilities

- **`testutil.SetupSQLiteDB()`** - Shared SQLite database setup for integration tests
- **`testFS`** - In-memory filesystem using `testing/fstest.MapFS`
- **Mock repositories and services** - Comprehensive mocking for isolated testing

## Code Generation

### SQLC Configuration (`sqlc.yaml`)

- Engine: PostgreSQL
- Schema: `sql/schema.sql`
- Queries: `sql/queries.sql` 
- Output: `internal/db/` package
- JSON tags enabled for API responses

### Templ Templates

- Location: `web/templates/*.templ`
- Generated: `web/templates/*_templ.go`
- Components: Base layout, package lists, package details, version details

## API Specification

Implements [Hosted Pub Repository Specification v2](https://github.com/dart-lang/pub/blob/master/doc/repository-spec-v2.md):

### API Endpoints
- `GET /api/packages/{package}` - Package metadata (JSON)
- `GET /api/packages/versions/new` - Publish workflow initiation
- `GET /api/packages/{package}/advisories` - Security advisories

### Web UI Endpoints  
- `GET /` - Homepage
- `GET /packages` - Package listing
- `GET /packages/{package}` - Package details
- `GET /packages/{package}/versions/{version}` - Version details

## Environment Configuration

```bash
DATABASE_URL=postgres://repub:repub@localhost:5432/repub?sslmode=disable
STORAGE_PATH=./storage
PORT=8080
BASE_URL=http://localhost:8080
LOG_LEVEL=debug  # debug, info, warn, error
```

## Dependencies and Tools

### Runtime Dependencies
- `github.com/jackc/pgx/v5` - PostgreSQL driver
- `github.com/a-h/templ` - Template engine
- `modernc.org/sqlite` - SQLite for testing

### Development Tools
- `github.com/sqlc-dev/sqlc` - SQL code generation
- `github.com/google/ko` - Container builds
- `github.com/cosmtrek/air` - Live reloading
- `github.com/a-h/templ/cmd/templ` - Template generation

## Development Environment

### Docker Compose
```bash
task dev  # Starts development environment with live reloading
```

- **Development**: `docker-compose --profile dev up --build` - Full development stack with Air live reloading

## Quality Metrics

- ✅ **0 lint issues** (golangci-lint + go vet)
- ✅ **41 tests passing** (0 failures)
- ✅ **75% test coverage** (core application code)
- ✅ **Clean architecture** with proper separation of concerns
- ✅ **Comprehensive error handling** throughout codebase

## Current Status

The project is a **development-focused implementation** with:
- Full pub.dev specification compliance for core endpoints
- Server-side rendered web interface
- Live reloading development environment
- Comprehensive testing and quality assurance
- Docker development environment with live reloading

### Remaining Features (Future Work)
- Package publishing workflow implementation
- Authentication and authorization system
- Advanced security advisories management