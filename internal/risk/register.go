package risk

import "github.com/imix/trustward/internal/model"

// Entry is one threat in the machine-readable risk register: the threat's
// identifying fields joined with its computed evaluation. The JSON field names
// are the export contract — renaming one breaks downstream consumers, the same
// way renaming a reportData field breaks templates.
type Entry struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Type         string   `json:"type,omitempty"`
	Target       string   `json:"target,omitempty"`
	Asset        string   `json:"asset,omitempty"`
	Violates     []string `json:"violates,omitempty"`
	BackedBy     []string `json:"backedBy,omitempty"`
	Severity     string   `json:"severity,omitempty"`
	Level        string   `json:"level"` // computed risk level (falls back to severity)
	Likelihood   string   `json:"likelihood,omitempty"`
	Accepted     bool     `json:"accepted"`
	Treated      bool     `json:"treated"`
	Open         bool     `json:"open"` // neither accepted nor treated — a CRA gap
	Treatment    string   `json:"treatment,omitempty"`
	Owner        string   `json:"owner,omitempty"`
	Decided      string   `json:"decided,omitempty"`
	Mitigations  []string `json:"mitigations,omitempty"`
	ResidualRisk string   `json:"residualRisk,omitempty"`
}

// Register returns the scored risk register in threat order — the
// machine-readable view of the threat analysis, sharing risk.Evaluate with the
// report and the CRA gate so the numbers always agree.
func Register(p *model.Project) []Entry {
	evals := Evaluate(p)
	out := make([]Entry, 0, len(p.Threats))
	for _, t := range p.Threats {
		e := evals[t.ID]
		out = append(out, Entry{
			ID: t.ID, Title: t.Title, Type: t.Type, Target: t.Target, Asset: t.Asset,
			Violates: t.Violates, BackedBy: t.BackedBy, Severity: t.Severity,
			Level: e.Level, Likelihood: e.Likelihood,
			Accepted: e.Accepted, Treated: e.Treated, Open: e.Open(),
			Treatment: t.Treatment, Owner: t.Owner, Decided: t.Decided,
			Mitigations: t.Mitigations, ResidualRisk: t.ResidualRisk,
		})
	}
	return out
}
