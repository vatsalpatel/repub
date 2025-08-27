package pubspec

import (
	"context"
	"repub/internal/domain"
)

// Repository defines the interface for pubspec operations
type Repository interface {
	// ParseYAML parses a pubspec.yaml string and returns a typed Pubspec
	ParseYAML(ctx context.Context, yamlContent string) (*domain.Pubspec, error)
	
	// ValidatePubspec validates a pubspec for required fields and constraints
	ValidatePubspec(ctx context.Context, pubspec *domain.Pubspec) error
	
	// ExtractDependencies extracts and normalizes dependencies from pubspec
	ExtractDependencies(ctx context.Context, pubspec *domain.Pubspec) (map[string]*domain.Dependency, error)
}