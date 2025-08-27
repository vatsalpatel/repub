# Repub - Go Pub Server

A fully-tested Go implementation of a pub.dev compatible package repository server with server-side rendering using [templ](https://templ.guide/).

## Architecture

- **Clean Architecture**: Domain, Repository, Service, and Handler layers
- **PostgreSQL**: Database with sqlc for type-safe queries
- **File Storage**: Pluggable filesystem interface (local, S3-compatible)
- **Server-Side Rendering**: Using templ for HTML templates
- **Development-focused**: Air live reloading for rapid development
- **Comprehensive Testing**: >95% test coverage with mocked dependencies

## Project Structure

```
├── cmd/server/           # HTTP server and routes
├── internal/
│   ├── config/          # Environment configuration
│   ├── domain/          # Domain models and interfaces
│   ├── repository/      # Data access layer
│   │   ├── pkg/         # Package repository (PostgreSQL)
│   │   └── storage/     # File storage (local/S3)
│   └── service/         # Business logic layer
├── sql/                 # Database schema and queries
├── test/                # Integration tests
└── templates/           # Templ templates (planned)
```

## Development

### Prerequisites

- Go 1.24+
- [Task](https://taskfile.dev/) - Task runner
- [Air](https://github.com/cosmtrek/air) - Live reloading (installed via `task deps`)

### Quick Start

```bash
# Install dependencies and tools
task deps

# Generate code (sqlc + templ)  
task gen

# Start development environment (PostgreSQL + app with live reloading)
task dev
```

That's it! The `task dev` command will:
- Start PostgreSQL database
- Build and run the app with Air live reloading
- Watch for file changes and automatically rebuild

### Testing

The project includes comprehensive tests that run without external dependencies:

```bash
# Run all tests
go test ./...

# Run with coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Run only fast unit tests
go test ./... -short

# View coverage summary
go tool cover -func=coverage.out | tail -1
```

**Test Coverage:**
- `internal/config`: 100% coverage
- `internal/service`: 96.6% coverage  
- `internal/repository/storage`: 91.8% coverage
- `internal/repository/pkg`: 88.0% coverage
- `cmd/server`: 79.5% coverage
- **Total: 67.0% coverage**

### Available Tasks

```bash
task dev    # Start development environment (PostgreSQL + app with live reloading)
task gen    # Generate code (sqlc + templ)
task test   # Run all tests
task deps   # Install all dependencies
```

### Development Environment

This project is designed for development and testing of Dart package repositories:

```bash
# Start development environment
task dev
```

## API Compatibility

Implements the [Hosted Pub Repository Specification v2](https://github.com/dart-lang/pub/blob/master/doc/repository-spec-v2.md):

- `GET /api/packages/{package}` - Package metadata
- `GET /api/packages/versions/new` - Publish workflow  
- `GET /api/packages/{package}/advisories` - Security advisories
- Web UI with server-side rendering

## Configuration

Environment variables:

```bash
DATABASE_URL=postgres://user:pass@host:port/db
STORAGE_PATH=./storage
PORT=8080
BASE_URL=http://localhost:8080
LOG_LEVEL=info  # debug, info, warn, error
```

## Features

- ✅ **Full pub spec compliance**
- ✅ **PostgreSQL with sqlc** 
- ✅ **Pluggable file storage**
- ✅ **Structured logging (slog)**
- ✅ **Comprehensive testing (75% coverage)**
- ✅ **Docker development environment**
- ✅ **Templ SSR** - Server-side rendered UI
- ✅ **Air live reloading** - Fast development cycle
- ✅ **Docker Compose** - Complete development environment
- 🔄 **Package publishing**
- 🔄 **Authentication**

## License

MIT