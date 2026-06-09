package model

type Version struct {
	Semver      string `yaml:"semver"`
	ReleaseDate string `yaml:"releasedate"`
}

type SystemMeta struct {
	ID          string `yaml:"id"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Logo        string `yaml:"logo"`
}

type Asset struct {
	ID             string `yaml:"id"`
	Type           string `yaml:"type"`
	Classification string `yaml:"classification"`
	Description    string `yaml:"description"`
}

type Component struct {
	ID          string   `yaml:"id"`
	Title       string   `yaml:"title"`
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

type Requirement struct {
	ID          string   `yaml:"id"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Satisfies   []string `yaml:"satisfies"`
}

type ControlCatalog struct {
	ID           string        `yaml:"id"`
	Title        string        `yaml:"title"`
	Requirements []Requirement `yaml:"requirements"`
}

type Control struct {
	ID          string   `yaml:"id"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Ref         string   `yaml:"ref"`
	Evidence    []string `yaml:"evidence"`
}

type ThreatPattern struct {
	ID       string `yaml:"id"`
	Title    string `yaml:"title"`
	Type     string `yaml:"type"`
	Severity string `yaml:"severity"`
	Notes    string `yaml:"notes"`
}

type ThreatCatalog struct {
	ID       string          `yaml:"id"`
	Title    string          `yaml:"title"`
	Patterns []ThreatPattern `yaml:"patterns"`
}

type Threat struct {
	ID           string   `yaml:"id"`
	Ref          string   `yaml:"ref"`
	Title        string   `yaml:"title"`
	Type         string   `yaml:"type"`
	Asset        string   `yaml:"asset"`
	Target       string   `yaml:"target"`
	Severity     string   `yaml:"severity"`
	Mitigations  []string `yaml:"mitigations"`
	ResidualRisk string   `yaml:"residualRisk"`
	Notes        string   `yaml:"notes"`
}
