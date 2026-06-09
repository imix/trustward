package quarto

import (
	"strings"
	"text/template"

	"sectrack/internal/model"
)

type threatModelData struct {
	Title       string
	Date        string
	Version     string
	Description string
	Threats     []model.Threat
	Controls    map[string]string
	Diagram     string
	PDF         bool
}

var threatModelTmpl = template.Must(template.New("threat-model").Funcs(template.FuncMap{
	"join": strings.Join,
	"controlTitle": func(controls map[string]string, id string) string {
		if title, ok := controls[id]; ok {
			return title + " (`" + id + "`)"
		}
		return "`" + id + "`"
	},
	"upper": strings.ToUpper,
}).Parse(`---
title: "Threat Model — {{ .Title }}"
date: "{{ .Date }}"
version: "{{ .Version }}"
format:
  html:
    toc: true
    theme: cosmo
{{ if .PDF -}}
  pdf:
    toc: true
{{ end -}}
---

## System Overview

{{ .Description }}

## Data Flow Diagram

` + "```{mermaid}" + `
{{ .Diagram }}
` + "```" + `

## Threats

### Summary

| Severity | ID | Title | Target | Residual Risk |
|---|---|---|---|---|
{{ range .Threats -}}
| {{ .Severity }} | {{ .ID }} | {{ .Title }} | {{ .Target }} | {{ .ResidualRisk }} |
{{ end }}

### Details
{{ range .Threats }}
#### {{ .Title }}

| Field | Value |
|---|---|
| **ID** | ` + "`" + `{{ .ID }}` + "`" + ` |
| **Type** | {{ .Type }} |
| **Target** | ` + "`" + `{{ .Target }}` + "`" + ` |
{{ if .Asset -}}
| **Asset** | ` + "`" + `{{ .Asset }}` + "`" + ` |
{{ end -}}
| **Severity** | {{ .Severity }} |
| **Residual Risk** | {{ .ResidualRisk }} |
{{ if .Mitigations -}}
| **Mitigations** | {{ range $i, $m := .Mitigations }}{{ if $i }}, {{ end }}{{ controlTitle $.Controls $m }}{{ end }} |
{{ else -}}
| **Mitigations** | none |
{{ end }}
{{ .Notes }}
{{ end -}}
`))

func ThreatModel(proj *model.Project, diagram string, pdf bool) (string, error) {
	controls := make(map[string]string, len(proj.Controls))
	for _, c := range proj.Controls {
		controls[c.ID] = c.Title
	}

	data := threatModelData{
		Threats:  proj.Threats,
		Controls: controls,
		Diagram:  strings.TrimRight(diagram, "\n"),
		PDF:      pdf,
	}
	if proj.SystemMeta != nil {
		data.Title = proj.SystemMeta.Title
		data.Description = proj.SystemMeta.Description
	}
	data.Date = proj.Version.ReleaseDate
	data.Version = proj.Version.Semver

	var b strings.Builder
	if err := threatModelTmpl.Execute(&b, data); err != nil {
		return "", err
	}
	return b.String(), nil
}
