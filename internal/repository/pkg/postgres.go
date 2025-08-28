package pkg

import (
	"context"
	"database/sql"
	"repub/internal/domain"
	"repub/internal/repository/pkg/postgres"
)

type postgresPackageRepository struct {
	queries Queries
}

func NewPostgresPackageRepository(queries Queries) Repository {
	return &postgresPackageRepository{queries: queries}
}

func (r *postgresPackageRepository) GetPackage(ctx context.Context, name string) (*domain.Package, error) {
	pkg, err := r.queries.GetPackage(ctx, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &domain.Package{
		ID:            pkg.ID,
		Name:          pkg.Name,
		Private:       pkg.Private,
		Description:   nullStringToPtr(pkg.Description),
		Homepage:      nullStringToPtr(pkg.Homepage),
		Repository:    nullStringToPtr(pkg.Repository),
		Documentation: nullStringToPtr(pkg.Documentation),
		CreatedAt:     pkg.CreatedAt,
		UpdatedAt:     pkg.UpdatedAt,
	}, nil
}

func (r *postgresPackageRepository) CreatePackage(ctx context.Context, name string, private bool) (*domain.Package, error) {
	pkg, err := r.queries.CreatePackage(ctx, postgres.CreatePackageParams{
		Name:    name,
		Private: private,
	})
	if err != nil {
		return nil, err
	}

	return &domain.Package{
		ID:            pkg.ID,
		Name:          pkg.Name,
		Private:       pkg.Private,
		Description:   nullStringToPtr(pkg.Description),
		Homepage:      nullStringToPtr(pkg.Homepage),
		Repository:    nullStringToPtr(pkg.Repository),
		Documentation: nullStringToPtr(pkg.Documentation),
		CreatedAt:     pkg.CreatedAt,
		UpdatedAt:     pkg.UpdatedAt,
	}, nil
}

func (r *postgresPackageRepository) ListPackages(ctx context.Context, limit, offset int32) ([]*domain.Package, error) {
	packages, err := r.queries.ListPackages(ctx, postgres.ListPackagesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Package, len(packages))
	for i, pkg := range packages {
		result[i] = &domain.Package{
			ID:            pkg.ID,
			Name:          pkg.Name,
			Private:       pkg.Private,
			Description:   nullStringToPtr(pkg.Description),
			Homepage:      nullStringToPtr(pkg.Homepage),
			Repository:    nullStringToPtr(pkg.Repository),
			Documentation: nullStringToPtr(pkg.Documentation),
			CreatedAt:     pkg.CreatedAt,
			UpdatedAt:     pkg.UpdatedAt,
		}
	}

	return result, nil
}

func (r *postgresPackageRepository) GetPackageVersions(ctx context.Context, packageID int32) ([]*domain.PackageVersion, error) {
	versions, err := r.queries.GetPackageVersions(ctx, packageID)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.PackageVersion, len(versions))
	for i, v := range versions {
		result[i] = &domain.PackageVersion{
			ID:            v.ID,
			PackageID:     v.PackageID,
			Version:       v.Version,
			Description:   nullStringToPtr(v.Description),
			PubspecYaml:   v.PubspecYaml,
			Readme:        nullStringToPtr(v.Readme),
			Changelog:     nullStringToPtr(v.Changelog),
			ArchivePath:   v.ArchivePath,
			ArchiveSha256: nullStringToPtr(v.ArchiveSha256),
			Uploader:      nullStringToPtr(v.Uploader),
			Retracted:     v.Retracted,
			CreatedAt:     v.CreatedAt,
		}
	}

	return result, nil
}

func (r *postgresPackageRepository) GetLatestVersion(ctx context.Context, packageID int32) (*domain.PackageVersion, error) {
	version, err := r.queries.GetLatestPackageVersion(ctx, packageID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &domain.PackageVersion{
		ID:            version.ID,
		PackageID:     version.PackageID,
		Version:       version.Version,
		Description:   nullStringToPtr(version.Description),
		PubspecYaml:   version.PubspecYaml,
		Readme:        nullStringToPtr(version.Readme),
		Changelog:     nullStringToPtr(version.Changelog),
		ArchivePath:   version.ArchivePath,
		ArchiveSha256: nullStringToPtr(version.ArchiveSha256),
		Uploader:      nullStringToPtr(version.Uploader),
		Retracted:     version.Retracted,
		CreatedAt:     version.CreatedAt,
	}, nil
}

func (r *postgresPackageRepository) CreateVersion(ctx context.Context, version *domain.PackageVersion) (*domain.PackageVersion, error) {
	var description sql.NullString
	if version.Description != nil {
		description = sql.NullString{String: *version.Description, Valid: true}
	}

	var archiveSha256 sql.NullString
	if version.ArchiveSha256 != nil {
		archiveSha256 = sql.NullString{String: *version.ArchiveSha256, Valid: true}
	}

	var uploader sql.NullString
	if version.Uploader != nil {
		uploader = sql.NullString{String: *version.Uploader, Valid: true}
	}

	var readme sql.NullString
	if version.Readme != nil {
		readme = sql.NullString{String: *version.Readme, Valid: true}
	}

	var changelog sql.NullString
	if version.Changelog != nil {
		changelog = sql.NullString{String: *version.Changelog, Valid: true}
	}

	created, err := r.queries.CreatePackageVersion(ctx, postgres.CreatePackageVersionParams{
		PackageID:     version.PackageID,
		Version:       version.Version,
		Description:   description,
		PubspecYaml:   version.PubspecYaml,
		Readme:        readme,
		Changelog:     changelog,
		ArchivePath:   version.ArchivePath,
		ArchiveSha256: archiveSha256,
		Uploader:      uploader,
	})
	if err != nil {
		return nil, err
	}

	return &domain.PackageVersion{
		ID:            created.ID,
		PackageID:     created.PackageID,
		Version:       created.Version,
		Description:   nullStringToPtr(created.Description),
		PubspecYaml:   created.PubspecYaml,
		Readme:        nullStringToPtr(created.Readme),
		Changelog:     nullStringToPtr(created.Changelog),
		ArchivePath:   created.ArchivePath,
		ArchiveSha256: nullStringToPtr(created.ArchiveSha256),
		Uploader:      nullStringToPtr(created.Uploader),
		Retracted:     created.Retracted,
		CreatedAt:     created.CreatedAt,
	}, nil
}

func (r *postgresPackageRepository) GetUploaders(ctx context.Context, packageID int32) ([]string, error) {
	uploaders, err := r.queries.GetPackageUploaders(ctx, packageID)
	return uploaders, err
}

func (r *postgresPackageRepository) AddUploader(ctx context.Context, packageID int32, uploader string) error {
	return r.queries.AddPackageUploader(ctx, postgres.AddPackageUploaderParams{
		PackageID: packageID,
		Uploader:  uploader,
	})
}

func nullStringToPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}
