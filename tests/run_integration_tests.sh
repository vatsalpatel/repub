#!/bin/bash

set -e

echo "🚀 Starting Repub Integration Tests"
echo "=================================="

# Check if dart is available
if ! command -v dart &> /dev/null; then
    echo "❌ Dart CLI is not available. Please install Dart SDK."
    echo "   Visit: https://dart.dev/get-dart"
    exit 1
fi

echo "✅ Dart CLI found: $(dart --version)"

# Check if curl is available
if ! command -v curl &> /dev/null; then
    echo "❌ curl is not available. Please install curl."
    exit 1
fi

echo "✅ curl found"

# Clean up any previous test artifacts
echo "🧹 Cleaning up previous test artifacts..."
rm -f ../repub-test
rm -f ../integration_test.db
rm -rf ../integration_test_storage

# Make sure we're in the tests directory
cd "$(dirname "$0")"

# Run the integration tests
echo "🧪 Running integration tests..."
go test -v -timeout 5m ./...

echo ""
echo "✅ Integration tests completed successfully!"
echo "🎉 Your pub server is working correctly with the Dart CLI!"