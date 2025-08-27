# Integration Test Examples

This document shows examples of how to run the integration tests and what they verify.

## Running Integration Tests

### Using Task (Recommended)
```bash
task test:integration
```

### Using Shell Script
```bash
cd tests/
./run_integration_tests.sh
```

### Using Go Directly
```bash
cd tests/
go test -v -timeout 5m ./...
```

## Example Output

When running successfully, you should see output like this:

```
🚀 Starting Repub Integration Tests
==================================
✅ Dart CLI found: Dart SDK version: 3.8.1 (stable)
✅ curl found
🧹 Cleaning up previous test artifacts...
🧪 Running integration tests...

=== RUN   TestIntegration
2025/08/19 16:45:00 Started server with PID 12345
2025/08/19 16:45:01 ✅ Server started successfully
=== RUN   TestIntegration/publish_hello_world
2025/08/19 16:45:02 📦 Publishing package: hello_world
2025/08/19 16:45:03 ✅ Successfully published hello_world
=== RUN   TestIntegration/publish_math_utils
2025/08/19 16:45:04 📦 Publishing package: math_utils
2025/08/19 16:45:05 ✅ Successfully published math_utils
=== RUN   TestIntegration/browse_packages
2025/08/19 16:45:06 🌐 Testing web interface...
2025/08/19 16:45:06 ✅ Web interface tests passed
=== RUN   TestIntegration/install_and_use_packages
2025/08/19 16:45:07 📝 Creating test Dart project...
2025/08/19 16:45:08 📥 Installing dependencies from hosted server...
2025/08/19 16:45:09 ✅ Successfully installed packages from hosted server
--- PASS: TestIntegration (10.23s)
=== RUN   TestCleanup
--- PASS: TestCleanup (0.01s)
PASS

✅ Integration tests completed successfully!
🎉 Your pub server is working correctly with the Dart CLI!
```

## What Gets Tested

### 1. Server Startup
- Builds the server binary from source
- Starts with test configuration (SQLite DB, test port)
- Verifies server responds to HTTP requests

### 2. Package Publishing
Tests publishing real Dart packages using `dart pub publish`:

**hello_world package:**
- Simple package structure
- Basic library with functions
- Tests basic pub workflow

**math_utils package:**
- More complex package with multiple functions
- Comprehensive documentation
- Tests advanced pub features

### 3. Web Interface
Verifies the web interface works:
- Homepage loads correctly
- Package list is accessible
- Individual package pages work
- Static file serving

### 4. Package Installation
Tests the complete consume workflow:
- Creates a new Dart project
- Attempts to install published packages
- Verifies package resolution works

## Manual Testing

You can also test manually while the server is running:

### Start Server Manually
```bash
# Set required environment
export AUTH_TOKEN="your-secret-token"
export PORT="9090"
export DATABASE_URL="sqlite:test.db"

# Build and run
go build -o repub cmd/server
./repub
```

### Publish a Package
```bash
cd tests/packages/hello_world
export PUB_HOSTED_URL="http://localhost:9090"
export PUB_TOKEN_HTTP_LOCALHOST_9090="your-secret-token"
dart pub publish --force
```

### Browse Packages
```bash
# Homepage
curl http://localhost:9090

# Package list
curl http://localhost:9090/packages

# Specific package
curl http://localhost:9090/packages/hello_world
```

### API Testing
```bash
# Get package info
curl http://localhost:9090/api/packages/hello_world

# Get version info  
curl http://localhost:9090/api/packages/hello_world/versions/1.0.0
```

## Package Structure

Each test package follows standard Dart conventions:

```
packages/hello_world/
├── pubspec.yaml          # Package metadata
├── README.md             # Package documentation
├── CHANGELOG.md          # Version history
├── lib/
│   └── hello_world.dart  # Main library
└── test/
    └── hello_world_test.dart  # Unit tests
```

## Environment Variables

The integration tests set up these environment variables:

- `AUTH_TOKEN`: `integration-test-token`
- `PORT`: `19090`
- `DATABASE_URL`: `sqlite:integration_test.db`
- `STORAGE_PATH`: `./integration_test_storage`
- `PUB_HOSTED_URL`: `http://localhost:19090`
- `PUB_CACHE`: Temporary directory for test isolation

## Troubleshooting

### Common Issues

**Port Already in Use**
```bash
# Kill any process on port 19090
lsof -ti:19090 | xargs kill -9
```

**Permission Denied**
```bash
# Make script executable
chmod +x run_integration_tests.sh
```

**Dart Not Found**
```bash
# Install Dart SDK
# macOS: brew install dart
# Linux: sudo apt install dart
# Or download from dart.dev
```

### Debug Mode

To run with more verbose output:

```bash
cd tests/
go test -v -timeout 10m ./... 2>&1 | tee integration.log
```

This will save all output to `integration.log` for debugging.