package risk

import (
	"testing"

	"github.com/imix/trustward/internal/model"
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

func TestAttackPotentialLevel(t *testing.T) {
	e := AttackPotential{}
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
		if got := e.Score(c.threat).Level; got != c.want {
			t.Errorf("%s: AttackPotential.Score().Level = %q, want %q", c.name, got, c.want)
		}
	}
}

// The derived likelihood must be surfaced, not discarded: an easy attack
// (sum 0) bands to "high", a hard one to "low".
func TestAttackPotentialLikelihoodSurfaced(t *testing.T) {
	atk := func(exp, kn, op, eq string) *model.AttackPotential {
		return &model.AttackPotential{Expertise: exp, Knowledge: kn, Opportunity: op, Equipment: eq}
	}
	easy := AttackPotential{}.Score(model.Threat{Impact: "high", Attack: atk("layman", "public", "unlimited", "standard")})
	if easy.Likelihood != "high" {
		t.Errorf("easy attack: want derived likelihood high, got %q", easy.Likelihood)
	}
	hard := AttackPotential{}.Score(model.Threat{Impact: "high", Attack: atk("expert", "critical", "difficult", "bespoke")})
	if hard.Likelihood != "low" {
		t.Errorf("hard attack: want derived likelihood low, got %q", hard.Likelihood)
	}
}

func TestEvaluateLevels(t *testing.T) {
	p := &model.Project{
		RiskPolicy: model.RiskPolicy{Method: "qualitative", Set: true},
		Threats: []model.Threat{
			{ID: "scored", Likelihood: "high", Impact: "high"}, // -> critical via matrix
			{ID: "legacy", Severity: "medium"},                 // -> medium via fallback
		},
	}
	got := Evaluate(p)
	if got["scored"].Level != "critical" {
		t.Errorf("scored: want critical, got %q", got["scored"].Level)
	}
	if got["scored"].Likelihood != "high" {
		t.Errorf("scored: want likelihood high, got %q", got["scored"].Likelihood)
	}
	if got["legacy"].Level != "medium" {
		t.Errorf("legacy (severity fallback): want medium, got %q", got["legacy"].Level)
	}
}

func TestEvaluate(t *testing.T) {
	p := &model.Project{
		RiskPolicy: model.RiskPolicy{Method: "qualitative", Accept: []string{"low"}, Set: true},
		Threats: []model.Threat{
			{ID: "accepted", Likelihood: "low", Impact: "low"},                                     // low ∈ accept
			{ID: "treated", Likelihood: "high", Impact: "high", Treatment: "mitigate", Owner: "a", Mitigations: []string{"c1"}}, // critical, mitigated by a control
			{ID: "plan", Likelihood: "high", Impact: "high", Treatment: "mitigate", Owner: "a"},                                 // mitigate, no control → still open
			{ID: "open", Likelihood: "high", Impact: "high"},                                                                    // critical, untreated
		},
	}
	e := Evaluate(p)
	if !e["accepted"].Accepted || e["accepted"].Open() {
		t.Errorf("accepted: %+v", e["accepted"])
	}
	if !e["treated"].Treated || e["treated"].Open() {
		t.Errorf("treated: %+v", e["treated"])
	}
	if e["plan"].Treated || !e["plan"].Open() {
		t.Errorf("mitigate without a control must stay open: %+v", e["plan"])
	}
	if !e["open"].Open() {
		t.Errorf("open should be open: %+v", e["open"])
	}
}

func TestScore_AttackPotentialMethod(t *testing.T) {
	p := &model.Project{
		RiskPolicy: model.RiskPolicy{Method: "attack-potential", Set: true},
		Threats: []model.Threat{
			{ID: "ap", Impact: "high",
				Attack: &model.AttackPotential{Expertise: "layman", Knowledge: "public", Opportunity: "unlimited", Equipment: "standard"}},
			{ID: "noattack", Severity: "low"}, // no attack block → severity fallback
		},
	}
	got := Evaluate(p)
	if got["ap"].Level != "critical" {
		t.Errorf("attack-potential-scored: want critical, got %q", got["ap"].Level)
	}
	if got["noattack"].Level != "low" {
		t.Errorf("no attack block: want severity fallback low, got %q", got["noattack"].Level)
	}
}
