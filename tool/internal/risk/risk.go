// Package risk computes a risk level from a threat's likelihood and impact.
// The scoring method is pluggable via Scorer so a model can select, e.g., a
// qualitative matrix today and an ETSI attack-potential profile later.
package risk

import "sectrack/internal/model"

// Scorer turns a likelihood and an impact into a risk level.
type Scorer interface {
	Level(likelihood, impact string) string
}

// rank maps the qualitative scale to a weight; product bands give the level.
var rank = map[string]int{"low": 1, "medium": 2, "high": 3}

// Qualitative is a 3×3 likelihood×impact matrix → {low,medium,high,critical}.
// Unknown inputs yield "" so validation can flag them as out of scale.
type Qualitative struct{}

func (Qualitative) Level(likelihood, impact string) string {
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

// InScale reports whether v is a valid likelihood/impact value for the method.
func InScale(method, v string) bool {
	_, ok := rank[v] // qualitative scale; method reserved for future profiles
	return ok
}

// scorerFor returns the scorer for a policy method. Empty defaults to
// qualitative; unknown methods return nil so callers can report it.
func scorerFor(method string) Scorer {
	switch method {
	case "", "qualitative":
		return Qualitative{}
	default:
		return nil
	}
}

// Score returns each threat's computed risk level, keyed by threat ID.
// A threat with explicit likelihood+impact is scored by the policy's method;
// otherwise it falls back to its severity (back-compat with pre-risk models).
func Score(p *model.Project) map[string]string {
	s := scorerFor(p.RiskPolicy.Method)
	if s == nil {
		s = Qualitative{}
	}
	out := make(map[string]string, len(p.Threats))
	for _, t := range p.Threats {
		if t.Likelihood != "" && t.Impact != "" {
			out[t.ID] = s.Level(t.Likelihood, t.Impact)
		} else {
			out[t.ID] = t.Severity
		}
	}
	return out
}
