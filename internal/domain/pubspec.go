package domain

// Pubspec represents a parsed pubspec.yaml file
type Pubspec struct {
	Name            string                 `json:"name" yaml:"name"`
	Version         string                 `json:"version" yaml:"version"`
	Description     string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Homepage        string                 `json:"homepage,omitempty" yaml:"homepage,omitempty"`
	Repository      string                 `json:"repository,omitempty" yaml:"repository,omitempty"`
	IssueTracker    string                 `json:"issue_tracker,omitempty" yaml:"issue_tracker,omitempty"`
	Documentation   string                 `json:"documentation,omitempty" yaml:"documentation,omitempty"`
	Dependencies    map[string]interface{} `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	DevDependencies map[string]interface{} `json:"dev_dependencies,omitempty" yaml:"dev_dependencies,omitempty"`
	Environment     *Environment           `json:"environment,omitempty" yaml:"environment,omitempty"`
	Executables     map[string]string      `json:"executables,omitempty" yaml:"executables,omitempty"`
	PublishTo       string                 `json:"publish_to,omitempty" yaml:"publish_to,omitempty"`
	Author          string                 `json:"author,omitempty" yaml:"author,omitempty"`
	Authors         []string               `json:"authors,omitempty" yaml:"authors,omitempty"`
	Funding         []string               `json:"funding,omitempty" yaml:"funding,omitempty"`
	Screenshots     []Screenshot           `json:"screenshots,omitempty" yaml:"screenshots,omitempty"`
	Topics          []string               `json:"topics,omitempty" yaml:"topics,omitempty"`
	Platforms       map[string]interface{} `json:"platforms,omitempty" yaml:"platforms,omitempty"`
	// Additional fields that might be present
	Extra map[string]interface{} `json:",inline" yaml:",inline"`
}

type Environment struct {
	SDK     string `json:"sdk,omitempty" yaml:"sdk,omitempty"`
	Flutter string `json:"flutter,omitempty" yaml:"flutter,omitempty"`
}

type Screenshot struct {
	Description string `json:"description" yaml:"description"`
	Path        string `json:"path" yaml:"path"`
}

// Dependency represents a package dependency
type Dependency struct {
	Version     string                 `json:"version,omitempty"`
	Hosted      string                 `json:"hosted,omitempty"`
	Git         *GitDependency         `json:"git,omitempty"`
	Path        string                 `json:"path,omitempty"`
	SDK         string                 `json:"sdk,omitempty"`
	Extra       map[string]interface{} `json:",inline"`
}

type GitDependency struct {
	URL  string `json:"url"`
	Ref  string `json:"ref,omitempty"`
	Path string `json:"path,omitempty"`
}

