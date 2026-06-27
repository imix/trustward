package model

// Index answers the relational questions over a Project: by-id lookup, display
// title for a threat target, and the asset/objective/control inversions.
// Building it once concentrates the joins that consumers (the report renderer,
// the diagram renderer) would otherwise each re-derive inline.
//
// It is a pure query surface — it reports nothing about the data's validity
// (duplicate ids, dangling references). That stays in the validator, which
// needs to surface those as issues rather than silently dedup them here.
type Index struct {
	components map[string]Component
	titles     map[string]string // component OR data-flow id → display title

	componentsByAsset     map[string][]string // asset id → component ids hosting it
	assetsByObjective     map[string][]string // objective id → asset ids upholding it
	componentsByControl   map[string][]string // control id → component ids implementing it
	controlsByRequirement map[string][]string // "catalog::req" → control ids implementing it
	controlTitles         map[string]string   // control id → title
}

// NewIndex builds the lookups and inversions for a project.
func NewIndex(p *Project) *Index {
	i := &Index{
		components:            make(map[string]Component, len(p.Components)),
		titles:                make(map[string]string, len(p.Components)+len(p.DataFlows)),
		componentsByAsset:     map[string][]string{},
		assetsByObjective:     map[string][]string{},
		componentsByControl:   map[string][]string{},
		controlsByRequirement: map[string][]string{},
		controlTitles:         make(map[string]string, len(p.Controls)),
	}
	for _, c := range p.Components {
		i.components[c.ID] = c
		i.titles[c.ID] = titleOr(c.Title, c.ID)
		for _, a := range c.Assets {
			i.componentsByAsset[a] = append(i.componentsByAsset[a], c.ID)
		}
		for _, ctrl := range c.Controls {
			i.componentsByControl[ctrl] = append(i.componentsByControl[ctrl], c.ID)
		}
	}
	for _, f := range p.DataFlows {
		i.titles[f.ID] = titleOr(f.Title, f.ID)
	}
	for _, a := range p.Assets {
		for _, o := range a.Objectives {
			i.assetsByObjective[o] = append(i.assetsByObjective[o], a.ID)
		}
	}
	for _, c := range p.Controls {
		i.controlTitles[c.ID] = c.Title
		if c.Ref != "" {
			i.controlsByRequirement[c.Ref] = append(i.controlsByRequirement[c.Ref], c.ID)
		}
	}
	return i
}

func titleOr(title, id string) string {
	if title != "" {
		return title
	}
	return id
}

// Label returns a component's display title, falling back to its id. Used for
// diagram nodes, which are always components.
func (i *Index) Label(id string) string {
	if c, ok := i.components[id]; ok {
		return titleOr(c.Title, id)
	}
	return id
}

// TargetTitle returns the display title for a threat target — a component or a
// data flow — falling back to the id.
func (i *Index) TargetTitle(id string) string {
	if t, ok := i.titles[id]; ok {
		return t
	}
	return id
}

// The inversions. Each maps an id to the ids that reference it the other way.
func (i *Index) ComponentsByAsset() map[string][]string     { return i.componentsByAsset }
func (i *Index) AssetsByObjective() map[string][]string     { return i.assetsByObjective }
func (i *Index) ComponentsByControl() map[string][]string   { return i.componentsByControl }
func (i *Index) ControlsByRequirement() map[string][]string { return i.controlsByRequirement }
func (i *Index) ControlTitles() map[string]string           { return i.controlTitles }
