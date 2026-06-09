package model

type ThreatModelFile struct {
	Version Version  `yaml:"version"`
	Imports []Import `yaml:"imports"`
	Threats []Threat `yaml:"threats"`
}

type Threat struct {
	ID           string   `yaml:"id"`
	Title        string   `yaml:"title"`
	Type         string   `yaml:"type"`
	Asset        string   `yaml:"asset"`
	Target       string   `yaml:"target"`
	Severity     string   `yaml:"severity"`
	Mitigations  []string `yaml:"mitigations"`
	ResidualRisk string   `yaml:"residualRisk"`
	Notes        string   `yaml:"notes"`
}
