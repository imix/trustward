package model

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
