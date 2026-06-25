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
	ID             string   `yaml:"id"`
	Type           string   `yaml:"type"`
	Classification string   `yaml:"classification"`
	Objectives     []string `yaml:"objectives"` // cybersecurity objectives this asset must uphold
	Description    string   `yaml:"description"`
}

// Objective is a cybersecurity objective an asset must uphold (prEN 40000-1-2
// §6.5.2). Type names a CIA-scale property the objective protects.
type Objective struct {
	ID          string `yaml:"id"`
	Title       string `yaml:"title"`
	Type        string `yaml:"type"` // confidentiality|integrity|availability|authenticity|accountability
	Description string `yaml:"description"`
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
	ID           string           `yaml:"id"`
	Ref          string           `yaml:"ref"`
	Title        string           `yaml:"title"`
	Type         string           `yaml:"type"`
	Asset        string           `yaml:"asset"`
	Target       string           `yaml:"target"`
	Violates     []string         `yaml:"violates"` // cybersecurity objectives this threat violates
	Severity     string           `yaml:"severity"`
	Likelihood   string           `yaml:"likelihood"` // qualitative: low|medium|high
	Impact       string           `yaml:"impact"`     // qualitative: low|medium|high
	Treatment    string           `yaml:"treatment"`  // mitigate|accept|transfer|avoid
	Owner        string           `yaml:"owner"`      // who signed off the treatment decision
	Decided      string           `yaml:"decided"`    // ISO sign-off date
	Attack       *AttackPotential `yaml:"attack"`     // ETSI attack-potential factors (etsi-tvra method)
	Mitigations  []string         `yaml:"mitigations"`
	ResidualRisk string           `yaml:"residualRisk"`
	Notes        string           `yaml:"notes"`
}

// AttackPotential holds the ETSI TS 102 165-1 attacker factors (clause 6.6.3),
// used by the etsi-tvra scoring method to derive a likelihood.
type AttackPotential struct {
	Expertise   string `yaml:"expertise"`
	Knowledge   string `yaml:"knowledge"`
	Opportunity string `yaml:"opportunity"`
	Equipment   string `yaml:"equipment"`
}

// RiskPolicy declares the scoring method and the risk acceptance criteria
// (CRA / prEN 40000-1-2 §6.3). When present, the CRA gate applies: any risk
// whose computed level is not in Accept must carry a treatment + owner.
type RiskPolicy struct {
	Method string   `yaml:"method"` // scoring profile; "" = qualitative
	Accept []string `yaml:"accept"` // risk levels acceptable without treatment
	Review string   `yaml:"review"` // monitoring and review cadence (§6.7)
	Set    bool     `yaml:"-"`      // true once a risk-policy: block was loaded
}
