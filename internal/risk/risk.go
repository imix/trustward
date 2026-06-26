// Package risk computes a risk level for a threat. The scoring method is
// pluggable via Scorer: a qualitative likelihood×impact matrix (default) or an
// ETSI attack-potential profile, selected by the project's risk-policy method.
package risk

import "github.com/imix/trustward/internal/model"

// Score is one threat's computed risk: the level, and the likelihood that
// produced it — declared (qualitative) or derived from attack potential
// (etsi-tvra). An empty Level means the inputs were absent or invalid, so the
// caller can fall back to the threat's severity.
type Score struct {
	Level      string
	Likelihood string
}

// Scorer computes the Score for a threat from whatever inputs its method
// needs (read from the threat).
type Scorer interface {
	Score(t model.Threat) Score
}

// rank maps the qualitative scale to a weight; product bands give the level.
var rank = map[string]int{"low": 1, "medium": 2, "high": 3}

// level is the shared likelihood×impact matrix → {low,medium,high,critical}.
// Both the qualitative and ETSI methods end here once they have a likelihood.
func level(likelihood, impact string) string {
	l, ok1 := rank[likelihood]
	i, ok2 := rank[impact]
	if !ok1 || !ok2 {
		return ""
	}
	switch p := l * i; {
	case p <= 2:
		return "low"
	case p <= 4:
		return "medium"
	case p <= 6:
		return "high"
	default:
		return "critical"
	}
}

// Qualitative scores from an explicit likelihood and impact on the threat.
type Qualitative struct{}

func (Qualitative) Score(t model.Threat) Score {
	lvl := level(t.Likelihood, t.Impact)
	if lvl == "" {
		return Score{} // unscorable → caller falls back to severity
	}
	return Score{Level: lvl, Likelihood: t.Likelihood}
}

// InScale reports whether v is a valid likelihood/impact value (qualitative scale).
func InScale(v string) bool {
	_, ok := rank[v]
	return ok
}

// MethodKnown reports whether a risk-policy method names a real scoring
// profile. Empty means the default (qualitative). The validator uses this to
// reject a typo'd method before it silently scores as qualitative.
func MethodKnown(method string) bool { return scorerFor(method) != nil }

// scorerFor returns the scorer for a policy method. Empty defaults to
// qualitative; unknown methods return nil so callers can report it.
func scorerFor(method string) Scorer {
	switch method {
	case "", "qualitative":
		return Qualitative{}
	case "etsi-tvra":
		return ETSI{}
	default:
		return nil
	}
}

// Eval is one threat's computed Score judged against the acceptance criteria
// (prEN 40000-1-2 §6.5.5). It embeds Score, so e.Level and e.Likelihood read
// straight through.
type Eval struct {
	Score         // computed level + likelihood
	Accepted bool // level is within the policy's acceptance criteria
	Treated  bool // a treatment decision and owner are recorded
}

// Open reports a risk that is neither accepted nor treated — a CRA gap.
func (e Eval) Open() bool { return !e.Accepted && !e.Treated }

// Evaluate scores every threat and judges it against the risk-policy's
// acceptance criteria — the single entry point shared by the validator (the
// CRA gate) and the report (risk register). The policy's method scores each
// threat; when it cannot (inputs absent or invalid), the threat's severity is
// used as the level (back-compat with pre-risk models).
func Evaluate(p *model.Project) map[string]Eval {
	s := scorerFor(p.RiskPolicy.Method)
	if s == nil {
		s = Qualitative{}
	}
	accept := make(map[string]bool, len(p.RiskPolicy.Accept))
	for _, lvl := range p.RiskPolicy.Accept {
		accept[lvl] = true
	}
	out := make(map[string]Eval, len(p.Threats))
	for _, t := range p.Threats {
		sc := s.Score(t)
		if sc.Level == "" {
			sc.Level = t.Severity
		}
		out[t.ID] = Eval{
			Score:    sc,
			Accepted: accept[sc.Level],
			Treated:  t.Treatment != "" && t.Owner != "",
		}
	}
	return out
}
