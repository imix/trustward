package model

type SystemFile struct {
	Version    Version     `yaml:"version"`
	Imports    []Import    `yaml:"imports"`
	SystemMeta *SystemMeta `yaml:"system"`
	Assets     []Asset     `yaml:"assets"`
	Components []Component `yaml:"components"`
	TrustZones []TrustZone `yaml:"trust-zones"`
	DataFlows  []DataFlow  `yaml:"data-flows"`
}

type Version struct {
	Semver      string `yaml:"semver"`
	ReleaseDate string `yaml:"releasedate"`
}

type Import struct {
	Path    string `yaml:"path"`
	Version string `yaml:"version"`
}

type SystemMeta struct {
	ID          string `yaml:"id"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
}

type Asset struct {
	ID             string `yaml:"id"`
	Type           string `yaml:"type"`
	Classification string `yaml:"classification"`
	Description    string `yaml:"description"`
}

type Component struct {
	ID          string   `yaml:"id"`
	Type        string   `yaml:"type"`
	Assets      []string `yaml:"assets"`
	Description string   `yaml:"description"`
}

type TrustZone struct {
	ID          string   `yaml:"id"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Members     []string `yaml:"members"`
}

type DataFlow struct {
	ID          string   `yaml:"id"`
	Title       string   `yaml:"title"`
	Connects    []string `yaml:"connects"`
	Assets      []string `yaml:"assets"`
	Description string   `yaml:"description"`
}
