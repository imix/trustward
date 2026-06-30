// Package dot renders a project's data flow diagram as Graphviz DOT. It is the
// PDF-friendly twin of internal/mermaid: Graphviz produces the diagram with the
// `dot` binary (no headless browser), so the rendered report — HTML and PDF
// alike — needs no Chromium. The standalone `diagram dataflow` command keeps
// using Mermaid, which renders client-side in any Markdown viewer.
package dot

import (
	"fmt"
	"strings"

	"github.com/imix/trustward/internal/model"
)

type colorSpec struct{ fill, stroke, text string }

// zoneColors cycles through distinct backgrounds for trust-zone clusters —
// the same palette as the Mermaid renderer, for visual parity.
var zoneColors = []colorSpec{
	{"#b8d4f0", "#5b8db8", "#1a3a5c"},
	{"#c8e8c0", "#5b8b5a", "#1a3c1a"},
	{"#f5e6c8", "#b8935b", "#3c2a1a"},
	{"#e8c8f0", "#8b5bb8", "#2a1a3c"},
}

// typeColors maps a component type to its node fill/stroke/text colours.
var typeColors = map[string]colorSpec{
	"embedded-device": {"#dce8f5", "#5b8db8", "#1a3a5c"},
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

// DataFlow renders a project as Graphviz DOT: components grouped into
// trust-zone clusters, with one labelled edge per data flow.
func DataFlow(proj *model.Project) string {
	var b strings.Builder
	idx := model.NewIndex(proj)

	compType := make(map[string]string, len(proj.Components))
	for _, c := range proj.Components {
		compType[c.ID] = c.Type
	}

	b.WriteString("digraph trustward {\n")
	b.WriteString("    rankdir=TB;\n")
	b.WriteString("    node [shape=box, style=\"filled,rounded\", fontname=\"Helvetica\"];\n")
	b.WriteString("    edge [fontname=\"Helvetica\", fontsize=10, color=\"#555555\"];\n\n")

	writeNode := func(indent, id string) {
		c := typeColor(compType[id])
		fmt.Fprintf(&b, "%s%s [label=%s, fillcolor=\"%s\", color=\"%s\", fontcolor=\"%s\"];\n",
			indent, quote(id), quote(idx.Label(id)), c.fill, c.stroke, c.text)
	}

	inZone := make(map[string]bool)
	for i, zone := range proj.TrustZones {
		zc := zoneColors[i%len(zoneColors)]
		fmt.Fprintf(&b, "    subgraph cluster_%s {\n", toDotID(zone.ID))
		fmt.Fprintf(&b, "        label=%s;\n", quote(zone.Title))
		fmt.Fprintf(&b, "        style=\"filled,rounded\"; color=\"%s\"; fillcolor=\"%s\"; fontcolor=\"%s\";\n", zc.stroke, zc.fill, zc.text)
		for _, memberID := range zone.Members {
			writeNode("        ", memberID)
			inZone[memberID] = true
		}
		b.WriteString("    }\n")
	}

	for _, comp := range proj.Components {
		if !inZone[comp.ID] {
			writeNode("    ", comp.ID)
		}
	}

	b.WriteString("\n")
	for _, flow := range proj.DataFlows {
		if len(flow.Connects) != 2 {
			continue
		}
		from, to := quote(flow.Connects[0]), quote(flow.Connects[1])
		if len(flow.Assets) > 0 {
			fmt.Fprintf(&b, "    %s -> %s [label=%s];\n", from, to, quote(strings.Join(flow.Assets, ", ")))
		} else {
			fmt.Fprintf(&b, "    %s -> %s;\n", from, to)
		}
	}

	b.WriteString("}\n")
	return b.String()
}

// quote wraps a string as a DOT double-quoted ID, escaping the two characters
// that are special inside one.
func quote(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return "\"" + s + "\""
}

// toDotID makes a cluster-name-safe identifier. Graphviz treats a subgraph as a
// cluster only when its name begins with "cluster"; keep the rest bare.
func toDotID(id string) string {
	return strings.NewReplacer("-", "_", " ", "_").Replace(id)
}
