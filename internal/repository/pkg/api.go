package pkg

import (
	"context"
	"repub/internal/db/postgres"
	"repub/internal/domain"
)

type Queries interface {
	GetPackage(ctx context.Context, name string) (postgres.Package, error)
	CreatePackage(ctx context.Context, params postgres.CreatePackageParams) (postgres.Package, error)
	ListPackages(ctx context.Context, params postgres.ListPackagesParams) ([]postgres.Package, error)
	GetPackageVersions(ctx context.Context, packageID int32) ([]postgres.PackageVersion, error)
	GetLatestPackageVersion(ctx context.Context, packageID int32) (postgres.PackageVersion, error)
	CreatePackageVersion(ctx context.Context, params postgres.CreatePackageVersionParams) (postgres.PackageVersion, error)
	GetPackageUploaders(ctx context.Context, packageID int32) ([]string, error)
	AddPackageUploader(ctx context.Context, params postgres.AddPackageUploaderParams) error
}

type Repository interface {
	GetPackage(ctx context.Context, name string) (*domain.Package, error)
	CreatePackage(ctx context.Context, name string, private bool) (*domain.Package, error)
	ListPackages(ctx context.Context, limit, offset int32) ([]*domain.Package, error)

	GetPackageVersions(ctx context.Context, packageID int32) ([]*domain.PackageVersion, error)
	GetLatestVersion(ctx context.Context, packageID int32) (*domain.PackageVersion, error)
	CreateVersion(ctx context.Context, version *domain.PackageVersion) (*domain.PackageVersion, error)

	GetUploaders(ctx context.Context, packageID int32) ([]string, error)
	AddUploader(ctx context.Context, packageID int32, uploader string) error
}
