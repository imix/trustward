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

	if got := len(proj.Assets); got != 1 {
		t.Errorf("Assets: want 1, got %d", got)
	} else if proj.Assets[0].ID != "asset-a" {
		t.Errorf("Assets[0].ID: want asset-a, got %q", proj.Assets[0].ID)
	}

	if got := len(proj.Components); got != 1 {
		t.Errorf("Components: want 1, got %d", got)
	} else if proj.Components[0].ID != "comp-a" {
		t.Errorf("Components[0].ID: want comp-a, got %q", proj.Components[0].ID)
	}

	if got := len(proj.TrustZones); got != 1 {
		t.Errorf("TrustZones: want 1, got %d", got)
	} else if proj.TrustZones[0].ID != "zone-a" {
		t.Errorf("TrustZones[0].ID: want zone-a, got %q", proj.TrustZones[0].ID)
	}

	if got := len(proj.DataFlows); got != 1 {
		t.Errorf("DataFlows: want 1, got %d", got)
	} else if proj.DataFlows[0].ID != "flow-a" {
		t.Errorf("DataFlows[0].ID: want flow-a, got %q", proj.DataFlows[0].ID)
	}
}

func TestLoad_ObjectivesAndAssetUpholdAndThreatViolates(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "system.yaml", `
objectives:
  - id: obj-confidentiality
    title: Sensor data confidentiality
    type: confidentiality
    description: Readings must not leak
assets:
  - id: asset-readings
    type: telemetry
    objectives: [obj-confidentiality]
threats:
  - id: threat-leak
    title: Eavesdrop readings
    violates: [obj-confidentiality]
`)

	proj, err := project.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got := len(proj.Objectives); got != 1 {
		t.Fatalf("Objectives: want 1, got %d", got)
	}
	if o := proj.Objectives[0]; o.ID != "obj-confidentiality" || o.Type != "confidentiality" {
		t.Errorf("Objectives[0]: want id/type confidentiality, got %q/%q", o.ID, o.Type)
	}
	if got := proj.Assets[0].Objectives; len(got) != 1 || got[0] != "obj-confidentiality" {
		t.Errorf("Assets[0].Objectives: want [obj-confidentiality], got %v", got)
	}
	if got := proj.Threats[0].Violates; len(got) != 1 || got[0] != "obj-confidentiality" {
		t.Errorf("Threats[0].Violates: want [obj-confidentiality], got %v", got)
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

	if got := len(proj.Controls); got != 2 {
		t.Errorf("Controls: want 2, got %d", got)
	} else if proj.Controls[0].ID != "ctrl-a" {
		t.Errorf("Controls[0].ID: want ctrl-a, got %q", proj.Controls[0].ID)
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

	if got := len(proj.Threats); got != 2 {
		t.Errorf("Threats: want 2, got %d", got)
	} else if proj.Threats[0].ID != "threat-a" {
		t.Errorf("Threats[0].ID: want threat-a, got %q", proj.Threats[0].ID)
	}
}

func TestLoad_ThreatRiskFieldsAndPolicy(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "system.yaml", `
imports:
  - path: "./extra.yaml"
risk-policy:
  method: qualitative
  accept: [low]
threats:
  - id: threat-a
    likelihood: high
    impact: high
    treatment: mitigate
    owner: alice
    decided: "2026-06-25"
`)
	// later file's risk-policy must be ignored (first wins)
	writeFile(t, dir, "extra.yaml", `
risk-policy:
  method: etsi-tvra
  accept: [low, medium]
`)

	proj, err := project.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !proj.RiskPolicy.Set || proj.RiskPolicy.Method != "qualitative" {
		t.Errorf("RiskPolicy: want qualitative/set, got %+v", proj.RiskPolicy)
	}
	if len(proj.RiskPolicy.Accept) != 1 || proj.RiskPolicy.Accept[0] != "low" {
		t.Errorf("RiskPolicy.Accept: want [low] (first wins), got %v", proj.RiskPolicy.Accept)
	}
	if len(proj.Threats) != 1 {
		t.Fatalf("Threats: want 1, got %d", len(proj.Threats))
	}
	th := proj.Threats[0]
	if th.Likelihood != "high" || th.Impact != "high" || th.Treatment != "mitigate" ||
		th.Owner != "alice" || th.Decided != "2026-06-25" {
		t.Errorf("threat risk fields not parsed: %+v", th)
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

	if got := len(proj.Assets); got != 1 {
		t.Errorf("Assets: want 1 (no duplicates), got %d", got)
	}
	if got := len(proj.Threats); got != 1 {
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

	if got := len(proj.Assets); got != 2 {
		t.Errorf("Assets: want 2, got %d", got)
	}
	if got := len(proj.Components); got != 2 {
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

	if proj.SystemMeta == nil {
		t.Fatal("SystemMeta is nil")
	}
	if proj.SystemMeta.ID != "primary" {
		t.Errorf("SystemMeta.ID: want primary, got %q", proj.SystemMeta.ID)
	}
}

func TestLoad_MissingEntryPoint(t *testing.T) {
	dir := t.TempDir()

	_, err := project.Load(dir)
	if err == nil {
		t.Fatal("want error when system.yaml is absent, got nil")
	}
}

func TestLoad_MalformedKeyReturnsError(t *testing.T) {
	cases := []struct {
		name string
		yaml string
	}{
		{"assets wrong type", "assets: \"not a list\""},
		{"components wrong type", "components: 42"},
		{"controls wrong type", "controls: true"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFile(t, dir, "system.yaml", tc.yaml)
			_, err := project.Load(dir)
			if err == nil {
				t.Fatalf("want error for malformed %q, got nil", tc.name)
			}
		})
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

	if got := len(proj.Threats); got != 0 {
		t.Errorf("Threats: want 0 (vocabulary object must not be parsed as threat list), got %d", got)
	}
	if got := len(proj.Controls); got != 1 {
		t.Errorf("Controls: want 1, got %d", got)
	}
}
