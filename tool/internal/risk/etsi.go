package risk

import "sectrack/internal/model"

// ETSI scores using the attack-potential method of ETSI TS 102 165-1 (TVRA):
// the four attacker factors sum to an attack potential, which maps (inversely)
// to a likelihood, then combined with the threat's impact via the shared matrix.
// A harder attack (higher potential) means a lower likelihood.
type ETSI struct{}

// Factor weights, clause 6.6.3. A threat's attack: block names one value per factor.
var (
	etsiExpertise   = map[string]int{"layman": 0, "proficient": 3, "expert": 6, "multiple-experts": 8}
	etsiKnowledge   = map[string]int{"public": 0, "restricted": 3, "sensitive": 7, "critical": 11}
	etsiOpportunity = map[string]int{"unlimited": 0, "easy": 1, "moderate": 4, "difficult": 10, "none": 999}
	etsiEquipment   = map[string]int{"standard": 0, "specialised": 3, "bespoke": 7, "multiple-bespoke": 9}
)

// etsiFactors maps an attack factor name to its weight table, for validation.
var etsiFactors = map[string]map[string]int{
	"expertise": etsiExpertise, "knowledge": etsiKnowledge,
	"opportunity": etsiOpportunity, "equipment": etsiEquipment,
}

// InAttackScale reports whether value is valid for the given attack factor.
func InAttackScale(factor, value string) bool {
	_, ok := etsiFactors[factor][value]
	return ok
}

func (ETSI) Level(t model.Threat) string {
	a := t.Attack
	if a == nil {
		return ""
	}
	e, ok1 := etsiExpertise[a.Expertise]
	k, ok2 := etsiKnowledge[a.Knowledge]
	o, ok3 := etsiOpportunity[a.Opportunity]
	q, ok4 := etsiEquipment[a.Equipment]
	if !ok1 || !ok2 || !ok3 || !ok4 {
		return "" // invalid/missing factor → unscored; validation reports it
	}
	return level(attackPotentialLikelihood(e+k+o+q), t.Impact)
}

// attackPotentialLikelihood maps the attack-potential sum (banded per clause
// 6.6.3.1) to a qualitative likelihood: the easier the attack, the likelier.
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
