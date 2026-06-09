package project

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"sectrack/internal/model"
)

// Project holds the accumulated security model for a directory.
// Content is merged across all files reachable via the import graph rooted at
// system.yaml. Filenames are arbitrary; content is routed by top-level key.
type Project struct {
	System      model.SystemFile
	ThreatModel model.ThreatModelFile
	Company     model.CompanyFile
}

// Load reads system.yaml from dir and follows its import graph depth-first,
// merging each file's content into the returned Project.
func Load(dir string) (*Project, error) {
	entry := filepath.Join(dir, "system.yaml")
	p := &Project{}
	visited := make(map[string]bool)
	if err := loadGraph(p, entry, visited); err != nil {
		return nil, err
	}
	return p, nil
}

func loadGraph(p *Project, path string, visited map[string]bool) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if visited[abs] {
		return nil
	}
	visited[abs] = true

	data, err := os.ReadFile(abs)
	if err != nil {
		return fmt.Errorf("%s: %w", filepath.Base(abs), err)
	}

	// Phase 1: detect content type from top-level keys.
	var keys map[string]interface{}
	if err := yaml.Unmarshal(data, &keys); err != nil {
		return fmt.Errorf("%s: parsing: %w", filepath.Base(abs), err)
	}

	// Phase 2: unmarshal into domain structs and merge.
	if hasAny(keys, "system", "assets", "components", "trust-zones", "data-flows") {
		var f model.SystemFile
		if err := yaml.Unmarshal(data, &f); err != nil {
			return fmt.Errorf("%s: unmarshaling system content: %w", filepath.Base(abs), err)
		}
		if p.System.SystemMeta == nil {
			p.System.SystemMeta = f.SystemMeta
		}
		if p.System.Version.Semver == "" {
			p.System.Version = f.Version
		}
		p.System.Assets = append(p.System.Assets, f.Assets...)
		p.System.Components = append(p.System.Components, f.Components...)
		p.System.TrustZones = append(p.System.TrustZones, f.TrustZones...)
		p.System.DataFlows = append(p.System.DataFlows, f.DataFlows...)
	}

	// "threats" can be either a list (threat model) or a mapping (company
	// vocabulary). Only treat it as threats if it is a sequence.
	if v, ok := keys["threats"]; ok {
		if _, isList := v.([]interface{}); isList {
			var f model.ThreatModelFile
			if err := yaml.Unmarshal(data, &f); err != nil {
				return fmt.Errorf("%s: unmarshaling threat content: %w", filepath.Base(abs), err)
			}
			p.ThreatModel.Threats = append(p.ThreatModel.Threats, f.Threats...)
			if p.ThreatModel.Version.Semver == "" {
				p.ThreatModel.Version = f.Version
			}
		}
	}

	if hasAny(keys, "controls") {
		var f model.CompanyFile
		if err := yaml.Unmarshal(data, &f); err != nil {
			return fmt.Errorf("%s: unmarshaling company content: %w", filepath.Base(abs), err)
		}
		p.Company.Controls = append(p.Company.Controls, f.Controls...)
	}

	// Phase 3: follow imports relative to this file's directory.
	var imported struct {
		Imports []model.Import `yaml:"imports"`
	}
	yaml.Unmarshal(data, &imported) //nolint:errcheck — already parsed above
	dir := filepath.Dir(abs)
	for _, imp := range imported.Imports {
		if err := loadGraph(p, filepath.Join(dir, imp.Path), visited); err != nil {
			return err
		}
	}

	return nil
}

func hasAny(keys map[string]interface{}, names ...string) bool {
	for _, name := range names {
		if _, ok := keys[name]; ok {
			return true
		}
	}
	return false
}
