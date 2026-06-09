package quarto

import (
	"strings"
	"text/template"

	"sectrack/internal/model"
)

// ReportMeta carries the system-level metadata the threat model report needs.
type ReportMeta struct {
	Title       string
	Date        string
	Version     string
	Description string
}

type threatModelData struct {
	Meta        ReportMeta
	ThreatModel *model.ThreatModelFile
	Controls    map[string]string // id → title
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
title: "Threat Model — {{ .Meta.Title }}"
date: "{{ .Meta.Date }}"
version: "{{ .Meta.Version }}"
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

{{ .Meta.Description }}

## Data Flow Diagram

` + "```{mermaid}" + `
{{ .Diagram }}
` + "```" + `

## Threats

### Summary

| Severity | ID | Title | Target | Residual Risk |
|---|---|---|---|---|
{{ range .ThreatModel.Threats -}}
| {{ .Severity }} | {{ .ID }} | {{ .Title }} | {{ .Target }} | {{ .ResidualRisk }} |
{{ end }}

### Details
{{ range .ThreatModel.Threats }}
#### {{ .Title }}

| Field | Value |
|---|---|
| **ID** | `+"`"+`{{ .ID }}`+"`"+` |
| **Type** | {{ .Type }} |
| **Target** | `+"`"+`{{ .Target }}`+"`"+` |
{{ if .Asset -}}
| **Asset** | `+"`"+`{{ .Asset }}`+"`"+` |
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

func ThreatModel(meta ReportMeta, tm *model.ThreatModelFile, company *model.CompanyFile, diagram string, pdf bool) (string, error) {
	controls := make(map[string]string, len(company.Controls))
	for _, c := range company.Controls {
		controls[c.ID] = c.Title
	}

	var b strings.Builder
	err := threatModelTmpl.Execute(&b, threatModelData{
		Meta:        meta,
		ThreatModel: tm,
		Controls:    controls,
		Diagram:     strings.TrimRight(diagram, "\n"),
		PDF:         pdf,
	})
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
