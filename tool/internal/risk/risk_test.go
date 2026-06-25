package risk

import (
	"testing"

	"sectrack/internal/model"
)

func TestQualitativeLevel(t *testing.T) {
	q := Qualitative{}
	cases := []struct {
		likelihood, impact, want string
	}{
		{"low", "low", "low"},
		{"low", "high", "medium"},
		{"high", "low", "medium"},
		{"medium", "medium", "medium"},
		{"medium", "high", "high"},
		{"high", "high", "critical"},
		{"bogus", "high", ""}, // unknown input -> no level
	}
	for _, c := range cases {
		if got := q.Level(c.likelihood, c.impact); got != c.want {
			t.Errorf("Level(%q,%q) = %q, want %q", c.likelihood, c.impact, got, c.want)
		}
	}
}

func TestScore(t *testing.T) {
	p := &model.Project{
		RiskPolicy: model.RiskPolicy{Method: "qualitative", Set: true},
		Threats: []model.Threat{
			{ID: "scored", Likelihood: "high", Impact: "high"}, // -> critical via matrix
			{ID: "legacy", Severity: "medium"},                 // -> medium via fallback
		},
	}
	got := Score(p)
	if got["scored"] != "critical" {
		t.Errorf("scored: want critical, got %q", got["scored"])
	}
	if got["legacy"] != "medium" {
		t.Errorf("legacy (severity fallback): want medium, got %q", got["legacy"])
	}
}
