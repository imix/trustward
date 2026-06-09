package project

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"sectrack/internal/model"
)

// Load reads system.yaml from dir and follows its import graph depth-first,
// merging each file's content into a single Project.
func Load(dir string) (*model.Project, error) {
	entry := filepath.Join(dir, "system.yaml")
	p := &model.Project{}
	visited := make(map[string]bool)
	if err := loadGraph(p, entry, visited); err != nil {
		return nil, err
	}
	resolveThreatRefs(p)
	return p, nil
}

// rawImport is used only during loading to follow import declarations.
type rawImport struct {
	Path string `yaml:"path"`
}

func loadGraph(p *model.Project, path string, visited map[string]bool) error {
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

	// Phase 2: unmarshal and merge by content type.
	if hasAny(keys, "version") && p.Version.Semver == "" {
		var v struct {
			Version model.Version `yaml:"version"`
		}
		yaml.Unmarshal(data, &v) //nolint:errcheck
		p.Version = v.Version
	}

	if hasAny(keys, "system") && p.SystemMeta == nil {
		var v struct {
			SystemMeta *model.SystemMeta `yaml:"system"`
		}
		yaml.Unmarshal(data, &v) //nolint:errcheck
		p.SystemMeta = v.SystemMeta
	}

	if hasAny(keys, "assets") {
		var v struct {
			Assets []model.Asset `yaml:"assets"`
		}
		yaml.Unmarshal(data, &v) //nolint:errcheck
		p.Assets = append(p.Assets, v.Assets...)
	}

	if hasAny(keys, "components") {
		var v struct {
			Components []model.Component `yaml:"components"`
		}
		yaml.Unmarshal(data, &v) //nolint:errcheck
		p.Components = append(p.Components, v.Components...)
	}

	if hasAny(keys, "trust-zones") {
		var v struct {
			TrustZones []model.TrustZone `yaml:"trust-zones"`
		}
		yaml.Unmarshal(data, &v) //nolint:errcheck
		p.TrustZones = append(p.TrustZones, v.TrustZones...)
	}

	if hasAny(keys, "data-flows") {
		var v struct {
			DataFlows []model.DataFlow `yaml:"data-flows"`
		}
		yaml.Unmarshal(data, &v) //nolint:errcheck
		p.DataFlows = append(p.DataFlows, v.DataFlows...)
	}

	// "threats" is a list in a threat model but a mapping in company vocabulary.
	// Only treat it as threats when the value is a sequence.
	if v, ok := keys["threats"]; ok {
		if _, isList := v.([]interface{}); isList {
			var f struct {
				Threats []model.Threat `yaml:"threats"`
			}
			yaml.Unmarshal(data, &f) //nolint:errcheck
			p.Threats = append(p.Threats, f.Threats...)
		}
	}

	if hasAny(keys, "controls") {
		var v struct {
			Controls []model.Control `yaml:"controls"`
		}
		yaml.Unmarshal(data, &v) //nolint:errcheck
		p.Controls = append(p.Controls, v.Controls...)
	}

	if hasAny(keys, "catalog") {
		var v struct {
			Catalog model.Catalog `yaml:"catalog"`
		}
		yaml.Unmarshal(data, &v) //nolint:errcheck
		if v.Catalog.ID != "" {
			p.Catalogs = append(p.Catalogs, v.Catalog)
		}
	}

	if hasAny(keys, "threat-catalog") {
		var v struct {
			ThreatCatalog model.ThreatCatalog `yaml:"threat-catalog"`
		}
		yaml.Unmarshal(data, &v) //nolint:errcheck
		if v.ThreatCatalog.ID != "" {
			p.ThreatCatalogs = append(p.ThreatCatalogs, v.ThreatCatalog)
		}
	}

	// Phase 3: follow imports relative to this file's directory.
	var imported struct {
		Imports []rawImport `yaml:"imports"`
	}
	yaml.Unmarshal(data, &imported) //nolint:errcheck
	dir := filepath.Dir(abs)
	for _, imp := range imported.Imports {
		if err := loadGraph(p, filepath.Join(dir, imp.Path), visited); err != nil {
			return err
		}
	}

	return nil
}

// resolveThreatRefs fills in fields on threats that reference a catalog pattern.
// Instance fields take precedence; catalog values are used only when the field is empty.
func resolveThreatRefs(p *model.Project) {
	patterns := make(map[string]model.ThreatPattern)
	for _, cat := range p.ThreatCatalogs {
		for _, pat := range cat.Patterns {
			patterns[cat.ID+"::"+pat.ID] = pat
		}
	}
	for i, t := range p.Threats {
		if t.Ref == "" {
			continue
		}
		pat, ok := patterns[t.Ref]
		if !ok {
			continue
		}
		if t.Title == "" {
			p.Threats[i].Title = pat.Title
		}
		if t.Type == "" {
			p.Threats[i].Type = pat.Type
		}
		if t.Severity == "" {
			p.Threats[i].Severity = pat.Severity
		}
		if t.Notes == "" {
			p.Threats[i].Notes = pat.Notes
		}
	}
}

func hasAny(keys map[string]interface{}, names ...string) bool {
	for _, name := range names {
		if _, ok := keys[name]; ok {
			return true
		}
	}
	return false
}
