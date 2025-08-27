CREATE TABLE packages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    private BOOLEAN NOT NULL DEFAULT FALSE,
    description TEXT,
    homepage TEXT,
    repository TEXT,
    documentation TEXT,
    topics TEXT, -- JSON string of topics like ["state-management", "bloc"]
    download_count INTEGER NOT NULL DEFAULT 0,
    like_count INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE package_versions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    package_id INTEGER NOT NULL REFERENCES packages(id) ON DELETE CASCADE,
    version TEXT NOT NULL,
    description TEXT,
    pubspec_yaml TEXT NOT NULL,
    readme TEXT, -- Store README content
    changelog TEXT, -- Store CHANGELOG content
    archive_path TEXT NOT NULL,
    archive_sha256 TEXT,
    uploader TEXT,
    retracted BOOLEAN NOT NULL DEFAULT FALSE,
    download_count INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(package_id, version)
);

CREATE TABLE package_uploaders (
    package_id INTEGER REFERENCES packages(id) ON DELETE CASCADE,
    uploader TEXT NOT NULL,
    PRIMARY KEY (package_id, uploader)
);

CREATE INDEX idx_packages_name ON packages(name);
CREATE INDEX idx_package_versions_package_id ON package_versions(package_id);