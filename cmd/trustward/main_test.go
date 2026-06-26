package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// buildBinary compiles trustward into a temp dir and returns its path.
func buildBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "trustward")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	return bin
}

func TestValidate_CleanModelExitsZero(t *testing.T) {
	bin := buildBinary(t)

	cmd := exec.Command(bin, "validate")
	cmd.Dir = "../../example/fire-protection-system"
	out, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("validate on example should succeed, got %v:\n%s", err, out)
	}
	if !strings.Contains(string(out), "model is consistent") {
		t.Errorf("output should confirm consistency, got:\n%s", out)
	}
}

func TestDiagram_EmitsMermaidFlowchart(t *testing.T) {
	bin := buildBinary(t)

	cmd := exec.Command(bin, "diagram", "dataflow")
	cmd.Dir = "../../example/fire-protection-system"
	out, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("diagram dataflow should succeed, got %v:\n%s", err, out)
	}
	if !strings.Contains(string(out), "flowchart TD") {
		t.Errorf("output should be a Mermaid flowchart, got:\n%s", out)
	}
}

func TestReport_RendersThreatModelDocument(t *testing.T) {
	bin := buildBinary(t)

	cmd := exec.Command(bin, "report")
	cmd.Dir = "../../example/fire-protection-system"
	out, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("report should succeed, got %v:\n%s", err, out)
	}
	if !strings.Contains(string(out), "Threat Model —") {
		t.Errorf("output should be the Quarto threat model document, got:\n%s", out)
	}
}

func TestTemplateExport_WritesThenRefusesToOverwrite(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()

	first := exec.Command(bin, "template", "export", "report")
	first.Dir = dir
	if out, err := first.CombinedOutput(); err != nil {
		t.Fatalf("first export should succeed, got %v:\n%s", err, out)
	}
	if _, err := os.Stat(filepath.Join(dir, "templates", "report.tmpl")); err != nil {
		t.Fatalf("export should write the template file: %v", err)
	}

	second := exec.Command(bin, "template", "export", "report")
	second.Dir = dir
	out, err := second.CombinedOutput()
	if err == nil {
		t.Fatalf("second export should refuse to overwrite, got success:\n%s", out)
	}
	if !strings.Contains(string(out), "already exists") {
		t.Errorf("output should explain the refusal, got:\n%s", out)
	}
}

func TestReport_PrefersProjectLocalTemplate(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "system.yaml"),
		[]byte("system:\n  id: s\n  title: My System\n  description: d\n"), 0644); err != nil {
		t.Fatalf("writing model: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "templates"), 0755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "templates", "report.tmpl"),
		[]byte("SENTINEL-TEMPLATE {{ .Title }}\n"), 0644); err != nil {
		t.Fatalf("writing template: %v", err)
	}

	cmd := exec.Command(bin, "report")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("report should succeed, got %v:\n%s", err, out)
	}
	if !strings.Contains(string(out), "SENTINEL-TEMPLATE My System") {
		t.Errorf("report should use the project-local template, got:\n%s", out)
	}
}

func TestValidate_BrokenModelExitsNonZeroAndNamesTheProblem(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	model := `
threats:
  - id: threat-x
    mitigations:
      - ctrl-missing
`
	if err := os.WriteFile(filepath.Join(dir, "system.yaml"), []byte(model), 0644); err != nil {
		t.Fatalf("writing model: %v", err)
	}

	cmd := exec.Command(bin, "validate")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatalf("want non-zero exit for broken model, got success:\n%s", out)
	}
	if !strings.Contains(string(out), "ctrl-missing") {
		t.Errorf("output should name the dangling control, got:\n%s", out)
	}
}
