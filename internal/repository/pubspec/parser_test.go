package pubspec

import (
	"context"
	"testing"
)

func TestParserRepository_ParseYAML(t *testing.T) {
	repo := NewParserRepository()

	tests := []struct {
		name      string
		yaml      string
		wantError bool
		wantName  string
		wantVer   string
	}{
		{
			name: "basic pubspec",
			yaml: `name: test_package
version: 1.0.0
description: A test package
homepage: https://example.com

dependencies:
  flutter:
    sdk: flutter
  http: ^0.13.0

dev_dependencies:
  flutter_test:
    sdk: flutter`,
			wantError: false,
			wantName:  "test_package",
			wantVer:   "1.0.0",
		},
		{
			name: "minimal pubspec",
			yaml: `name: minimal
version: 0.1.0`,
			wantError: false,
			wantName:  "minimal",
			wantVer:   "0.1.0",
		},
		{
			name: "missing name",
			yaml: `version: 1.0.0
description: Missing name`,
			wantError: true,
		},
		{
			name: "missing version",
			yaml: `name: missing_version
description: Missing version`,
			wantError: true,
		},
		{
			name:      "empty yaml",
			yaml:      "",
			wantError: true,
		},
		{
			name: "invalid yaml",
			yaml: `name: test
version: 1.0.0
invalid: yaml: content: [}`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.ParseYAML(context.Background(), tt.yaml)

			if tt.wantError && err == nil {
				t.Error("Expected error, got nil")
			} else if !tt.wantError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if !tt.wantError && result != nil {
				if result.Name != tt.wantName {
					t.Errorf("Expected name %s, got %s", tt.wantName, result.Name)
				}
				if result.Version != tt.wantVer {
					t.Errorf("Expected version %s, got %s", tt.wantVer, result.Version)
				}
			}
		})
	}
}


func TestParserRepository_ExtraFields(t *testing.T) {
	repo := NewParserRepository()

	yaml := `name: test_package
version: 1.0.0
description: A test package
custom_field: custom_value
nested_custom:
  field1: value1
  field2: value2

dependencies:
  flutter:
    sdk: flutter
  http: 
    version: ^0.13.0
    custom_dep_field: custom_dep_value`

	parsed, err := repo.ParseYAML(context.Background(), yaml)
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	// Test that Extra fields are captured in the pubspec
	if parsed.Extra == nil {
		t.Fatal("Extra should not be nil")
	}

	if parsed.Extra["custom_field"] != "custom_value" {
		t.Errorf("Expected custom_field 'custom_value', got %v", parsed.Extra["custom_field"])
	}

	nested, ok := parsed.Extra["nested_custom"].(map[string]interface{})
	if !ok {
		t.Error("nested_custom should be a map")
	} else {
		if nested["field1"] != "value1" {
			t.Errorf("Expected field1 'value1', got %v", nested["field1"])
		}
	}

	// Test dependency extra fields
	deps, err := repo.ExtractDependencies(context.Background(), parsed)
	if err != nil {
		t.Fatalf("ExtractDependencies failed: %v", err)
	}

	httpDep := deps["http"]
	if httpDep == nil {
		t.Fatal("http dependency should exist")
	}

	if httpDep.Version != "^0.13.0" {
		t.Errorf("Expected version '^0.13.0', got %s", httpDep.Version)
	}

	if httpDep.Extra == nil {
		t.Fatal("Dependency Extra should not be nil")
	}

	if httpDep.Extra["custom_dep_field"] != "custom_dep_value" {
		t.Errorf("Expected custom_dep_field 'custom_dep_value', got %v", httpDep.Extra["custom_dep_field"])
	}
}

