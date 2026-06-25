package mermaid

import (
	"fmt"
	"strings"

	"github.com/imix/trustward/internal/model"
)

type colorSpec struct{ fill, stroke, text string }

// zoneColors cycles through distinct backgrounds for trust zone subgraphs.
var zoneColors = []colorSpec{
	{"#b8d4f0", "#5b8db8", "#1a3a5c"},
	{"#c8e8c0", "#5b8b5a", "#1a3c1a"},
	{"#f5e6c8", "#b8935b", "#3c2a1a"},
	{"#e8c8f0", "#8b5bb8", "#2a1a3c"},
}

// typeColors maps Mermaid-safe component type names to fill/stroke/text colours.
var typeColors = map[string]colorSpec{
	"embedded_device": {"#dce8f5", "#5b8db8", "#1a3a5c"},
	"service":         {"#dcf5e4", "#5b8b6a", "#1a3c1a"},
	"database":        {"#f5f0dc", "#b8a05b", "#3c2a1a"},
	"gateway":         {"#f5dcf0", "#b85b8b", "#3c1a3a"},
}

func typeColor(t string) colorSpec {
	if c, ok := typeColors[t]; ok {
		return c
	}
	return colorSpec{"#f0f0f0", "#999999", "#333333"}
}

// DataFlow renders a project as a Mermaid flowchart showing components
// grouped by trust zone, with labelled edges for each data flow.
func DataFlow(proj *model.Project) string {
	var b strings.Builder

	compByID := make(map[string]model.Component, len(proj.Components))
	for _, c := range proj.Components {
		compByID[c.ID] = c
	}

	nodeLabel := func(id string) string {
		if c, ok := compByID[id]; ok && c.Title != "" {
			return c.Title
		}
		return id
	}

	b.WriteString("flowchart TD\n")

	inZone := make(map[string]bool)

	for i, zone := range proj.TrustZones {
		zoneID := toMermaidID(zone.ID)
		b.WriteString(fmt.Sprintf("    subgraph %s[\"%s\"]\n", zoneID, zone.Title))
		for _, memberID := range zone.Members {
			b.WriteString(fmt.Sprintf("        %s[\"%s\"]\n", toMermaidID(memberID), nodeLabel(memberID)))
			inZone[memberID] = true
		}
		b.WriteString("    end\n")
		zc := zoneColors[i%len(zoneColors)]
		b.WriteString(fmt.Sprintf("    style %s fill:%s,stroke:%s,color:%s\n", zoneID, zc.fill, zc.stroke, zc.text))
	}

	for _, comp := range proj.Components {
		if !inZone[comp.ID] {
			b.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", toMermaidID(comp.ID), nodeLabel(comp.ID)))
		}
	}

	// classDef per unique component type
	seen := make(map[string]bool)
	var types []string
	for _, comp := range proj.Components {
		t := toMermaidID(comp.Type)
		if t != "" && !seen[t] {
			seen[t] = true
			types = append(types, t)
		}
	}
	if len(types) > 0 {
		b.WriteString("\n")
		for _, t := range types {
			c := typeColor(t)
			b.WriteString(fmt.Sprintf("    classDef %s fill:%s,stroke:%s,color:%s\n", t, c.fill, c.stroke, c.text))
		}
		for _, comp := range proj.Components {
			t := toMermaidID(comp.Type)
			if t != "" {
				b.WriteString(fmt.Sprintf("    class %s %s\n", toMermaidID(comp.ID), t))
			}
		}
	}

	b.WriteString("\n")

	for _, flow := range proj.DataFlows {
		if len(flow.Connects) != 2 {
			continue
		}
		from := toMermaidID(flow.Connects[0])
		to := toMermaidID(flow.Connects[1])
		if len(flow.Assets) > 0 {
			b.WriteString(fmt.Sprintf("    %s -->|\"%s\"| %s\n", from, strings.Join(flow.Assets, ", "), to))
		} else {
			b.WriteString(fmt.Sprintf("    %s --> %s\n", from, to))
		}
	}

	return b.String()
}

// toMermaidID converts a kebab-case ID to a Mermaid-safe identifier.
// Mermaid interprets hyphens as minus signs in bare IDs.
func toMermaidID(id string) string {
	return strings.ReplaceAll(id, "-", "_")
}
