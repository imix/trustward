package mermaid

import (
	"fmt"
	"strings"

	"sectrack/internal/model"
)

// DataFlow renders a system file as a Mermaid flowchart showing components
// grouped by trust zone, with labelled edges for each data flow.
func DataFlow(sys *model.SystemFile) string {
	var b strings.Builder

	b.WriteString("flowchart TD\n")

	inZone := make(map[string]bool)

	for _, zone := range sys.TrustZones {
		zoneID := toMermaidID(zone.ID)
		b.WriteString(fmt.Sprintf("    subgraph %s[\"%s\"]\n", zoneID, zone.Title))
		for _, memberID := range zone.Members {
			b.WriteString(fmt.Sprintf("        %s[\"%s\"]\n", toMermaidID(memberID), memberID))
			inZone[memberID] = true
		}
		b.WriteString("    end\n")
	}

	for _, comp := range sys.Components {
		if !inZone[comp.ID] {
			b.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", toMermaidID(comp.ID), comp.ID))
		}
	}

	b.WriteString("\n")

	for _, flow := range sys.DataFlows {
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
