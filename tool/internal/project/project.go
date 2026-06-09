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

	name := filepath.Base(abs)

	// Phase 1: detect content type from top-level keys.
	var keys map[string]interface{}
	if err := yaml.Unmarshal(data, &keys); err != nil {
		return fmt.Errorf("%s: parsing: %w", name, err)
	}

	// Phase 2: unmarshal and merge by content type.
	if hasAny(keys, "version") && p.Version.Semver == "" {
		var v struct {
			Version model.Version `yaml:"version"`
		}
		if err := yaml.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("%s: version: %w", name, err)
		}
		p.Version = v.Version
	}

	if hasAny(keys, "system") && p.SystemMeta == nil {
		var v struct {
			SystemMeta *model.SystemMeta `yaml:"system"`
		}
		if err := yaml.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("%s: system: %w", name, err)
		}
		p.SystemMeta = v.SystemMeta
	}

	if hasAny(keys, "assets") {
		var v struct {
			Assets []model.Asset `yaml:"assets"`
		}
		if err := yaml.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("%s: assets: %w", name, err)
		}
		p.Assets = append(p.Assets, v.Assets...)
	}

	if hasAny(keys, "components") {
		var v struct {
			Components []model.Component `yaml:"components"`
		}
		if err := yaml.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("%s: components: %w", name, err)
		}
		p.Components = append(p.Components, v.Components...)
	}

	if hasAny(keys, "trust-zones") {
		var v struct {
			TrustZones []model.TrustZone `yaml:"trust-zones"`
		}
		if err := yaml.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("%s: trust-zones: %w", name, err)
		}
		p.TrustZones = append(p.TrustZones, v.TrustZones...)
	}

	if hasAny(keys, "data-flows") {
		var v struct {
			DataFlows []model.DataFlow `yaml:"data-flows"`
		}
		if err := yaml.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("%s: data-flows: %w", name, err)
		}
		p.DataFlows = append(p.DataFlows, v.DataFlows...)
	}

	// "threats" is a list in a threat model but a mapping in company vocabulary.
	// Only treat it as threats when the value is a sequence.
	if v, ok := keys["threats"]; ok {
		if _, isList := v.([]interface{}); isList {
			var f struct {
				Threats []model.Threat `yaml:"threats"`
			}
			if err := yaml.Unmarshal(data, &f); err != nil {
				return fmt.Errorf("%s: threats: %w", name, err)
			}
			p.Threats = append(p.Threats, f.Threats...)
		}
	}

	if hasAny(keys, "controls") {
		var v struct {
			Controls []model.Control `yaml:"controls"`
		}
		if err := yaml.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("%s: controls: %w", name, err)
		}
		p.Controls = append(p.Controls, v.Controls...)
	}

	if hasAny(keys, "catalog") {
		var v struct {
			Catalog model.Catalog `yaml:"catalog"`
		}
		if err := yaml.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("%s: catalog: %w", name, err)
		}
		if v.Catalog.ID != "" {
			p.Catalogs = append(p.Catalogs, v.Catalog)
		}
	}

	if hasAny(keys, "threat-catalog") {
		var v struct {
			ThreatCatalog model.ThreatCatalog `yaml:"threat-catalog"`
		}
		if err := yaml.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("%s: threat-catalog: %w", name, err)
		}
		if v.ThreatCatalog.ID != "" {
			p.ThreatCatalogs = append(p.ThreatCatalogs, v.ThreatCatalog)
		}
	}

	// Phase 3: follow imports relative to this file's directory.
	var imported struct {
		Imports []rawImport `yaml:"imports"`
	}
	if err := yaml.Unmarshal(data, &imported); err != nil {
		return fmt.Errorf("%s: imports: %w", name, err)
	}
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
