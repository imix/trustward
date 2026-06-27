package quarto

import (
	_ "embed"
	"strings"
	"text/template"

	"github.com/imix/trustward/internal/model"
	"github.com/imix/trustward/internal/risk"
)

//go:embed templates/report.tmpl
var defaultTmplContent []byte

// funcMap is the set of functions available inside all report templates.
var funcMap = template.FuncMap{
	"join": strings.Join,
	"trim": strings.TrimSpace,
	"controlTitle": func(controls map[string]string, id string) string {
		if title, ok := controls[id]; ok {
			return title + " (`" + id + "`)"
		}
		return "`" + id + "`"
	},
	"upper": strings.ToUpper,
}

// DefaultTemplateContent returns the raw bytes of the built-in report
// template. Pass these to ParseTemplate, or write them to a file as a
// customisation starting point.
func DefaultTemplateContent() []byte {
	return defaultTmplContent
}

// ParseTemplate compiles a report template from raw content.
// The trustward function set (controlTitle, join, upper) is registered
// automatically so user templates can call the same helpers.
func ParseTemplate(content []byte) (*template.Template, error) {
	return template.New("report").Funcs(funcMap).Parse(string(content))
}

// DefaultTemplate returns the compiled built-in report template.
func DefaultTemplate() *template.Template {
	return template.Must(ParseTemplate(defaultTmplContent))
}

// ThreatGroup groups threats by target for per-component report sections.
type ThreatGroup struct {
	TargetID    string
	TargetTitle string
	Threats     []model.Threat
}

// reportData is the context passed to every report template.
// Field names are part of the template API — renaming them is a breaking change.
type reportData struct {
	Title               string
	Date                string
	Version             string
	Description         string
	Logo                string
	References          []model.Reference // external versioned docs (variant register, requirements, standards, SBOM)
	AssetList           []model.Asset
	AssetComponents     map[string][]string  // asset id → component ids that hold it
	ObjectiveList       []model.Objective    // cybersecurity objectives (§6.5.2)
	ObjectiveAssets     map[string][]string  // objective id → asset ids that uphold it
	ThreatGroups        []ThreatGroup        // threats grouped by target, in encounter order
	ThreatList          []model.Threat       // flat list, for the risk register
	RiskEval            map[string]risk.Eval // threat id → computed score + evaluation vs acceptance criteria
	RiskMethod          string               // scoring method (risk-policy)
	RiskAccept          []string             // accepted risk levels (risk-policy)
	RiskReview          string               // monitoring and review cadence (§6.7)
	RiskPolicySet       bool                 // a risk-policy is declared → show register
	Controls            map[string]string    // id → title, for the controlTitle helper
	ControlList         []model.Control
	ControlComponents   map[string][]string // control id → component ids that implement it
	ComponentList       []model.Component
	CatalogList         []model.ControlCatalog
	RequirementControls map[string][]string // "catalog-id::req-id" → control IDs
	Diagram             string
	PDF                 bool
}

// Report renders the risk-management report using the provided template.
// Pass DefaultTemplate() or a template compiled with ParseTemplate().
func Report(proj *model.Project, tmpl *template.Template, diagram string, pdf bool) (string, error) {
	idx := model.NewIndex(proj)

	// Threats group by target in encounter order — a presentation concern, so
	// it stays here; the Index supplies the target's display title.
	var targetOrder []string
	targetSeen := make(map[string]bool)
	threatMap := make(map[string][]model.Threat)
	for _, t := range proj.Threats {
		if !targetSeen[t.Target] {
			targetSeen[t.Target] = true
			targetOrder = append(targetOrder, t.Target)
		}
		threatMap[t.Target] = append(threatMap[t.Target], t)
	}
	groups := make([]ThreatGroup, 0, len(targetOrder))
	for _, targetID := range targetOrder {
		groups = append(groups, ThreatGroup{
			TargetID:    targetID,
			TargetTitle: idx.TargetTitle(targetID),
			Threats:     threatMap[targetID],
		})
	}

	data := reportData{
		References:          proj.References,
		AssetList:           proj.Assets,
		AssetComponents:     idx.ComponentsByAsset(),
		ObjectiveList:       proj.Objectives,
		ObjectiveAssets:     idx.AssetsByObjective(),
		ThreatGroups:        groups,
		ThreatList:          proj.Threats,
		RiskEval:            risk.Evaluate(proj),
		RiskMethod:          proj.RiskPolicy.Method,
		RiskAccept:          proj.RiskPolicy.Accept,
		RiskReview:          proj.RiskPolicy.Review,
		RiskPolicySet:       proj.RiskPolicy.Set,
		Controls:            idx.ControlTitles(),
		ControlList:         proj.Controls,
		ControlComponents:   idx.ComponentsByControl(),
		ComponentList:       proj.Components,
		CatalogList:         proj.Catalogs,
		RequirementControls: idx.ControlsByRequirement(),
		Diagram:             strings.TrimRight(diagram, "\n"),
		PDF:                 pdf,
	}
	if proj.SystemMeta != nil {
		data.Title = proj.SystemMeta.Title
		data.Description = proj.SystemMeta.Description
		data.Logo = proj.SystemMeta.Logo
	}
	data.Date = proj.Version.ReleaseDate
	data.Version = proj.Version.Semver

	var b strings.Builder
	if err := tmpl.Execute(&b, data); err != nil {
		return "", err
	}
	return b.String(), nil
}
