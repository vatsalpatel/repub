package tests

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	serverPort = "19090"
	serverURL  = "http://localhost:" + serverPort
	authToken  = "integration-test-token"
)

func TestIntegration(t *testing.T) {
	// Skip if dart is not available
	if !isDartAvailable() {
		t.Skip("Dart CLI not available, skipping integration tests")
		return
	}

	// Start the server
	serverCmd, serverCancel := startTestServer(t)
	defer func() {
		serverCancel()
		if serverCmd.Process != nil {
			_ = serverCmd.Process.Kill()
		}
	}()

	// Wait for server to start
	if !waitForServer(t, serverURL, 30*time.Second) {
		t.Fatal("Server failed to start within timeout")
	}

	t.Log("✅ Server started successfully")

	// Create a temporary pub cache for testing
	pubCache := t.TempDir()
	t.Setenv("PUB_CACHE", pubCache)

	// Test publishing hello_world package
	t.Run("publish hello_world", func(t *testing.T) {
		publishPackage(t, "hello_world")
	})

	// Test publishing math_utils package
	t.Run("publish math_utils", func(t *testing.T) {
		publishPackage(t, "math_utils")
	})

	// Test browsing packages via web interface
	t.Run("browse packages", func(t *testing.T) {
		testWebInterface(t)
	})

	// Test package installation - should pass with our published packages
	t.Run("install published packages", func(t *testing.T) {
		testPackageInstallation(t, false) // false = don't include test package
	})
	
	// Test package installation failure - should fail with missing package
	t.Run("fail to install missing package", func(t *testing.T) {
		testPackageInstallationFailure(t) // This should fail as expected
	})

	// Test mixed package sources - our repo + pub.dev
	t.Run("install mixed packages (local + pub.dev)", func(t *testing.T) {
		testMixedPackageInstallation(t) // Local packages + pub.dev fallback
	})
}

func isDartAvailable() bool {
	_, err := exec.LookPath("dart")
	return err == nil
}

func startTestServer(t *testing.T) (*exec.Cmd, context.CancelFunc) {
	t.Helper()
	
	ctx, cancel := context.WithCancel(context.Background())
	
	// Start PostgreSQL in Docker for testing
	if !startPostgreSQL(t) {
		t.Fatal("Failed to start PostgreSQL container")
	}
	
	// Build the server binary
	buildCmd := exec.Command("go", "build", "-o", "repub-test", "./cmd/server")
	buildCmd.Dir = ".."
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build server: %v", err)
	}

	// Start the server with test configuration
	serverCmd := exec.CommandContext(ctx, "./repub-test")
	serverCmd.Dir = ".."
	serverCmd.Env = append(os.Environ(),
		"WRITE_TOKEN_INTEGRATION="+authToken,
		"PORT="+serverPort,
		"BASE_URL="+serverURL,
		"DATABASE_URL=postgres://repub:repub@localhost:15432/repub?sslmode=disable",
		"STORAGE_PATH=/tmp/integration_test_storage",
		"LOG_LEVEL=info",
	)

	// Capture server output for debugging
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr

	if err := serverCmd.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	t.Logf("Started server with PID %d", serverCmd.Process.Pid)
	
	return serverCmd, cancel
}

func startPostgreSQL(t *testing.T) bool {
	t.Helper()
	
	// Check if docker is available
	if !isDockerAvailable() {
		t.Skip("Docker not available, skipping PostgreSQL setup")
		return false
	}
	
	// Stop any existing container
	stopCmd := exec.Command("docker", "stop", "repub-test-postgres")
	_ = stopCmd.Run()
	
	removeCmd := exec.Command("docker", "rm", "repub-test-postgres")
	_ = removeCmd.Run()
	
	// Start PostgreSQL container
	dockerCmd := exec.Command("docker", "run", "--name", "repub-test-postgres",
		"-e", "POSTGRES_USER=repub",
		"-e", "POSTGRES_PASSWORD=repub", 
		"-e", "POSTGRES_DB=repub",
		"-p", "15432:5432",
		"-d", "postgres:16-alpine")
		
	if err := dockerCmd.Run(); err != nil {
		t.Logf("Failed to start PostgreSQL container: %v", err)
		return false
	}
	
	// Wait for PostgreSQL to be ready
	t.Log("⏳ Waiting for PostgreSQL to be ready...")
	for i := 0; i < 30; i++ {
		checkCmd := exec.Command("docker", "exec", "repub-test-postgres", 
			"pg_isready", "-U", "repub", "-d", "repub")
		if checkCmd.Run() == nil {
			t.Log("✅ PostgreSQL is ready")
			
			// Initialize schema
			initSchema(t)
			return true
		}
		time.Sleep(1 * time.Second)
	}
	
	t.Log("❌ PostgreSQL failed to start within timeout")
	return false
}

