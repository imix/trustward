package risk

import "github.com/imix/trustward/internal/model"

// MatrixCell is one likelihood×impact position: the risk level that position
// resolves to (its colour, independent of contents) and the threats that land
// there.
type MatrixCell struct {
	Level   string
	Threats []string // threat IDs
}

// MatrixRow is one likelihood band across all impact columns.
type MatrixRow struct {
	Likelihood string
	Cells      []MatrixCell // aligned with Matrix.Impacts
}

// Matrix is the likelihood×impact heatmap. Rows run high→low likelihood and
// columns low→high impact — the conventional orientation, worst risk top-right.
// Unplaced holds threats with no likelihood or impact on the scale (scored by
// the severity fallback, so they have no grid coordinate).
type Matrix struct {
	Impacts  []string
	Rows     []MatrixRow
	Unplaced []string
}

// MatrixOf buckets every threat into the likelihood×impact grid. Likelihood
// comes from the computed Eval (declared or derived), impact from the threat;
// cell colour is level() of the coordinate, so empty cells are still shaded.
func MatrixOf(p *model.Project) Matrix {
	impacts := []string{"low", "medium", "high"}
	likelihoods := []string{"high", "medium", "low"}

	evals := Evaluate(p)
	bucket := map[string]map[string][]string{}
	for _, lk := range likelihoods {
		bucket[lk] = map[string][]string{}
	}
	var unplaced []string
	for _, t := range p.Threats {
		lk, im := evals[t.ID].Likelihood, t.Impact
		if !InScale(lk) || !InScale(im) {
			unplaced = append(unplaced, t.ID)
			continue
		}
		bucket[lk][im] = append(bucket[lk][im], t.ID)
	}

	rows := make([]MatrixRow, 0, len(likelihoods))
	for _, lk := range likelihoods {
		cells := make([]MatrixCell, 0, len(impacts))
		for _, im := range impacts {
			cells = append(cells, MatrixCell{Level: level(lk, im), Threats: bucket[lk][im]})
		}
		rows = append(rows, MatrixRow{Likelihood: lk, Cells: cells})
	}
	return Matrix{Impacts: impacts, Rows: rows, Unplaced: unplaced}
}
