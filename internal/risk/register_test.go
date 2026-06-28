package risk

import (
	"testing"

	"github.com/imix/trustward/internal/model"
)

// The register's job beyond a field copy is the accepted/treated/open join,
// driven by Evaluate against the policy. One scored-and-open, one accepted,
// one treated covers the three states.
func TestRegister_OpenTreatedAccepted(t *testing.T) {
	p := &model.Project{
		RiskPolicy: model.RiskPolicy{Method: "qualitative", Accept: []string{"low"}, Set: true},
		Threats: []model.Threat{
			{ID: "open", Likelihood: "high", Impact: "high"},                                 // critical, not accepted, untreated
			{ID: "accepted", Likelihood: "low", Impact: "low"},                               // low → accepted by policy
			{ID: "treated", Likelihood: "high", Impact: "high", Treatment: "mitigate", Owner: "Sec", Mitigations: []string{"c1"}}, // critical but treated by a control
		},
	}
	got := map[string]Entry{}
	for _, e := range Register(p) {
		got[e.ID] = e
	}

	if e := got["open"]; !e.Open || e.Accepted || e.Treated || e.Level != "critical" {
		t.Errorf("open: got %+v", e)
	}
	if e := got["accepted"]; e.Open || !e.Accepted || e.Level != "low" {
		t.Errorf("accepted: got %+v", e)
	}
	if e := got["treated"]; e.Open || !e.Treated || e.Owner != "Sec" {
		t.Errorf("treated: got %+v", e)
	}
}
