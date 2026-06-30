package schema

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The committed example models must pass the schema cleanly — a false positive
// here would break every user's CI.
func TestExamplesValidate(t *testing.T) {
	for _, dir := range []string{
		"../example/fire-protection-system",
		"../example/access-control",
	} {
		if msgs := Check(dir); len(msgs) > 0 {
			t.Errorf("%s: expected no schema issues, got:\n%s", dir, strings.Join(msgs, "\n"))
		}
	}
}

// The closed vocabularies the schema now solely owns (consolidated out of
// internal/validate): each typo must be caught, named by the offending field.
func TestCatchesVocabularyErrors(t *testing.T) {
	cases := []struct {
		name, body, want string
	}{
		{"severity", "threats: [{id: t, severity: hgh}]", "severity"},
		{"likelihood", "threats: [{id: t, likelihood: probably}]", "likelihood"},
		{"impact", "threats: [{id: t, impact: huge}]", "impact"},
		{"treatment", "threats: [{id: t, treatment: frobnicate}]", "treatment"},
		{"residualRisk", "threats: [{id: t, residualRisk: meh}]", "residualRisk"},
		{"attack factor", "threats: [{id: t, attack: {expertise: wizard}}]", "expertise"},
		{"objective type", "objectives: [{id: o, type: speed}]", "type"},
		{"risk-policy method", "risk-policy: {method: qualitatve}", "method"},
		{"data-flow arity", "data-flows: [{id: f, connects: [a]}]", "connects"},
		{"reference version", "references: [{id: r, title: x}]", "version"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "system.yaml"), []byte(tc.body), 0644); err != nil {
				t.Fatal(err)
			}
			msgs := Check(dir)
			if !strings.Contains(strings.Join(msgs, "\n"), tc.want) {
				t.Errorf("expected a message mentioning %q, got: %v", tc.want, msgs)
			}
		})
	}
}

// The mapping form of threats (ignored vocabulary) must not trip the item checks.
func TestThreatsMappingAllowed(t *testing.T) {
	dir := t.TempDir()
	const vocab = `
threats:
  types:
    - {id: spoofing, title: Spoofing}
`
	if err := os.WriteFile(filepath.Join(dir, "system.yaml"), []byte(vocab), 0644); err != nil {
		t.Fatal(err)
	}
	if msgs := Check(dir); len(msgs) > 0 {
		t.Errorf("threats mapping should be allowed, got: %v", msgs)
	}
}