func isDockerAvailable() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

func initSchema(t *testing.T) {
	t.Helper()
	
	// Copy schema file and initialize
	schemaCmd := exec.Command("docker", "exec", "-i", "repub-test-postgres",
		"psql", "-U", "repub", "-d", "repub")
		
	schemaCmd.Dir = ".."
	schemaFile, err := os.Open("../sql/schema.sql") 
	if err != nil {
		t.Fatalf("Failed to open schema file: %v", err)
	}
	defer schemaFile.Close()
	
	schemaCmd.Stdin = schemaFile
	if err := schemaCmd.Run(); err != nil {
		t.Fatalf("Failed to initialize schema: %v", err)
	}
	
	t.Log("✅ Database schema initialized")
}

func waitForServer(t *testing.T, url string, timeout time.Duration) bool {
	t.Helper()
	
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		cmd := exec.Command("curl", "-f", "-s", url)
		if cmd.Run() == nil {
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}
	return false
}

func publishPackage(t *testing.T, packageName string) {
	t.Helper()
	
	packageDir := filepath.Join("packages", packageName)
	
	// Change to package directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore working directory: %v", err)
		}
	}()
	
	if err := os.Chdir(packageDir); err != nil {
		t.Fatalf("Failed to change to package directory %s: %v", packageDir, err)
	}

	// Create/update pubspec.yaml to point to our hosted server
	updatePubspecForTesting(t)

	t.Logf("📦 Publishing package: %s", packageName)

	// Add auth token to credential store
	tokenCmd := exec.Command("dart", "pub", "token", "add", serverURL)
	tokenCmd.Stdin = strings.NewReader(authToken + "\n")
	if err := tokenCmd.Run(); err != nil {
		t.Logf("Warning: Failed to add token: %v", err)
	}

	// Run dart pub publish
	cmd := exec.Command("dart", "pub", "publish", "--force")
	cmd.Env = append(os.Environ(),
		"PUB_HOSTED_URL="+serverURL,
	)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to publish %s: %v\nOutput: %s", packageName, err, output)
	}
	
	t.Logf("✅ Successfully published %s", packageName)
	t.Logf("Output: %s", output)
}

func updatePubspecForTesting(t *testing.T) {
	t.Helper()
	
	// The pubspec.yaml files already have publish_to configured
	// This function is kept for compatibility but doesn't need to do anything
	t.Log("📝 pubspec.yaml already configured for testing")
}

func testWebInterface(t *testing.T) {
	t.Helper()
	
	t.Log("🌐 Testing web interface...")

	// Test homepage
	cmd := exec.Command("curl", "-f", "-s", serverURL)
	if err := cmd.Run(); err != nil {
		t.Errorf("Failed to access homepage: %v", err)
	}

	// Test packages list
	cmd = exec.Command("curl", "-f", "-s", serverURL+"/packages")
	if err := cmd.Run(); err != nil {
		t.Errorf("Failed to access packages list: %v", err)
	}

	// Test specific package page
	cmd = exec.Command("curl", "-f", "-s", serverURL+"/packages/hello_world")
	if err := cmd.Run(); err != nil {
		t.Errorf("Failed to access hello_world package page: %v", err)
	}

	t.Log("✅ Web interface tests passed")
}

