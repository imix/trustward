package model_test

import (
	"reflect"
	"testing"

	"github.com/imix/trustward/internal/model"
)

func TestIndex_InversionsAndTitles(t *testing.T) {
	p := &model.Project{
		Assets: []model.Asset{
			{ID: "data", Objectives: []string{"integrity"}},
			{ID: "key", Objectives: []string{"integrity", "confidentiality"}},
		},
		Components: []model.Component{
			{ID: "ctrl", Title: "Controller", Assets: []string{"data", "key"}, Controls: []string{"c-sign"}},
			{ID: "reader", Assets: []string{"key"}}, // no title → falls back to id
		},
		DataFlows: []model.DataFlow{{ID: "flow-sync", Title: "Rights sync"}},
		Controls:  []model.Control{{ID: "c-sign", Title: "Signed updates", Ref: "cat::req-1"}},
	}
	i := model.NewIndex(p)

	if got := i.ComponentsByAsset()["key"]; !reflect.DeepEqual(got, []string{"ctrl", "reader"}) {
		t.Errorf("ComponentsByAsset[key] = %v, want [ctrl reader]", got)
	}
	if got := i.AssetsByObjective()["integrity"]; !reflect.DeepEqual(got, []string{"data", "key"}) {
		t.Errorf("AssetsByObjective[integrity] = %v, want [data key]", got)
	}
	if got := i.ComponentsByControl()["c-sign"]; !reflect.DeepEqual(got, []string{"ctrl"}) {
		t.Errorf("ComponentsByControl[c-sign] = %v, want [ctrl]", got)
	}
	if got := i.ControlsByRequirement()["cat::req-1"]; !reflect.DeepEqual(got, []string{"c-sign"}) {
		t.Errorf("ControlsByRequirement[cat::req-1] = %v, want [c-sign]", got)
	}
	if got := i.Label("reader"); got != "reader" { // title fallback to id
		t.Errorf("Label(reader) = %q, want id fallback", got)
	}
	if got := i.TargetTitle("flow-sync"); got != "Rights sync" { // data-flow target
		t.Errorf("TargetTitle(flow-sync) = %q, want \"Rights sync\"", got)
	}
}