func TestParserRepository_ExtractDependencies(t *testing.T) {
	repo := NewParserRepository()

	yaml := `name: test_package
version: 1.0.0

dependencies:
  flutter:
    sdk: flutter
  http: ^0.13.0
  local_package:
    path: ../local
  git_package:
    git:
      url: https://github.com/example/package.git
      ref: main
      path: subdir
  hosted_package:
    hosted: custom-host.com
  complex_git:
    git:
      url: https://github.com/example/complex.git

dev_dependencies:
  flutter_test:
    sdk: flutter`

	parsed, err := repo.ParseYAML(context.Background(), yaml)
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	deps, err := repo.ExtractDependencies(context.Background(), parsed)
	if err != nil {
		t.Fatalf("ExtractDependencies failed: %v", err)
	}

	// Check simple version dependency
	if deps["http"] == nil {
		t.Error("http dependency should exist")
	} else if deps["http"].Version != "^0.13.0" {
		t.Errorf("Expected http version '^0.13.0', got %s", deps["http"].Version)
	}

	// Check SDK dependency
	if deps["flutter"] == nil {
		t.Error("flutter dependency should exist")
	} else if deps["flutter"].SDK != "flutter" {
		t.Errorf("Expected flutter SDK 'flutter', got %s", deps["flutter"].SDK)
	}

	// Check path dependency
	if deps["local_package"] == nil {
		t.Error("local_package dependency should exist")
	} else if deps["local_package"].Path != "../local" {
		t.Errorf("Expected path '../local', got %s", deps["local_package"].Path)
	}

	// Check hosted dependency
	if deps["hosted_package"] == nil {
		t.Error("hosted_package dependency should exist")
	} else if deps["hosted_package"].Hosted != "custom-host.com" {
		t.Errorf("Expected hosted 'custom-host.com', got %s", deps["hosted_package"].Hosted)
	}

	// Check git dependency with all fields
	if deps["git_package"] == nil {
		t.Error("git_package dependency should exist")
	} else {
		git := deps["git_package"].Git
		if git == nil {
			t.Error("git dependency should have Git field")
		} else {
			if git.URL != "https://github.com/example/package.git" {
				t.Errorf("Expected git URL 'https://github.com/example/package.git', got %s", git.URL)
			}
			if git.Ref != "main" {
				t.Errorf("Expected git ref 'main', got %s", git.Ref)
			}
			if git.Path != "subdir" {
				t.Errorf("Expected git path 'subdir', got %s", git.Path)
			}
		}
	}

	// Check git dependency without optional fields
	if deps["complex_git"] == nil {
		t.Error("complex_git dependency should exist")
	} else {
		git := deps["complex_git"].Git
		if git == nil {
			t.Error("complex_git dependency should have Git field")
		} else {
			if git.URL != "https://github.com/example/complex.git" {
				t.Errorf("Expected git URL 'https://github.com/example/complex.git', got %s", git.URL)
			}
		}
	}

	// Check dev dependency
	if deps["flutter_test"] == nil {
		t.Error("flutter_test dev dependency should exist")
	}
}

func TestParserRepository_ExtractDependencies_ErrorCases(t *testing.T) {
	repo := NewParserRepository()

	tests := []struct {
		name      string
		yaml      string
		wantError bool
		errorMsg  string
	}{
		{
			name: "invalid dependency format",
			yaml: `name: test_package
version: 1.0.0

dependencies:
  invalid_dep: 123`,
			wantError: true,
			errorMsg:  "unsupported dependency format",
		},
		{
			name: "invalid dev dependency format",
			yaml: `name: test_package
version: 1.0.0

dev_dependencies:
  invalid_dep: []`,
			wantError: true,
			errorMsg:  "unsupported dependency format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := repo.ParseYAML(context.Background(), tt.yaml)
			if err != nil {
				t.Fatalf("ParseYAML failed: %v", err)
			}

			_, err = repo.ExtractDependencies(context.Background(), parsed)
			
			if tt.wantError && err == nil {
				t.Error("Expected error, got nil")
			} else if !tt.wantError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if tt.wantError && err != nil && !contains(err.Error(), tt.errorMsg) {
				t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
			}
		})
	}
}

