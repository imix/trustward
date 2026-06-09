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

func render(t *testing.T, meta quarto.ReportMeta, tm *model.ThreatModelFile, company *model.CompanyFile, diagram string, pdf bool) string {
	t.Helper()
	if tm == nil {
		tm = &model.ThreatModelFile{}
	}
	if company == nil {
		company = &model.CompanyFile{}
	}
	out, err := quarto.ThreatModel(meta, tm, company, diagram, pdf)
	if err != nil {
		t.Fatalf("ThreatModel: %v", err)
	}
	return out
}

func TestThreatModel_FrontMatterMeta(t *testing.T) {
	meta := quarto.ReportMeta{
		Title:       "My System",
		Date:        "2026-06-09",
		Version:     "1.2.3",
		Description: "A test system.",
	}

	got := render(t, meta, nil, nil, "", false)

	assertContains(t, got, `title: "Threat Model — My System"`)
	assertContains(t, got, `date: "2026-06-09"`)
	assertContains(t, got, `version: "1.2.3"`)
	assertContains(t, got, "A test system.")
}

func TestThreatModel_DiagramEmbedded(t *testing.T) {
	diagram := "flowchart TD\n    a --> b"

	got := render(t, quarto.ReportMeta{}, nil, nil, diagram, false)

	assertContains(t, got, "```{mermaid}")
	assertContains(t, got, "flowchart TD\n    a --> b")
	assertContains(t, got, "```")
}

func TestThreatModel_ThreatSummaryRow(t *testing.T) {
	tm := &model.ThreatModelFile{
		Threats: []model.Threat{
			{ID: "t-001", Title: "Spoof sensor", Target: "comp-a", Severity: "critical", ResidualRisk: "high"},
		},
	}

	got := render(t, quarto.ReportMeta{}, tm, nil, "", false)

	assertContains(t, got, "| critical | t-001 | Spoof sensor | comp-a | high |")
}

func TestThreatModel_MitigationResolvesToControlTitle(t *testing.T) {
	tm := &model.ThreatModelFile{
		Threats: []model.Threat{
			{ID: "t-001", Title: "Threat", Mitigations: []string{"ctrl-iam"}},
		},
	}
	company := &model.CompanyFile{
		Controls: []model.Control{
			{ID: "ctrl-iam", Title: "Identity and Access Management"},
		},
	}

	got := render(t, quarto.ReportMeta{}, tm, company, "", false)

	assertContains(t, got, "Identity and Access Management (`ctrl-iam`)")
}

func TestThreatModel_UnknownMitigationFallsBackToID(t *testing.T) {
	tm := &model.ThreatModelFile{
		Threats: []model.Threat{
			{ID: "t-001", Title: "Threat", Mitigations: []string{"ctrl-unknown"}},
		},
	}

	got := render(t, quarto.ReportMeta{}, tm, nil, "", false)

	assertContains(t, got, "`ctrl-unknown`")
	assertNotContains(t, got, "Identity")
}

func TestThreatModel_NoMitigationsRendersNone(t *testing.T) {
	tm := &model.ThreatModelFile{
		Threats: []model.Threat{
			{ID: "t-001", Title: "Threat"},
		},
	}

	got := render(t, quarto.ReportMeta{}, tm, nil, "", false)

	assertContains(t, got, "| **Mitigations** | none |")
}

func TestThreatModel_PDFFalse_NoPDFSection(t *testing.T) {
	got := render(t, quarto.ReportMeta{}, nil, nil, "", false)

	assertNotContains(t, got, "pdf:")
}

func TestThreatModel_PDFTrue_PDFSectionPresent(t *testing.T) {
	got := render(t, quarto.ReportMeta{}, nil, nil, "", true)

	assertContains(t, got, "pdf:")
}
