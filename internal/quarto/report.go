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
	controls := make(map[string]string, len(proj.Controls))
	for _, c := range proj.Controls {
		controls[c.ID] = c.Title
	}

	controlComponents := make(map[string][]string)
	for _, comp := range proj.Components {
		for _, cid := range comp.Controls {
			controlComponents[cid] = append(controlComponents[cid], comp.ID)
		}
	}

	reqControls := make(map[string][]string)
	for _, c := range proj.Controls {
		if c.Ref != "" {
			reqControls[c.Ref] = append(reqControls[c.Ref], c.ID)
		}
	}

	assetComponents := make(map[string][]string)
	for _, comp := range proj.Components {
		for _, assetID := range comp.Assets {
			assetComponents[assetID] = append(assetComponents[assetID], comp.ID)
		}
	}

	objectiveAssets := make(map[string][]string)
	for _, a := range proj.Assets {
		for _, oid := range a.Objectives {
			objectiveAssets[oid] = append(objectiveAssets[oid], a.ID)
		}
	}

	// A threat target is a component or a data flow — title lookup must cover both.
	targetTitles := make(map[string]string, len(proj.Components)+len(proj.DataFlows))
	for _, c := range proj.Components {
		t := c.Title
		if t == "" {
			t = c.ID
		}
		targetTitles[c.ID] = t
	}
	for _, f := range proj.DataFlows {
		t := f.Title
		if t == "" {
			t = f.ID
		}
		targetTitles[f.ID] = t
	}

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
		title := targetTitles[targetID]
		if title == "" {
			title = targetID
		}
		groups = append(groups, ThreatGroup{
			TargetID:    targetID,
			TargetTitle: title,
			Threats:     threatMap[targetID],
		})
	}

	data := reportData{
		AssetList:           proj.Assets,
		AssetComponents:     assetComponents,
		ObjectiveList:       proj.Objectives,
		ObjectiveAssets:     objectiveAssets,
		ThreatGroups:        groups,
		ThreatList:          proj.Threats,
		RiskEval:            risk.Evaluate(proj),
		RiskMethod:          proj.RiskPolicy.Method,
		RiskAccept:          proj.RiskPolicy.Accept,
		RiskReview:          proj.RiskPolicy.Review,
		RiskPolicySet:       proj.RiskPolicy.Set,
		Controls:            controls,
		ControlList:         proj.Controls,
		ControlComponents:   controlComponents,
		ComponentList:       proj.Components,
		CatalogList:         proj.Catalogs,
		RequirementControls: reqControls,
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
