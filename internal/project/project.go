package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/imix/trustward/internal/model"
	"gopkg.in/yaml.v3"
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

	// Parse the file once into its top-level nodes; decode each key on demand.
	var nodes map[string]yaml.Node
	if err := yaml.Unmarshal(data, &nodes); err != nil {
		return fmt.Errorf("%s: parsing: %w", name, err)
	}

	// List keys merge by appending every file's contribution.
	for _, err := range []error{
		mergeList(&p.Assets, nodes, "assets", name),
		mergeList(&p.Objectives, nodes, "objectives", name),
		mergeList(&p.Components, nodes, "components", name),
		mergeList(&p.TrustZones, nodes, "trust-zones", name),
		mergeList(&p.DataFlows, nodes, "data-flows", name),
		mergeList(&p.Controls, nodes, "controls", name),
	} {
		if err != nil {
			return err
		}
	}

	// Singletons: first occurrence in the graph wins.
	if n, ok := nodes["version"]; ok && p.Version.Semver == "" {
		if err := n.Decode(&p.Version); err != nil {
			return fmt.Errorf("%s: version: %w", name, err)
		}
	}
	if n, ok := nodes["system"]; ok && p.SystemMeta == nil {
		if err := n.Decode(&p.SystemMeta); err != nil {
			return fmt.Errorf("%s: system: %w", name, err)
		}
	}
	if n, ok := nodes["risk-policy"]; ok && !p.RiskPolicy.Set {
		if err := n.Decode(&p.RiskPolicy); err != nil {
			return fmt.Errorf("%s: risk-policy: %w", name, err)
		}
		p.RiskPolicy.Set = true
	}

	// "threats" is a list in a threat model but a mapping in company vocabulary.
	// Only treat it as threats when the value is a sequence.
	if n, ok := nodes["threats"]; ok && n.Kind == yaml.SequenceNode {
		var threats []model.Threat
		if err := n.Decode(&threats); err != nil {
			return fmt.Errorf("%s: threats: %w", name, err)
		}
		p.Threats = append(p.Threats, threats...)
	}

	// Catalogs: each file contributes at most one; skip if it has no id.
	if n, ok := nodes["catalog"]; ok {
		var cat model.ControlCatalog
		if err := n.Decode(&cat); err != nil {
			return fmt.Errorf("%s: catalog: %w", name, err)
		}
		if cat.ID != "" {
			p.Catalogs = append(p.Catalogs, cat)
		}
	}
	if n, ok := nodes["threat-catalog"]; ok {
		var tc model.ThreatCatalog
		if err := n.Decode(&tc); err != nil {
			return fmt.Errorf("%s: threat-catalog: %w", name, err)
		}
		if tc.ID != "" {
			p.ThreatCatalogs = append(p.ThreatCatalogs, tc)
		}
	}

	// Follow imports relative to this file's directory.
	if n, ok := nodes["imports"]; ok {
		var imports []rawImport
		if err := n.Decode(&imports); err != nil {
			return fmt.Errorf("%s: imports: %w", name, err)
		}
		dir := filepath.Dir(abs)
		for _, imp := range imports {
			if err := loadGraph(p, filepath.Join(dir, imp.Path), visited); err != nil {
				return err
			}
		}
	}

	return nil
}

// mergeList decodes a list-valued top-level key, if present, and appends it to
// dst — the single place the list-merge rule lives. An absent key is a no-op.
func mergeList[T any](dst *[]T, nodes map[string]yaml.Node, key, name string) error {
	n, ok := nodes[key]
	if !ok {
		return nil
	}
	var items []T
	if err := n.Decode(&items); err != nil {
		return fmt.Errorf("%s: %s: %w", name, key, err)
	}
	*dst = append(*dst, items...)
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
