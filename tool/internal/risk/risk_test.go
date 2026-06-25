package risk

import (
	"testing"

	"sectrack/internal/model"
)

func TestLevelMatrix(t *testing.T) {
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
		if got := level(c.likelihood, c.impact); got != c.want {
			t.Errorf("level(%q,%q) = %q, want %q", c.likelihood, c.impact, got, c.want)
		}
	}
}

func TestETSILevel(t *testing.T) {
	e := ETSI{}
	atk := func(exp, kn, op, eq string) *model.AttackPotential {
		return &model.AttackPotential{Expertise: exp, Knowledge: kn, Opportunity: op, Equipment: eq}
	}
	cases := []struct {
		name   string
		threat model.Threat
		want   string
	}{
		// easy attack (sum 0) → high likelihood × high impact → critical
		{"easy/high-impact", model.Threat{Impact: "high",
			Attack: atk("layman", "public", "unlimited", "standard")}, "critical"},
		// hard attack (expert+critical+difficult+bespoke = 6+11+10+7 = 34) → low likelihood × high impact → medium
		{"hard/high-impact", model.Threat{Impact: "high",
			Attack: atk("expert", "critical", "difficult", "bespoke")}, "medium"},
		// no attack block → unscored
		{"no-attack", model.Threat{Impact: "high"}, ""},
		// invalid factor → unscored
		{"bad-factor", model.Threat{Impact: "high",
			Attack: atk("wizard", "public", "easy", "standard")}, ""},
	}
	for _, c := range cases {
		if got := e.Level(c.threat); got != c.want {
			t.Errorf("%s: ETSI.Level = %q, want %q", c.name, got, c.want)
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

func TestEvaluate(t *testing.T) {
	p := &model.Project{
		RiskPolicy: model.RiskPolicy{Method: "qualitative", Accept: []string{"low"}, Set: true},
		Threats: []model.Threat{
			{ID: "accepted", Likelihood: "low", Impact: "low"},                                     // low ∈ accept
			{ID: "treated", Likelihood: "high", Impact: "high", Treatment: "mitigate", Owner: "a"}, // critical, but treated
			{ID: "open", Likelihood: "high", Impact: "high"},                                       // critical, untreated
		},
	}
	e := Evaluate(p)
	if !e["accepted"].Accepted || e["accepted"].Open() {
		t.Errorf("accepted: %+v", e["accepted"])
	}
	if !e["treated"].Treated || e["treated"].Open() {
		t.Errorf("treated: %+v", e["treated"])
	}
	if !e["open"].Open() {
		t.Errorf("open should be open: %+v", e["open"])
	}
}

func TestScore_ETSIMethod(t *testing.T) {
	p := &model.Project{
		RiskPolicy: model.RiskPolicy{Method: "etsi-tvra", Set: true},
		Threats: []model.Threat{
			{ID: "etsi", Impact: "high",
				Attack: &model.AttackPotential{Expertise: "layman", Knowledge: "public", Opportunity: "unlimited", Equipment: "standard"}},
			{ID: "noattack", Severity: "low"}, // no attack block → severity fallback
		},
	}
	got := Score(p)
	if got["etsi"] != "critical" {
		t.Errorf("etsi-scored: want critical, got %q", got["etsi"])
	}
	if got["noattack"] != "low" {
		t.Errorf("no attack block: want severity fallback low, got %q", got["noattack"])
	}
}
