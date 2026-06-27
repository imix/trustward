package risk

import "github.com/imix/trustward/internal/model"

// AttackPotential scores using the attack-potential method: the four attacker
// factors sum to an attack potential, which maps (inversely) to a likelihood,
// then combines with the threat's impact via the shared matrix. A harder attack
// (higher potential) means a lower likelihood.
//
// The factor scale is the one shared by Common Criteria (ISO/IEC 18045) and
// ETSI TS 102 165-1 (TVRA), clause 6.6.3 — the method belongs to neither.
type AttackPotential struct{}

// Factor weights, clause 6.6.3. A threat's attack: block names one value per factor.
var (
	expertiseWeights   = map[string]int{"layman": 0, "proficient": 3, "expert": 6, "multiple-experts": 8}
	knowledgeWeights   = map[string]int{"public": 0, "restricted": 3, "sensitive": 7, "critical": 11}
	opportunityWeights = map[string]int{"unlimited": 0, "easy": 1, "moderate": 4, "difficult": 10, "none": 999}
	equipmentWeights   = map[string]int{"standard": 0, "specialised": 3, "bespoke": 7, "multiple-bespoke": 9}
)

// factorWeights maps an attack factor name to its weight table, for validation.
var factorWeights = map[string]map[string]int{
	"expertise": expertiseWeights, "knowledge": knowledgeWeights,
	"opportunity": opportunityWeights, "equipment": equipmentWeights,
}

// InAttackScale reports whether value is valid for the given attack factor.
func InAttackScale(factor, value string) bool {
	_, ok := factorWeights[factor][value]
	return ok
}

func (AttackPotential) Score(t model.Threat) Score {
	a := t.Attack
	if a == nil {
		return Score{}
	}
	e, ok1 := expertiseWeights[a.Expertise]
	k, ok2 := knowledgeWeights[a.Knowledge]
	o, ok3 := opportunityWeights[a.Opportunity]
	q, ok4 := equipmentWeights[a.Equipment]
	if !ok1 || !ok2 || !ok3 || !ok4 {
		return Score{} // invalid/missing factor → unscored; validation reports it
	}
	lk := attackPotentialLikelihood(e + k + o + q)
	return Score{Level: level(lk, t.Impact), Likelihood: lk}
}

// attackPotentialLikelihood maps the attack-potential sum (banded per clause
// 6.6.3) to a qualitative likelihood: the easier the attack, the likelier.
func attackPotentialLikelihood(sum int) string {
	switch {
	case sum < 7: // Basic / Enhanced-Basic
		return "high"
	case sum < 14: // Moderate
		return "medium"
	default: // High / Beyond High
		return "low"
	}
}
