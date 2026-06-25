package quarto_test

import (
	"strings"
	"testing"

	"sectrack/internal/model"
	"sectrack/internal/quarto"
)

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Errorf("output does not contain %q\ngot:\n%s", want, got)
	}
}

func assertNotContains(t *testing.T, got, want string) {
	t.Helper()
	if strings.Contains(got, want) {
		t.Errorf("output should not contain %q\ngot:\n%s", want, got)
	}
}

func render(t *testing.T, proj *model.Project, diagram string, pdf bool) string {
	t.Helper()
	if proj == nil {
		proj = &model.Project{}
	}
	out, err := quarto.ThreatModel(proj, quarto.DefaultTemplate(), diagram, pdf)
	if err != nil {
		t.Fatalf("ThreatModel: %v", err)
	}
	return out
}

func TestThreatModel_RiskRegister(t *testing.T) {
	proj := &model.Project{
		RiskPolicy: model.RiskPolicy{Method: "qualitative", Accept: []string{"low"}, Set: true},
		Components: []model.Component{{ID: "central-unit", Title: "Central Unit"}},
		Threats: []model.Threat{{
			ID: "threat-x", Title: "Config Tampering", Target: "central-unit",
			Likelihood: "high", Impact: "high", // -> critical
			Treatment: "mitigate", Owner: "alice", ResidualRisk: "medium",
		}},
	}
	got := render(t, proj, "", false)

	assertContains(t, got, "Risk Register")
	assertContains(t, got, "threat-x")
	assertContains(t, got, "critical") // computed risk level
	assertContains(t, got, "mitigate") // treatment
	assertContains(t, got, "alice")    // owner
}

func TestThreatModel_CRASections(t *testing.T) {
	proj := &model.Project{
		RiskPolicy: model.RiskPolicy{Method: "qualitative", Accept: []string{"low"}, Set: true},
		Components: []model.Component{{ID: "cu", Title: "Central Unit"}},
		Threats: []model.Threat{
			{ID: "t-open", Title: "Untreated", Target: "cu", Likelihood: "high", Impact: "high"}, // critical, open
			{ID: "t-ok", Title: "Low", Target: "cu", Likelihood: "low", Impact: "low"},           // low, accepted
		},
	}
	got := render(t, proj, "", false)

	// §6.3 methodology + acceptance criteria
	assertContains(t, got, "Risk Acceptance Criteria")
	assertContains(t, got, "qualitative") // method
	// §6.5.5 evaluation status visible in the register
	assertContains(t, got, "accepted")
	assertContains(t, got, "open")
}

func TestThreatModel_FrontMatterMeta(t *testing.T) {
	proj := &model.Project{
		Version:    model.Version{Semver: "1.2.3", ReleaseDate: "2026-06-09"},
		SystemMeta: &model.SystemMeta{Title: "My System", Description: "A test system."},
	}

	got := render(t, proj, "", false)

	assertContains(t, got, `title: "Threat Model — My System"`)
	assertContains(t, got, `date: "2026-06-09"`)
	assertContains(t, got, `version: "1.2.3"`)
	assertContains(t, got, "A test system.")
}

func TestThreatModel_DiagramEmbedded(t *testing.T) {
	diagram := "flowchart TD\n    a --> b"

	got := render(t, nil, diagram, false)

	assertContains(t, got, "```{mermaid}")
	assertContains(t, got, "flowchart TD\n    a --> b")
	assertContains(t, got, "```")
}

func TestThreatModel_ThreatSummaryRow(t *testing.T) {
	proj := &model.Project{
		Threats: []model.Threat{
			{ID: "t-001", Title: "Spoof sensor", Target: "comp-a", Severity: "critical", ResidualRisk: "high"},
		},
	}

	got := render(t, proj, "", false)

	// threats are grouped by target; summary table omits target column
	assertContains(t, got, "### comp-a")
	assertContains(t, got, "| critical |")
	assertContains(t, got, "| Spoof sensor |")
	assertContains(t, got, "| high |")
}

func TestThreatModel_MitigationResolvesToControlTitle(t *testing.T) {
	proj := &model.Project{
		Threats: []model.Threat{
			{ID: "t-001", Title: "Threat", Mitigations: []string{"ctrl-iam"}},
		},
		Controls: []model.Control{
			{ID: "ctrl-iam", Title: "Identity and Access Management"},
		},
	}

	got := render(t, proj, "", false)

	assertContains(t, got, "Identity and Access Management (`ctrl-iam`)")
}

func TestThreatModel_UnknownMitigationFallsBackToID(t *testing.T) {
	proj := &model.Project{
		Threats: []model.Threat{
			{ID: "t-001", Title: "Threat", Mitigations: []string{"ctrl-unknown"}},
		},
	}

	got := render(t, proj, "", false)

	assertContains(t, got, "`ctrl-unknown`")
	assertNotContains(t, got, "Identity")
}

func TestThreatModel_NoMitigationsRendersNone(t *testing.T) {
	proj := &model.Project{
		Threats: []model.Threat{
			{ID: "t-001", Title: "Threat"},
		},
	}

	got := render(t, proj, "", false)

	assertContains(t, got, "| **Mitigations** | none |")
}

func TestThreatModel_PDFFalse_NoPDFSection(t *testing.T) {
	got := render(t, nil, "", false)

	assertNotContains(t, got, "pdf:")
}

func TestThreatModel_PDFTrue_PDFSectionPresent(t *testing.T) {
	got := render(t, nil, "", true)

	assertContains(t, got, "pdf:")
}
