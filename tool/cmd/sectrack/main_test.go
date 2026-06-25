package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// buildBinary compiles sectrack into a temp dir and returns its path.
func buildBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "sectrack")
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
	cmd.Dir = "../../../example/fire-protection-system"
	out, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("validate on example should succeed, got %v:\n%s", err, out)
	}
	if !strings.Contains(string(out), "model is consistent") {
		t.Errorf("output should confirm consistency, got:\n%s", out)
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
