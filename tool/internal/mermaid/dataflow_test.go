package mermaid_test

import (
	"strings"
	"testing"

	"sectrack/internal/mermaid"
	"sectrack/internal/model"
)

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Errorf("output does not contain %q\ngot:\n%s", want, got)
	}
}

func assertNotContains(t *testing.T, got, want string) {
	t.Helper()
	if strings.Contains(got, want) {
		t.Errorf("output should not contain %q\ngot:\n%s", want, got)
	}
}

func TestDataFlow_ComponentInTrustZone(t *testing.T) {
	sys := &model.SystemFile{
		TrustZones: []model.TrustZone{
			{ID: "zone-a", Title: "Zone A", Members: []string{"comp-a"}},
		},
		Components: []model.Component{
			{ID: "comp-a", Type: "server"},
		},
	}

	got := mermaid.DataFlow(sys)

	assertContains(t, got, `subgraph zone_a["Zone A"]`)
	assertContains(t, got, `comp_a["comp-a"]`)
}

func TestDataFlow_UnzonedComponentIsTopLevel(t *testing.T) {
	sys := &model.SystemFile{
		TrustZones: []model.TrustZone{
			{ID: "zone-a", Title: "Zone A", Members: []string{"comp-a"}},
		},
		Components: []model.Component{
			{ID: "comp-a", Type: "server"},
			{ID: "comp-b", Type: "server"},
		},
	}

	got := mermaid.DataFlow(sys)

	// comp-b has no zone — must appear as a bare node, not inside a subgraph
	assertContains(t, got, `comp_b["comp-b"]`)
	// verify it is not listed inside the zone-a subgraph block
	zoneBlock := got[strings.Index(got, "subgraph zone_a"):strings.Index(got, "    end")]
	assertNotContains(t, zoneBlock, `comp_b`)
}

func TestDataFlow_EdgeWithLabel(t *testing.T) {
	sys := &model.SystemFile{
		Components: []model.Component{
			{ID: "comp-a", Type: "server"},
			{ID: "comp-b", Type: "server"},
		},
		DataFlows: []model.DataFlow{
			{ID: "flow-ab", Title: "A to B", Connects: []string{"comp-a", "comp-b"}, Assets: []string{"asset-x"}},
		},
	}

	got := mermaid.DataFlow(sys)

	assertContains(t, got, `comp_a -->|"asset-x"| comp_b`)
}

func TestDataFlow_EdgeWithoutLabel(t *testing.T) {
	sys := &model.SystemFile{
		Components: []model.Component{
			{ID: "comp-a", Type: "server"},
			{ID: "comp-b", Type: "server"},
		},
		DataFlows: []model.DataFlow{
			{ID: "flow-ab", Title: "A to B", Connects: []string{"comp-a", "comp-b"}},
		},
	}

	got := mermaid.DataFlow(sys)

	assertContains(t, got, `comp_a --> comp_b`)
}

func TestDataFlow_HyphensConvertedToUnderscores(t *testing.T) {
	sys := &model.SystemFile{
		TrustZones: []model.TrustZone{
			{ID: "my-zone", Title: "My Zone", Members: []string{"my-comp"}},
		},
		Components: []model.Component{
			{ID: "my-comp", Type: "server"},
		},
	}

	got := mermaid.DataFlow(sys)

	assertContains(t, got, `subgraph my_zone["My Zone"]`)
	assertContains(t, got, `my_comp["my-comp"]`)
	assertNotContains(t, got, `subgraph my-zone`)
	assertNotContains(t, got, `my-comp[`)
}

func TestDataFlow_MultipleAssetsJoinedInLabel(t *testing.T) {
	sys := &model.SystemFile{
		Components: []model.Component{
			{ID: "comp-a", Type: "server"},
			{ID: "comp-b", Type: "server"},
		},
		DataFlows: []model.DataFlow{
			{
				ID:       "flow-ab",
				Title:    "A to B",
				Connects: []string{"comp-a", "comp-b"},
				Assets:   []string{"asset-x", "asset-y"},
			},
		},
	}

	got := mermaid.DataFlow(sys)

	assertContains(t, got, `comp_a -->|"asset-x, asset-y"| comp_b`)
}
