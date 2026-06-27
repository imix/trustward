package quarto_test

import (
	"strings"
	"testing"

	"github.com/imix/trustward/internal/model"
	"github.com/imix/trustward/internal/quarto"
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
	out, err := quarto.Report(proj, quarto.DefaultTemplate(), diagram, pdf)
	if err != nil {
		t.Fatalf("Report: %v", err)
	}
	return out
}

func TestReport_SectionStructure(t *testing.T) {
	proj := &model.Project{
		RiskPolicy: model.RiskPolicy{Method: "qualitative", Accept: []string{"low"}, Set: true},
		Assets:     []model.Asset{{ID: "a", Type: "data"}},
		Components: []model.Component{{ID: "cu", Title: "Central Unit"}},
		Controls:   []model.Control{{ID: "ctrl-a", Title: "Control A"}},
		Threats:    []model.Threat{{ID: "t", Title: "T", Target: "cu", Likelihood: "low", Impact: "low"}},
	}

	got := render(t, proj, "flowchart TD", false)

	// Sections follow the prEN 40000-1-2 §6 process but stay standard-agnostic:
	// no hardcoded clause numbers (Pandoc numbers sections when enabled).
	ordered := []string{
		"## Product Context",
		"## Risk Acceptance Criteria and Methodology",
		"## Risk Assessment",
		"### Asset and Cybersecurity Objective Identification",
		"### Threat Identification",
		"### Risk Register",
		"## Risk Treatment",
	}
	last := -1
	for _, h := range ordered {
		i := strings.Index(got, h)
		if i < 0 {
			t.Errorf("missing heading %q", h)
			continue
		}
		if i < last {
			t.Errorf("heading %q is out of clause order", h)
		}
		last = i
	}

	// A heading needs a blank line before it or Pandoc (blank_before_header)
	// folds it into the preceding paragraph. RiskPolicySet is true here, so the
	// acceptance paragraph sits directly above Risk Assessment — guard the gap.
	if !strings.Contains(got, "\n\n## Risk Assessment") {
		t.Error("## Risk Assessment must be preceded by a blank line")
	}
}

func TestReport_RiskRegister(t *testing.T) {
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

func TestReport_CRASections(t *testing.T) {
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

func TestReport_CybersecurityObjectives(t *testing.T) {
	proj := &model.Project{
		Objectives: []model.Objective{
			{ID: "obj-conf", Title: "Reading confidentiality", Type: "confidentiality", Description: "No leaks"},
		},
		Assets: []model.Asset{
			{ID: "asset-readings", Type: "telemetry", Objectives: []string{"obj-conf"}},
		},
	}

	got := render(t, proj, "", false)

	assertContains(t, got, "Cybersecurity Objectives") // §6.5.2 section
	assertContains(t, got, "obj-conf")
	assertContains(t, got, "Reading confidentiality") // title
	assertContains(t, got, "confidentiality")         // CIA type
	assertContains(t, got, "asset-readings")          // upheld-by trace
}

func TestReport_NoObjectivesNoSection(t *testing.T) {
	got := render(t, &model.Project{Assets: []model.Asset{{ID: "a", Type: "data"}}}, "", false)

	// the §6.5.2 clause heading is always present; the objectives subsection is not
	assertNotContains(t, got, "#### Cybersecurity Objectives")
}

func TestReport_MonitoringAndReview(t *testing.T) {
	proj := &model.Project{
		RiskPolicy: model.RiskPolicy{
			Method: "qualitative", Accept: []string{"low"}, Set: true,
			Review: "Reviewed quarterly by the OT security lead.",
		},
	}

	got := render(t, proj, "", false)

	assertContains(t, got, "## Risk Monitoring and Review")
	assertContains(t, got, "Reviewed quarterly by the OT security lead.")
}

func TestReport_MonitoringAndReviewEmpty(t *testing.T) {
	// Policy set but no review cadence: the clause heading still renders (like other
	// narrative sections), with no special-case placeholder text.
	proj := &model.Project{
		RiskPolicy: model.RiskPolicy{Method: "qualitative", Accept: []string{"low"}, Set: true},
	}

	got := render(t, proj, "", false)

	assertContains(t, got, "## Risk Monitoring and Review")
	assertNotContains(t, got, "No risk monitoring and review cadence")
}

func TestReport_FrontMatterMeta(t *testing.T) {
	proj := &model.Project{
		Version:    model.Version{Semver: "1.2.3", ReleaseDate: "2026-06-09"},
		SystemMeta: &model.SystemMeta{Title: "My System", Description: "A test system."},
	}

	got := render(t, proj, "", false)

	assertContains(t, got, `title: "Threat Model — My System"`)
	assertContains(t, got, `date: "2026-06-09"`)
	assertContains(t, got, `subtitle: "Version 1.2.3"`)
	assertContains(t, got, "A test system.")
}

func TestReport_DiagramEmbedded(t *testing.T) {
	diagram := "flowchart TD\n    a --> b"

	got := render(t, nil, diagram, false)

	assertContains(t, got, "```{mermaid}")
	assertContains(t, got, "flowchart TD\n    a --> b")
	assertContains(t, got, "```")
}

func TestReport_ThreatSummaryRow(t *testing.T) {
	proj := &model.Project{
		Threats: []model.Threat{
			{ID: "t-001", Title: "Spoof sensor", Target: "comp-a", Severity: "critical", ResidualRisk: "high"},
		},
	}

	got := render(t, proj, "", false)

	// threats are grouped by target under §6.5.3; summary table omits target column
	assertContains(t, got, "#### comp-a")
	assertContains(t, got, "| critical |")
	assertContains(t, got, "| Spoof sensor |")
	assertContains(t, got, "| high |")
}

func TestReport_FlowTargetUsesFlowTitle(t *testing.T) {
	// A threat targeting a data flow groups under the flow's title, not its raw ID.
	proj := &model.Project{
		DataFlows: []model.DataFlow{{ID: "flow-ocpp", Title: "OCPP control channel"}},
		Threats:   []model.Threat{{ID: "t-1", Title: "Rogue server", Target: "flow-ocpp"}},
	}

	got := render(t, proj, "", false)

	assertContains(t, got, "#### OCPP control channel")
	assertNotContains(t, got, "#### flow-ocpp")
}

func TestReport_MitigationResolvesToControlTitle(t *testing.T) {
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

func TestReport_UnknownMitigationFallsBackToID(t *testing.T) {
	proj := &model.Project{
		Threats: []model.Threat{
			{ID: "t-001", Title: "Threat", Mitigations: []string{"ctrl-unknown"}},
		},
	}

	got := render(t, proj, "", false)

	assertContains(t, got, "`ctrl-unknown`")
	assertNotContains(t, got, "Identity")
}

func TestReport_NoMitigationsRendersNone(t *testing.T) {
	proj := &model.Project{
		Threats: []model.Threat{
			{ID: "t-001", Title: "Threat"},
		},
	}

	got := render(t, proj, "", false)

	assertContains(t, got, "| **Mitigations** | none |")
}

func TestReport_PDFFalse_NoPDFSection(t *testing.T) {
	got := render(t, nil, "", false)

	assertNotContains(t, got, "pdf:")
}

func TestReport_PDFTrue_PDFSectionPresent(t *testing.T) {
	got := render(t, nil, "", true)

	assertContains(t, got, "pdf:")
}