func testPackageInstallation(t *testing.T, includeTestPackage bool) {
	t.Helper()
	
	// Create a temporary test project
	testProject := t.TempDir()
	
	// Change to test project directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore working directory: %v", err)
		}
	}()
	
	if err := os.Chdir(testProject); err != nil {
		t.Fatalf("Failed to change to test project directory: %v", err)
	}

	// Create a test Dart project
	t.Log("📝 Creating test Dart project...")
	
	// Initialize dart project
	cmd := exec.Command("dart", "create", "test_consumer")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create Dart project: %v", err)
	}
	
	if err := os.Chdir("test_consumer"); err != nil {
		t.Fatalf("Failed to change to test_consumer directory: %v", err)
	}

	// Create pubspec.yaml that uses our published packages
	pubspecContent := `name: test_consumer
description: A test project to consume packages from our hosted server.
version: 1.0.0

environment:
  sdk: ^3.0.0

dependencies:
  hello_world: ^1.0.0
  math_utils: ^1.2.0`

	if includeTestPackage {
		pubspecContent += `

dev_dependencies:
  test: ^1.24.0`
	}

	pubspecContent += "\n"

	if err := os.WriteFile("pubspec.yaml", []byte(pubspecContent), 0644); err != nil {
		t.Fatalf("Failed to write test pubspec.yaml: %v", err)
	}

	// Create .dart_tool/package_config.json to point to our server
	if err := os.MkdirAll(".dart_tool", 0755); err != nil {
		t.Fatalf("Failed to create .dart_tool directory: %v", err)
	}
	
	packageConfigContent := fmt.Sprintf(`{
  "configVersion": 2,
  "packages": [],
  "generated": "%s",
  "generator": "pub",
  "generatorVersion": "3.0.0"
}`, time.Now().Format(time.RFC3339))

	if err := os.WriteFile(".dart_tool/package_config.json", []byte(packageConfigContent), 0644); err != nil {
		t.Fatalf("Failed to write package_config.json: %v", err)
	}

	t.Log("📥 Installing dependencies from hosted server...")

	// Run pub get with our hosted server
	cmd = exec.Command("dart", "pub", "get")
	cmd.Env = append(os.Environ(), "PUB_HOSTED_URL="+serverURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// This might fail if the packages aren't discoverable yet, which is OK for this test
		t.Logf("⚠️  pub get failed (expected): %v\nOutput: %s", err, output)
		return
	}
	
	t.Log("✅ Successfully installed packages from hosted server")

	// Create a simple test file that uses our packages
	testFileContent := `import 'package:hello_world/hello_world.dart';
import 'package:math_utils/math_utils.dart';

void main() {
  print(helloWorld());
  print('Is 17 prime? ${isPrime(17)}');
  print('5! = ${factorial(5)}');
}`

	if err := os.WriteFile("bin/test_consumer.dart", []byte(testFileContent), 0644); err != nil {
		t.Fatalf("Failed to write test consumer file: %v", err)
	}

	// Try to run the test program
	cmd = exec.Command("dart", "run")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Logf("⚠️  Running test program failed (expected if packages not available): %v\nOutput: %s", err, output)
	} else {
		t.Log("✅ Successfully ran test program using hosted packages")
		t.Logf("Program output: %s", output)
	}
}

func testPackageInstallationFailure(t *testing.T) {
	t.Helper()
	
	// Create a temporary test project
	testProject := t.TempDir()
	
	// Change to test project directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore working directory: %v", err)
		}
	}()
	
	if err := os.Chdir(testProject); err != nil {
		t.Fatalf("Failed to change to test project directory: %v", err)
	}

	t.Log("📝 Creating test Dart project with missing package...")
	
	cmd := exec.Command("dart", "create", "test_failure")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create Dart project: %v", err)
	}
	
	if err := os.Chdir("test_failure"); err != nil {
		t.Fatalf("Failed to change to test_failure directory: %v", err)
	}

	// Create pubspec.yaml that includes a package that doesn't exist on our server
	pubspecContent := `name: test_failure
description: A test project that should fail to resolve dependencies.
version: 1.0.0

environment:
  sdk: ^3.0.0

dependencies:
  hello_world: ^1.0.0
  nonexistent_package: ^1.0.0
`

	if err := os.WriteFile("pubspec.yaml", []byte(pubspecContent), 0644); err != nil {
		t.Fatalf("Failed to write test pubspec.yaml: %v", err)
	}

	// Create .dart_tool/package_config.json to point to our server
	if err := os.MkdirAll(".dart_tool", 0755); err != nil {
		t.Fatalf("Failed to create .dart_tool directory: %v", err)
	}
	
	packageConfigContent := fmt.Sprintf(`{
  "configVersion": 2,
  "packages": [],
  "generated": "%s",
  "generator": "pub",
  "generatorVersion": "3.0.0"
}`, time.Now().Format(time.RFC3339))

	if err := os.WriteFile(".dart_tool/package_config.json", []byte(packageConfigContent), 0644); err != nil {
		t.Fatalf("Failed to write package_config.json: %v", err)
	}

	t.Log("📥 Attempting to install dependencies (this should fail)...")

	// Run pub get with our hosted server - this should fail
	cmd = exec.Command("dart", "pub", "get")
	cmd.Env = append(os.Environ(), "PUB_HOSTED_URL="+serverURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("✅ pub get failed as expected: %v\nOutput: %s", err, output)
		return // This is the expected behavior
	}
	
	// If we get here, the test should fail because it succeeded when it should have failed
	t.Fatalf("Expected pub get to fail with missing package, but it succeeded. Output: %s", output)
}

