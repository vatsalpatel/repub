package testutil

import (
	"context"
	"database/sql"

	"repub/internal/domain"
	"repub/internal/repository/pkg/sqlite"
)

// sqlitePackageRepository implements pkg.Repository using SQLite
type sqlitePackageRepository struct {
	queries *sqlite.Queries
}

func newSQLitePackageRepository(queries *sqlite.Queries) *sqlitePackageRepository {
	return &sqlitePackageRepository{queries: queries}
}

func (r *sqlitePackageRepository) GetPackage(ctx context.Context, name string) (*domain.Package, error) {
	pkg, err := r.queries.GetPackage(ctx, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &domain.Package{
		ID:            int32(pkg.ID),
		Name:          pkg.Name,
		Private:       pkg.Private,
		Description:   sqliteNullStringToPtr(pkg.Description),
		Homepage:      sqliteNullStringToPtr(pkg.Homepage),
		Repository:    sqliteNullStringToPtr(pkg.Repository),
		Documentation: sqliteNullStringToPtr(pkg.Documentation),
		CreatedAt:     pkg.CreatedAt,
		UpdatedAt:     pkg.UpdatedAt,
	}, nil
}

func (r *sqlitePackageRepository) CreatePackage(ctx context.Context, name string, private bool) (*domain.Package, error) {
	pkg, err := r.queries.CreatePackage(ctx, sqlite.CreatePackageParams{
		Name:    name,
		Private: private,
	})
	if err != nil {
		return nil, err
	}

	return &domain.Package{
		ID:            int32(pkg.ID),
		Name:          pkg.Name,
		Private:       pkg.Private,
		Description:   sqliteNullStringToPtr(pkg.Description),
		Homepage:      sqliteNullStringToPtr(pkg.Homepage),
		Repository:    sqliteNullStringToPtr(pkg.Repository),
		Documentation: sqliteNullStringToPtr(pkg.Documentation),
		CreatedAt:     pkg.CreatedAt,
		UpdatedAt:     pkg.UpdatedAt,
	}, nil
}

func (r *sqlitePackageRepository) ListPackages(ctx context.Context, limit, offset int32) ([]*domain.Package, error) {
	packages, err := r.queries.ListPackages(ctx, sqlite.ListPackagesParams{
		Limit:  int64(limit),
		Offset: int64(offset),
	})
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Package, len(packages))
	for i, pkg := range packages {
		result[i] = &domain.Package{
			ID:            int32(pkg.ID),
			Name:          pkg.Name,
			Private:       pkg.Private,
			Description:   sqliteNullStringToPtr(pkg.Description),
			Homepage:      sqliteNullStringToPtr(pkg.Homepage),
			Repository:    sqliteNullStringToPtr(pkg.Repository),
			Documentation: sqliteNullStringToPtr(pkg.Documentation),
			CreatedAt:     pkg.CreatedAt,
			UpdatedAt:     pkg.UpdatedAt,
		}
	}

	return result, nil
}

func (r *sqlitePackageRepository) GetPackageVersions(ctx context.Context, packageID int32) ([]*domain.PackageVersion, error) {
	versions, err := r.queries.GetPackageVersions(ctx, int64(packageID))
	if err != nil {
		return nil, err
	}

	result := make([]*domain.PackageVersion, len(versions))
	for i, v := range versions {
		result[i] = &domain.PackageVersion{
			ID:            int32(v.ID),
			PackageID:     int32(v.PackageID),
			Version:       v.Version,
			Description:   sqliteNullStringToPtr(v.Description),
			PubspecYaml:   v.PubspecYaml,
			Readme:        sqliteNullStringToPtr(v.Readme),
			Changelog:     sqliteNullStringToPtr(v.Changelog),
			ArchivePath:   v.ArchivePath,
			ArchiveSha256: sqliteNullStringToPtr(v.ArchiveSha256),
			Uploader:      sqliteNullStringToPtr(v.Uploader),
			Retracted:     v.Retracted,
			CreatedAt:     v.CreatedAt,
		}
	}

	return result, nil
}

func (r *sqlitePackageRepository) GetLatestVersion(ctx context.Context, packageID int32) (*domain.PackageVersion, error) {
	version, err := r.queries.GetLatestPackageVersion(ctx, int64(packageID))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &domain.PackageVersion{
		ID:            int32(version.ID),
		PackageID:     int32(version.PackageID),
		Version:       version.Version,
		Description:   sqliteNullStringToPtr(version.Description),
		PubspecYaml:   version.PubspecYaml,
		Readme:        sqliteNullStringToPtr(version.Readme),
		Changelog:     sqliteNullStringToPtr(version.Changelog),
		ArchivePath:   version.ArchivePath,
		ArchiveSha256: sqliteNullStringToPtr(version.ArchiveSha256),
		Uploader:      sqliteNullStringToPtr(version.Uploader),
		Retracted:     version.Retracted,
		CreatedAt:     version.CreatedAt,
	}, nil
}

func (r *sqlitePackageRepository) CreateVersion(ctx context.Context, version *domain.PackageVersion) (*domain.PackageVersion, error) {
	var description sql.NullString
	if version.Description != nil {
		description = sql.NullString{String: *version.Description, Valid: true}
	}

	var readme sql.NullString
	if version.Readme != nil {
		readme = sql.NullString{String: *version.Readme, Valid: true}
	}

	var changelog sql.NullString
	if version.Changelog != nil {
		changelog = sql.NullString{String: *version.Changelog, Valid: true}
	}

	var archiveSha256 sql.NullString
	if version.ArchiveSha256 != nil {
		archiveSha256 = sql.NullString{String: *version.ArchiveSha256, Valid: true}
	}

	var uploader sql.NullString
	if version.Uploader != nil {
		uploader = sql.NullString{String: *version.Uploader, Valid: true}
	}

	created, err := r.queries.CreatePackageVersion(ctx, sqlite.CreatePackageVersionParams{
		PackageID:     int64(version.PackageID),
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
		ID:            int32(created.ID),
		PackageID:     int32(created.PackageID),
		Version:       created.Version,
		Description:   sqliteNullStringToPtr(created.Description),
		PubspecYaml:   created.PubspecYaml,
		Readme:        sqliteNullStringToPtr(created.Readme),
		Changelog:     sqliteNullStringToPtr(created.Changelog),
		ArchivePath:   created.ArchivePath,
		ArchiveSha256: sqliteNullStringToPtr(created.ArchiveSha256),
		Uploader:      sqliteNullStringToPtr(created.Uploader),
		Retracted:     created.Retracted,
		CreatedAt:     created.CreatedAt,
	}, nil
}

func (r *sqlitePackageRepository) GetUploaders(ctx context.Context, packageID int32) ([]string, error) {
	return r.queries.GetPackageUploaders(ctx, sql.NullInt64{Int64: int64(packageID), Valid: true})
}

func (r *sqlitePackageRepository) AddUploader(ctx context.Context, packageID int32, uploader string) error {
	return r.queries.AddPackageUploader(ctx, sqlite.AddPackageUploaderParams{
		PackageID: sql.NullInt64{Int64: int64(packageID), Valid: true},
		Uploader:  uploader,
	})
}

func sqliteNullStringToPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}
