package project_test

import (
	"os"
	"path/filepath"
	"testing"

	"sectrack/internal/project"
)

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatalf("writeFile %s: %v", name, err)
	}
}

func TestLoad_SystemContent(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "system.yaml", `
assets:
  - id: asset-a
    type: data
components:
  - id: comp-a
    type: server
trust-zones:
  - id: zone-a
    title: Zone A
data-flows:
  - id: flow-a
    title: Flow A
    connects: [comp-a, comp-b]
`)

	proj, err := project.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got := len(proj.System.Assets); got != 1 {
		t.Errorf("Assets: want 1, got %d", got)
	} else if proj.System.Assets[0].ID != "asset-a" {
		t.Errorf("Assets[0].ID: want asset-a, got %q", proj.System.Assets[0].ID)
	}

	if got := len(proj.System.Components); got != 1 {
		t.Errorf("Components: want 1, got %d", got)
	} else if proj.System.Components[0].ID != "comp-a" {
		t.Errorf("Components[0].ID: want comp-a, got %q", proj.System.Components[0].ID)
	}

	if got := len(proj.System.TrustZones); got != 1 {
		t.Errorf("TrustZones: want 1, got %d", got)
	} else if proj.System.TrustZones[0].ID != "zone-a" {
		t.Errorf("TrustZones[0].ID: want zone-a, got %q", proj.System.TrustZones[0].ID)
	}

	if got := len(proj.System.DataFlows); got != 1 {
		t.Errorf("DataFlows: want 1, got %d", got)
	} else if proj.System.DataFlows[0].ID != "flow-a" {
		t.Errorf("DataFlows[0].ID: want flow-a, got %q", proj.System.DataFlows[0].ID)
	}
}

func TestLoad_ImportsCompanyControls(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "system.yaml", `
imports:
  - path: "./company.yaml"
`)
	writeFile(t, dir, "company.yaml", `
controls:
  - id: ctrl-a
    title: Control A
  - id: ctrl-b
    title: Control B
`)

	proj, err := project.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got := len(proj.Company.Controls); got != 2 {
		t.Errorf("Controls: want 2, got %d", got)
	} else if proj.Company.Controls[0].ID != "ctrl-a" {
		t.Errorf("Controls[0].ID: want ctrl-a, got %q", proj.Company.Controls[0].ID)
	}
}

func TestLoad_ImportsThreats(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "system.yaml", `
imports:
  - path: "./threats.yaml"
`)
	writeFile(t, dir, "threats.yaml", `
threats:
  - id: threat-a
    title: Threat A
    severity: high
  - id: threat-b
    title: Threat B
    severity: critical
`)

	proj, err := project.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got := len(proj.ThreatModel.Threats); got != 2 {
		t.Errorf("Threats: want 2, got %d", got)
	} else if proj.ThreatModel.Threats[0].ID != "threat-a" {
		t.Errorf("Threats[0].ID: want threat-a, got %q", proj.ThreatModel.Threats[0].ID)
	}
}

func TestLoad_CycleDetection(t *testing.T) {
	dir := t.TempDir()
	// system.yaml imports threats.yaml; threats.yaml imports back to system.yaml
	writeFile(t, dir, "system.yaml", `
imports:
  - path: "./threats.yaml"
assets:
  - id: asset-a
    type: data
`)
	writeFile(t, dir, "threats.yaml", `
imports:
  - path: "./system.yaml"
threats:
  - id: threat-a
    title: Threat A
    severity: high
`)

	proj, err := project.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got := len(proj.System.Assets); got != 1 {
		t.Errorf("Assets: want 1 (no duplicates), got %d", got)
	}
	if got := len(proj.ThreatModel.Threats); got != 1 {
		t.Errorf("Threats: want 1 (no duplicates), got %d", got)
	}
}

func TestLoad_MergesAcrossFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "system.yaml", `
imports:
  - path: "./extra.yaml"
assets:
  - id: asset-a
    type: data
components:
  - id: comp-a
    type: server
`)
	writeFile(t, dir, "extra.yaml", `
assets:
  - id: asset-b
    type: firmware
components:
  - id: comp-b
    type: embedded-device
`)

	proj, err := project.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got := len(proj.System.Assets); got != 2 {
		t.Errorf("Assets: want 2, got %d", got)
	}
	if got := len(proj.System.Components); got != 2 {
		t.Errorf("Components: want 2, got %d", got)
	}
}

func TestLoad_FirstSystemMetaWins(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "system.yaml", `
imports:
  - path: "./extra.yaml"
system:
  id: primary
  title: Primary System
`)
	writeFile(t, dir, "extra.yaml", `
system:
  id: secondary
  title: Secondary System
`)

	proj, err := project.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if proj.System.SystemMeta == nil {
		t.Fatal("SystemMeta is nil")
	}
	if proj.System.SystemMeta.ID != "primary" {
		t.Errorf("SystemMeta.ID: want primary, got %q", proj.System.SystemMeta.ID)
	}
}

func TestLoad_MissingEntryPoint(t *testing.T) {
	dir := t.TempDir()

	_, err := project.Load(dir)
	if err == nil {
		t.Fatal("want error when system.yaml is absent, got nil")
	}
}

func TestLoad_CompanyVocabularyThreatsIgnored(t *testing.T) {
	dir := t.TempDir()
	// company.yaml uses "threats:" as a vocabulary mapping, not a threat list
	writeFile(t, dir, "system.yaml", `
imports:
  - path: "./company.yaml"
`)
	writeFile(t, dir, "company.yaml", `
controls:
  - id: ctrl-a
    title: Control A
threats:
  types:
    - spoofing
    - tampering
  severity:
    - low
    - high
`)

	proj, err := project.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got := len(proj.ThreatModel.Threats); got != 0 {
		t.Errorf("Threats: want 0 (vocabulary object must not be parsed as threat list), got %d", got)
	}
	if got := len(proj.Company.Controls); got != 1 {
		t.Errorf("Controls: want 1, got %d", got)
	}
}