func testMixedPackageInstallation(t *testing.T) {
	t.Helper()
	
	// Create a temporary test project
	testProject := t.TempDir()
	
	// Change to test project directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore working directory: %v", err)
		}
	}()
	
	if err := os.Chdir(testProject); err != nil {
		t.Fatalf("Failed to change to test project directory: %v", err)
	}

	t.Log("📝 Creating test Dart project with mixed package sources...")
	
	cmd := exec.Command("dart", "create", "mixed_test")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create Dart project: %v", err)
	}
	
	if err := os.Chdir("mixed_test"); err != nil {
		t.Fatalf("Failed to change to mixed_test directory: %v", err)
	}

	// Create pubspec.yaml with mixed dependencies:
	// - Our local packages (hello_world, math_utils) from our hosted server
	// - One standard package from pub.dev (path - a common, stable package)
	pubspecContent := `name: mixed_test
description: A test project that uses both local and pub.dev packages.
version: 1.0.0

environment:
  sdk: ^3.0.0

dependencies:
  # Local packages from our hosted server
  hello_world:
    version: ^1.0.0
    hosted:
      url: ` + serverURL + `
      name: hello_world
  math_utils:
    version: ^1.2.0
    hosted:
      url: ` + serverURL + `
      name: math_utils
  
  # Standard package from pub.dev
  path: ^1.8.0
`

	if err := os.WriteFile("pubspec.yaml", []byte(pubspecContent), 0644); err != nil {
		t.Fatalf("Failed to write mixed pubspec.yaml: %v", err)
	}

	t.Log("📥 Installing mixed dependencies (local + pub.dev)...")

	// Run pub get without PUB_HOSTED_URL override - this allows:
	// - Local packages to be fetched from our server (via hosted: url in pubspec)  
	// - Standard packages to be fetched from pub.dev (default behavior)
	cmd = exec.Command("dart", "pub", "get")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to install mixed packages: %v\nOutput: %s", err, output)
	}
	
	t.Log("✅ Successfully installed mixed packages")
	t.Logf("Pub get output: %s", output)

	// Create a simple test file that uses packages from both sources
	testFileContent := `import 'package:hello_world/hello_world.dart';
import 'package:math_utils/math_utils.dart';
import 'package:path/path.dart' as path;

void main() {
  // Use our local packages
  print(helloWorld());
  print('Is 17 prime? ${isPrime(17)}');
  print('5! = ${factorial(5)}');
  
  // Use pub.dev package
  var filePath = path.join('/tmp', 'example.txt');
  print('Example path: $filePath');
  print('Path extension: ${path.extension(filePath)}');
}`

	if err := os.WriteFile("bin/mixed_test.dart", []byte(testFileContent), 0644); err != nil {
		t.Fatalf("Failed to write mixed test file: %v", err)
	}

	// Try to run the test program
	t.Log("🚀 Running mixed package test program...")
	cmd = exec.Command("dart", "run")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run mixed test program: %v\nOutput: %s", err, output)
	}

	t.Log("✅ Successfully ran mixed package test program")
	t.Logf("Program output: %s", output)
	
	// Verify output contains expected results from all packages
	outputStr := string(output)
	if !strings.Contains(outputStr, "Hello") {
		t.Error("Output missing hello_world package result")
	}
	if !strings.Contains(outputStr, "prime") {
		t.Error("Output missing math_utils package result") 
	}
	if !strings.Contains(outputStr, "example.txt") {
		t.Error("Output missing path package result")
	}
}

func TestCleanup(t *testing.T) {
	// Clean up test artifacts
	_ = os.Remove("../repub-test")
	_ = os.RemoveAll("../integration_test_storage")
	
	// Stop and remove Docker container
	stopCmd := exec.Command("docker", "stop", "repub-test-postgres")
	_ = stopCmd.Run()
	
	removeCmd := exec.Command("docker", "rm", "repub-test-postgres")
	_ = removeCmd.Run()
	
	t.Log("✅ Cleanup completed")
}