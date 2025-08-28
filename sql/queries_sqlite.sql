-- name: CreatePackage :one
INSERT INTO packages (name, private, description, homepage, repository, documentation)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING id, name, private, description, homepage, repository, documentation, created_at, updated_at;

-- name: GetPackage :one
SELECT id, name, private, description, homepage, repository, documentation, created_at, updated_at FROM packages WHERE name = ?;

-- name: ListPackages :many
SELECT id, name, private, description, homepage, repository, documentation, created_at, updated_at FROM packages 
ORDER BY name
LIMIT ? OFFSET ?;

-- name: UpdatePackageMetadata :exec
UPDATE packages 
SET description = ?, homepage = ?, repository = ?, documentation = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: CreatePackageVersion :one
INSERT INTO package_versions (
    package_id, version, description, pubspec_yaml, readme, changelog,
    archive_path, archive_sha256, uploader
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id, package_id, version, description, pubspec_yaml, readme, changelog, archive_path, archive_sha256, uploader, retracted, created_at;

-- name: GetPackageVersions :many
SELECT id, package_id, version, description, pubspec_yaml, readme, changelog, archive_path, archive_sha256, uploader, retracted, created_at FROM package_versions 
WHERE package_id = ? 
ORDER BY created_at DESC;

-- name: GetLatestPackageVersion :one
SELECT id, package_id, version, description, pubspec_yaml, readme, changelog, archive_path, archive_sha256, uploader, retracted, created_at FROM package_versions 
WHERE package_id = ? AND retracted = false
ORDER BY created_at DESC 
LIMIT 1;

-- name: AddPackageUploader :exec
INSERT INTO package_uploaders (package_id, uploader)
VALUES (?, ?)
ON CONFLICT (package_id, uploader) DO NOTHING;

-- name: GetPackageUploaders :many
SELECT uploader FROM package_uploaders WHERE package_id = ?;