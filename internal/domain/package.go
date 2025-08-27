package domain

import "time"

type Package struct {
	ID            int32     `json:"id"`
	Name          string    `json:"name"`
	Private       bool      `json:"private"`
	Description   *string   `json:"description"`
	Homepage      *string   `json:"homepage"`
	Repository    *string   `json:"repository"`
	Documentation *string   `json:"documentation"`
	Topics        []string  `json:"topics"`
	DownloadCount int64     `json:"download_count"`
	LikeCount     int32     `json:"like_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type PackageVersion struct {
	ID            int32     `json:"id"`
	PackageID     int32     `json:"package_id"`
	Version       string    `json:"version"`
	Description   *string   `json:"description"`
	PubspecYaml   string    `json:"pubspec_yaml"`
	Readme        *string   `json:"readme"`
	Changelog     *string   `json:"changelog"`
	ArchivePath   string    `json:"archive_path"`
	ArchiveSha256 *string   `json:"archive_sha256"`
	Uploader      *string   `json:"uploader"`
	Retracted     bool      `json:"retracted"`
	DownloadCount int64     `json:"download_count"`
	CreatedAt     time.Time `json:"created_at"`
}

type PackageResponse struct {
	Name           string            `json:"name"`
	IsDiscontinued bool              `json:"isDiscontinued,omitempty"`
	Latest         VersionResponse   `json:"latest"`
	Versions       []VersionResponse `json:"versions"`
}

// Extended package info for UI display
type PackageDetail struct {
	Package      *Package        `json:"package"`
	Latest       *PackageVersion `json:"latest"`
	Versions     []*PackageVersion `json:"versions"`
	TotalDownloads int64         `json:"total_downloads"`
}

type VersionResponse struct {
	Version       string         `json:"version"`
	Retracted     bool           `json:"retracted,omitempty"`
	ArchiveURL    string         `json:"archive_url"`
	ArchiveSha256 string         `json:"archive_sha256,omitempty"`
	Pubspec       map[string]any `json:"pubspec"`
}

type PublishRequest struct {
	Archive  []byte
	Uploader string
}

type PublishResponse struct {
	URL    string            `json:"url"`
	Fields map[string]string `json:"fields"`
}

type AdvisoriesResponse struct {
	Advisories        []Advisory `json:"advisories"`
	AdvisoriesUpdated string     `json:"advisoriesUpdated"`
}

type Advisory struct {
	ID               string   `json:"id"`
	Aliases          []string `json:"aliases,omitempty"`
	Summary          string   `json:"summary"`
	Details          string   `json:"details,omitempty"`
	Modified         string   `json:"modified"`
	Published        string   `json:"published,omitempty"`
	DatabaseSpecific struct {
		Severity string `json:"severity,omitempty"`
	} `json:"database_specific,omitempty"`
	Affected []struct {
		Package struct {
			Name      string `json:"name"`
			Ecosystem string `json:"ecosystem"`
		} `json:"package"`
		Ranges []struct {
			Type   string `json:"type"`
			Events []struct {
				Introduced string `json:"introduced,omitempty"`
				Fixed      string `json:"fixed,omitempty"`
			} `json:"events"`
		} `json:"ranges"`
		Versions []string `json:"versions,omitempty"`
	} `json:"affected,omitempty"`
}
