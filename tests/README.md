# Integration Tests

This directory contains integration tests for the Repub server that use the real Dart CLI to test end-to-end functionality.

## Structure

```
tests/
├── integration_test.go          # Main integration test file
├── run_integration_tests.sh     # Helper script to run tests
├── packages/                    # Example Dart packages for testing
│   ├── hello_world/            # Simple hello world package
│   └── math_utils/             # Math utility package with multiple functions
└── README.md                   # This file
```

## Prerequisites

1. **Dart SDK**: Install from [dart.dev/get-dart](https://dart.dev/get-dart)
2. **curl**: Should be available on most systems
3. **Go**: Required to build and run the server

## Running Tests

### Quick Start

```bash
# Run the integration tests
cd tests/
./run_integration_tests.sh
```

### Manual Testing

```bash
# Run just the Go integration tests
go test -v ./...

# Or run with more detailed output
go test -v -timeout 5m ./...
```

## What Gets Tested

The integration tests verify:

1. **Server Startup**: Server starts correctly with test configuration
2. **Package Publishing**: Publishing packages using `dart pub publish`
3. **Web Interface**: Homepage, package list, and package detail pages
4. **Package Installation**: Installing published packages with `dart pub get`
5. **End-to-End Flow**: Complete workflow from publish to consume

## Test Packages

### hello_world (v1.0.0)
- Simple hello world functionality
- Basic package structure
- Tests basic pub publish workflow

### math_utils (v1.2.0) 
- Mathematical utility functions
- More complex package with multiple exports
- Tests package with changelog and documentation

## Configuration

The integration tests use these settings:

- **Server Port**: 19090 (to avoid conflicts)
- **Auth Token**: `integration-test-token`
- **Database**: SQLite file (`integration_test.db`)
- **Storage**: Local directory (`integration_test_storage/`)

## Cleanup

Test artifacts are automatically cleaned up, including:
- Built server binary (`repub-test`)
- Test database (`integration_test.db`)
- Storage directory (`integration_test_storage/`)

## Troubleshooting

### Dart Not Found
```bash
# Install Dart SDK
# macOS
brew install dart

# Linux
sudo apt update && sudo apt install dart

# Or download from dart.dev
```

### Server Won't Start
- Check if port 19090 is already in use
- Ensure `AUTH_TOKEN` is set in the test environment
- Check server logs for detailed error messages

### Publish Fails
- Verify the package structure in `tests/packages/`
- Ensure pubspec.yaml is valid
- Check that the server is accepting connections

## Adding New Test Packages

To add a new test package:

1. Create directory under `tests/packages/your_package/`
2. Add `pubspec.yaml`, `README.md`, `CHANGELOG.md`
3. Create `lib/` directory with your Dart code
4. Create `test/` directory with unit tests
5. Update `integration_test.go` to publish your package