package pubspec

import (
	"context"
	"fmt"
	"repub/internal/domain"
	"strings"

	"gopkg.in/yaml.v3"
)

type parserRepository struct{}

// NewParserRepository creates a new pubspec parser repository
func NewParserRepository() Repository {
	return &parserRepository{}
}

func (p *parserRepository) ParseYAML(ctx context.Context, yamlContent string) (*domain.Pubspec, error) {
	if strings.TrimSpace(yamlContent) == "" {
		return nil, fmt.Errorf("pubspec content is empty")
	}

	// Parse YAML to strongly typed struct
	var pubspec domain.Pubspec
	if err := yaml.Unmarshal([]byte(yamlContent), &pubspec); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate required fields
	if err := p.ValidatePubspec(ctx, &pubspec); err != nil {
		return nil, fmt.Errorf("pubspec validation failed: %w", err)
	}

	return &pubspec, nil
}

func (p *parserRepository) ValidatePubspec(ctx context.Context, pubspec *domain.Pubspec) error {
	if pubspec.Name == "" {
		return fmt.Errorf("package name is required")
	}

	if pubspec.Version == "" {
		return fmt.Errorf("package version is required")
	}

	// Validate package name format
	if !isValidPackageName(pubspec.Name) {
		return fmt.Errorf("invalid package name format: %s", pubspec.Name)
	}

	// Validate version format (basic semantic versioning)
	if !isValidVersion(pubspec.Version) {
		return fmt.Errorf("invalid version format: %s", pubspec.Version)
	}

	return nil
}


func (p *parserRepository) ExtractDependencies(ctx context.Context, pubspec *domain.Pubspec) (map[string]*domain.Dependency, error) {
	dependencies := make(map[string]*domain.Dependency)

	// Process regular dependencies
	for name, dep := range pubspec.Dependencies {
		parsed, err := p.parseDependency(name, dep)
		if err != nil {
			return nil, fmt.Errorf("failed to parse dependency %s: %w", name, err)
		}
		dependencies[name] = parsed
	}

	// Process dev dependencies
	for name, dep := range pubspec.DevDependencies {
		parsed, err := p.parseDependency(name, dep)
		if err != nil {
			return nil, fmt.Errorf("failed to parse dev dependency %s: %w", name, err)
		}
		dependencies[name] = parsed
	}

	return dependencies, nil
}

func (p *parserRepository) parseDependency(name string, dep interface{}) (*domain.Dependency, error) {
	switch v := dep.(type) {
	case string:
		// Simple version constraint
		return &domain.Dependency{Version: v}, nil
	case map[string]interface{}:
		// Complex dependency specification
		dependency := &domain.Dependency{
			Extra: make(map[string]interface{}),
		}

		for key, value := range v {
			switch key {
			case "version":
				if str, ok := value.(string); ok {
					dependency.Version = str
				}
			case "hosted":
				if str, ok := value.(string); ok {
					dependency.Hosted = str
				}
			case "git":
				if gitMap, ok := value.(map[string]interface{}); ok {
					git := &domain.GitDependency{}
					if url, ok := gitMap["url"].(string); ok {
						git.URL = url
					}
					if ref, ok := gitMap["ref"].(string); ok {
						git.Ref = ref
					}
					if path, ok := gitMap["path"].(string); ok {
						git.Path = path
					}
					dependency.Git = git
				}
			case "path":
				if str, ok := value.(string); ok {
					dependency.Path = str
				}
			case "sdk":
				if str, ok := value.(string); ok {
					dependency.SDK = str
				}
			default:
				dependency.Extra[key] = value
			}
		}

		return dependency, nil
	default:
		return nil, fmt.Errorf("unsupported dependency format for %s", name)
	}
}

// isValidPackageName checks if package name follows pub.dev conventions
func isValidPackageName(name string) bool {
	if len(name) == 0 || len(name) > 64 {
		return false
	}

	// Must start with letter or underscore
	first := name[0]
	if (first < 'a' || first > 'z') && (first < 'A' || first > 'Z') && first != '_' {
		return false
	}

	// Can contain letters, numbers, underscores
	for _, char := range name {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && (char < '0' || char > '9') && char != '_' {
			return false
		}
	}

	return true
}

// isValidVersion checks basic semantic versioning format
func isValidVersion(version string) bool {
	if len(version) == 0 {
		return false
	}

	// Basic check for semantic versioning pattern - needs at least 3 parts (major.minor.patch)
	parts := strings.Split(version, ".")
	if len(parts) < 3 {
		return false
	}

	// Check each part is numeric (simplified check)
	for _, part := range parts {
		if len(part) == 0 {
			return false
		}
		// Allow pre-release and build metadata
		if strings.Contains(part, "-") || strings.Contains(part, "+") {
			continue
		}
		// Check if part starts with digit
		if part[0] < '0' || part[0] > '9' {
			return false
		}
	}

	return true
}