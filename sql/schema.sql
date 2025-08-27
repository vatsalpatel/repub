CREATE TABLE packages (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    private BOOLEAN NOT NULL DEFAULT FALSE,
    description TEXT,
    homepage TEXT,
    repository TEXT,
    documentation TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE package_versions (
    id SERIAL PRIMARY KEY,
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
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(package_id, version)
);

CREATE TABLE package_uploaders (
    package_id INTEGER REFERENCES packages(id) ON DELETE CASCADE,
    uploader TEXT NOT NULL,
    PRIMARY KEY (package_id, uploader)
);

CREATE INDEX idx_packages_name ON packages(name);
CREATE INDEX idx_package_versions_package_id ON package_versions(package_id);