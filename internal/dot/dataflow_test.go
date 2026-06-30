package dot_test

import (
	"strings"
	"testing"

	"github.com/imix/trustward/internal/dot"
	"github.com/imix/trustward/internal/model"
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

func TestDataFlow_ComponentInTrustZoneCluster(t *testing.T) {
	proj := &model.Project{
		TrustZones: []model.TrustZone{
			{ID: "zone-a", Title: "Zone A", Members: []string{"comp-a"}},
		},
		Components: []model.Component{
			{ID: "comp-a", Type: "server"},
		},
	}

	got := dot.DataFlow(proj)

	assertContains(t, got, "subgraph cluster_zone_a {")
	assertContains(t, got, `label="Zone A"`)
	assertContains(t, got, `"comp-a" [label="comp-a"`) // no title — falls back to ID
}

func TestDataFlow_ComponentTitleUsedAsLabel(t *testing.T) {
	proj := &model.Project{
		Components: []model.Component{
			{ID: "comp-a", Title: "My Component", Type: "server"},
		},
	}

	got := dot.DataFlow(proj)

	assertContains(t, got, `"comp-a" [label="My Component"`)
}

func TestDataFlow_UnzonedComponentIsTopLevel(t *testing.T) {
	proj := &model.Project{
		TrustZones: []model.TrustZone{
			{ID: "zone-a", Title: "Zone A", Members: []string{"comp-a"}},
		},
		Components: []model.Component{
			{ID: "comp-a", Type: "server"},
			{ID: "comp-b", Type: "server"},
		},
	}

	got := dot.DataFlow(proj)

	assertContains(t, got, `"comp-b" [label="comp-b"`)
	// comp-b must not be declared inside the zone-a cluster block.
	clusterBlock := got[strings.Index(got, "subgraph cluster_zone_a"):strings.Index(got, "    }")]
	assertNotContains(t, clusterBlock, `"comp-b"`)
}

func TestDataFlow_EdgeWithLabel(t *testing.T) {
	proj := &model.Project{
		Components: []model.Component{
			{ID: "comp-a", Type: "server"},
			{ID: "comp-b", Type: "server"},
		},
		DataFlows: []model.DataFlow{
			{ID: "flow-ab", Connects: []string{"comp-a", "comp-b"}, Assets: []string{"asset-x"}},
		},
	}

	got := dot.DataFlow(proj)

	assertContains(t, got, `"comp-a" -> "comp-b" [label="asset-x"];`)
}

func TestDataFlow_EdgeWithoutLabel(t *testing.T) {
	proj := &model.Project{
		Components: []model.Component{
			{ID: "comp-a", Type: "server"},
			{ID: "comp-b", Type: "server"},
		},
		DataFlows: []model.DataFlow{
			{ID: "flow-ab", Connects: []string{"comp-a", "comp-b"}},
		},
	}

	got := dot.DataFlow(proj)

	assertContains(t, got, `"comp-a" -> "comp-b";`)
}

func TestDataFlow_MultipleAssetsJoinedInLabel(t *testing.T) {
	proj := &model.Project{
		Components: []model.Component{
			{ID: "comp-a", Type: "server"},
			{ID: "comp-b", Type: "server"},
		},
		DataFlows: []model.DataFlow{
			{ID: "flow-ab", Connects: []string{"comp-a", "comp-b"}, Assets: []string{"asset-x", "asset-y"}},
		},
	}

	got := dot.DataFlow(proj)

	assertContains(t, got, `[label="asset-x, asset-y"]`)
}

func TestDataFlow_ClusterNameSanitisedNodeIDsQuoted(t *testing.T) {
	proj := &model.Project{
		TrustZones: []model.TrustZone{
			{ID: "my-zone", Title: "My Zone", Members: []string{"my-comp"}},
		},
		Components: []model.Component{
			{ID: "my-comp", Type: "server"},
		},
	}

	got := dot.DataFlow(proj)

	assertContains(t, got, "subgraph cluster_my_zone {") // cluster name sanitised
	assertContains(t, got, `"my-comp" [label="my-comp"`) // node id stays quoted kebab
	assertNotContains(t, got, "subgraph cluster_my-zone")
}
