// Package risk computes a risk level for a threat. The scoring method is
// pluggable via Scorer: a qualitative likelihood×impact matrix (default) or an
// ETSI attack-potential profile, selected by the project's risk-policy method.
package risk

import "sectrack/internal/model"

// Scorer computes the risk level for a threat from whatever inputs its method
// needs (read from the threat). It returns "" when the inputs are absent or
// invalid, so Score can fall back to the threat's severity.
type Scorer interface {
	Level(t model.Threat) string
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

func (Qualitative) Level(t model.Threat) string { return level(t.Likelihood, t.Impact) }

// InScale reports whether v is a valid likelihood/impact value (qualitative scale).
func InScale(method, v string) bool {
	_, ok := rank[v] // method reserved; only the qualitative scale exists here
	return ok
}

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

// Score returns each threat's computed risk level, keyed by threat ID.
// The policy's method scores each threat; when it cannot (inputs absent or
// invalid), the threat's severity is used (back-compat with pre-risk models).
func Score(p *model.Project) map[string]string {
	s := scorerFor(p.RiskPolicy.Method)
	if s == nil {
		s = Qualitative{}
	}
	out := make(map[string]string, len(p.Threats))
	for _, t := range p.Threats {
		lvl := s.Level(t)
		if lvl == "" {
			lvl = t.Severity
		}
		out[t.ID] = lvl
	}
	return out
}
