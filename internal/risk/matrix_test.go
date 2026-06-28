package risk

import (
	"testing"

	"github.com/imix/trustward/internal/model"
)

// MatrixOf must bucket scored threats by (likelihood, impact), colour each cell
// by its coordinate, and divert anything off-scale to Unplaced.
func TestMatrixOf(t *testing.T) {
	p := &model.Project{
		Threats: []model.Threat{
			{ID: "a", Likelihood: "high", Impact: "high"},   // top-right
			{ID: "b", Likelihood: "high", Impact: "high"},   // same cell
			{ID: "c", Likelihood: "low", Impact: "low"},     // bottom-left
			{ID: "d", Severity: "high"},                     // no l/i → unplaced
		},
	}
	m := MatrixOf(p)

	// rows high→low, impacts low→high
	if m.Rows[0].Likelihood != "high" || m.Impacts[2] != "high" {
		t.Fatalf("orientation: rows[0]=%s impacts=%v", m.Rows[0].Likelihood, m.Impacts)
	}
	topRight := m.Rows[0].Cells[2]
	if len(topRight.Threats) != 2 || topRight.Level != "critical" {
		t.Errorf("top-right: got %+v, want 2 threats / critical", topRight)
	}
	bottomLeft := m.Rows[2].Cells[0]
	if len(bottomLeft.Threats) != 1 || bottomLeft.Level != "low" {
		t.Errorf("bottom-left: got %+v, want 1 threat / low", bottomLeft)
	}
	if len(m.Unplaced) != 1 || m.Unplaced[0] != "d" {
		t.Errorf("unplaced: got %v, want [d]", m.Unplaced)
	}
}