func TestParserRepository_ParseDependency_EdgeCases(t *testing.T) {
	repo := NewParserRepository()

	yaml := `name: test_package
version: 1.0.0

dependencies:
  version_in_map:
    version: "^1.0.0"
  git_with_non_string_values:
    git:
      url: https://github.com/example/package.git
      ref: 123
      path: 456
  non_string_hosted:
    hosted: 123
  non_string_path:
    path: 123
  non_string_sdk:
    sdk: 123
  non_string_version:
    version: 123`

	parsed, err := repo.ParseYAML(context.Background(), yaml)
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	deps, err := repo.ExtractDependencies(context.Background(), parsed)
	if err != nil {
		t.Fatalf("ExtractDependencies failed: %v", err)
	}

	// Test version in map format
	if deps["version_in_map"] == nil {
		t.Error("version_in_map dependency should exist")
	} else if deps["version_in_map"].Version != "^1.0.0" {
		t.Errorf("Expected version '^1.0.0', got %s", deps["version_in_map"].Version)
	}

	// Test that non-string values are ignored gracefully
	if deps["non_string_hosted"] != nil && deps["non_string_hosted"].Hosted != "" {
		t.Error("non-string hosted should be ignored")
	}
}

func TestParserRepository_ValidatePubspec(t *testing.T) {
	repo := NewParserRepository()

	tests := []struct {
		name      string
		yaml      string
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid pubspec",
			yaml: `name: valid_package
version: 1.0.0`,
			wantError: false,
		},
		{
			name: "invalid package name with spaces",
			yaml: `name: "invalid package"
version: 1.0.0`,
			wantError: true,
			errorMsg:  "invalid package name format",
		},
		{
			name: "invalid package name starting with number",
			yaml: `name: 1invalid
version: 1.0.0`,
			wantError: true,
			errorMsg:  "invalid package name format",
		},
		{
			name: "invalid version format",
			yaml: `name: valid_package
version: invalid_version`,
			wantError: true,
			errorMsg:  "invalid version format",
		},
		{
			name: "valid version with pre-release",
			yaml: `name: valid_package
version: 1.0.0-beta.1`,
			wantError: false,
		},
		{
			name: "valid version with build metadata",
			yaml: `name: valid_package
version: 1.0.0+build.1`,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := repo.ParseYAML(context.Background(), tt.yaml)

			if tt.wantError && err == nil {
				t.Error("Expected error, got nil")
			} else if !tt.wantError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if tt.wantError && err != nil && tt.errorMsg != "" {
				if !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
			}
		})
	}
}

func TestIsValidPackageName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid_lowercase", "valid_package", true},
		{"valid_with_numbers", "package123", true},
		{"valid_underscore_start", "_package", true},
		{"valid_camelCase", "packageName", true},
		{"invalid_start_with_number", "1package", false},
		{"invalid_with_dash", "package-name", false},
		{"invalid_with_space", "package name", false},
		{"invalid_empty", "", false},
		{"invalid_too_long", "this_is_a_very_long_package_name_that_exceeds_the_maximum_length_allowed_by_pub", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidPackageName(tt.input)
			if result != tt.expected {
				t.Errorf("isValidPackageName(%s) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsValidVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid_semver", "1.0.0", true},
		{"valid_with_patch", "1.2.3", true},
		{"valid_with_prerelease", "1.0.0-beta.1", true},
		{"valid_with_build", "1.0.0+build.1", true},
		{"valid_complex", "1.0.0-alpha.1+build.123", true},
		{"invalid_empty", "", false},
		{"invalid_single_number", "1", false},
		{"invalid_no_patch", "1.0", false},
		{"invalid_letters", "abc", false},
		{"invalid_start_with_letter", "v1.0.0", false},
		{"invalid_empty_part", "1..0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidVersion(tt.input)
			if result != tt.expected {
				t.Errorf("isValidVersion(%s) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(substr) == 0 || len(s) >= len(substr) &&
		(s == substr || s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(s) > len(substr) && findInString(s, substr))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

