package quarto

import (
	_ "embed"
	"strings"
	"text/template"

	"sectrack/internal/model"
)

//go:embed templates/threat-model.tmpl
var defaultTmplContent []byte

// funcMap is the set of functions available inside all threat model templates.
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

// DefaultTemplateContent returns the raw bytes of the built-in threat model
// template. Pass these to ParseTemplate, or write them to a file as a
// customisation starting point.
func DefaultTemplateContent() []byte {
	return defaultTmplContent
}

// ParseTemplate compiles a threat model template from raw content.
// The sectrack function set (controlTitle, join, upper) is registered
// automatically so user templates can call the same helpers.
func ParseTemplate(content []byte) (*template.Template, error) {
	return template.New("threat-model").Funcs(funcMap).Parse(string(content))
}

// DefaultTemplate returns the compiled built-in threat model template.
func DefaultTemplate() *template.Template {
	return template.Must(ParseTemplate(defaultTmplContent))
}

// threatModelData is the context passed to every threat model template.
// Field names are part of the template API — renaming them is a breaking change.
type threatModelData struct {
	Title       string
	Date        string
	Version     string
	Description string
	Threats     []model.Threat
	Controls           map[string]string   // id → title, for the controlTitle helper
	ControlList        []model.Control     // full control objects for rendering a controls section
	ControlComponents  map[string][]string // control id → component ids that implement it
	Diagram     string
	PDF         bool
}

// ThreatModel renders a threat model report using the provided template.
// Pass DefaultTemplate() or a template compiled with ParseTemplate().
func ThreatModel(proj *model.Project, tmpl *template.Template, diagram string, pdf bool) (string, error) {
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

	data := threatModelData{
		Threats:           proj.Threats,
		Controls:          controls,
		ControlList:       proj.Controls,
		ControlComponents: controlComponents,
		Diagram:     strings.TrimRight(diagram, "\n"),
		PDF:         pdf,
	}
	if proj.SystemMeta != nil {
		data.Title = proj.SystemMeta.Title
		data.Description = proj.SystemMeta.Description
	}
	data.Date = proj.Version.ReleaseDate
	data.Version = proj.Version.Semver

	var b strings.Builder
	if err := tmpl.Execute(&b, data); err != nil {
		return "", err
	}
	return b.String(), nil
}
