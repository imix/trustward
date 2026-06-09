package model

type Version struct {
	Semver      string `yaml:"semver"`
	ReleaseDate string `yaml:"releasedate"`
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
	Controls    []string `yaml:"controls"`
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
