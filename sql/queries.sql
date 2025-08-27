-- name: CreatePackage :one
INSERT INTO packages (name, private, description, homepage, repository, documentation, topics)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetPackage :one
SELECT * FROM packages WHERE name = $1;

-- name: ListPackages :many
SELECT * FROM packages 
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: CreatePackageVersion :one
INSERT INTO package_versions (
    package_id, version, description, pubspec_yaml, readme, changelog,
    archive_path, archive_sha256, uploader
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetPackageVersions :many
SELECT * FROM package_versions 
WHERE package_id = $1 
ORDER BY created_at DESC;

-- name: GetLatestPackageVersion :one
SELECT * FROM package_versions 
WHERE package_id = $1 AND retracted = false
ORDER BY created_at DESC 
LIMIT 1;

-- name: AddPackageUploader :exec
INSERT INTO package_uploaders (package_id, uploader)
VALUES ($1, $2)
ON CONFLICT (package_id, uploader) DO NOTHING;

-- name: GetPackageUploaders :many
SELECT uploader FROM package_uploaders WHERE package_id = $1;

-- name: IncrementDownloadCount :exec
UPDATE package_versions SET download_count = download_count + 1 
WHERE package_id = $1 AND version = $2;

-- name: IncrementPackageDownloadCount :exec
UPDATE packages SET download_count = download_count + 1 WHERE id = $1;

-- name: UpdatePackageMetadata :exec
UPDATE packages 
SET description = $2, homepage = $3, repository = $4, documentation = $5, topics = $6, updated_at = NOW()
WHERE id = $1;