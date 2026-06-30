// Package validate checks referential integrity of a loaded project:
// every cross-reference between entities must resolve to a declared ID,
// and IDs must be unique within their entity kind.
// Requirement `satisfies` entries are exempt — they point at external
// standards that need not be loaded into the model.
//
// Structural shape and the closed vocabularies (severity, treatment, objective
// type, attack factors, risk-policy method, level scales, data-flow arity,
// reference version) are the schema's job — package schema, which `validate`
// runs first. This package covers only what a per-file schema can't: relations
// between entities and the CRA gate.
package validate

import (
	"fmt"

	"github.com/imix/trustward/internal/model"
	"github.com/imix/trustward/internal/risk"
)

// Issue is a single referential-integrity finding.
type Issue struct {
	Subject string // the entity holding the bad reference, e.g. `threat "threat-x"`
	Message string
}

func (i Issue) String() string {
	return i.Subject + ": " + i.Message
}

// checker accumulates issues while building ID lookup sets.
type checker struct {
	issues []Issue
}

// idSet builds the set of declared IDs for one entity kind, reporting
// missing (empty) and duplicate IDs along the way. Entities with an empty
// ID are left out of the set so they neither resolve references nor get
// double-reported as duplicates.
func idSet[E any](c *checker, kind string, items []E, id func(E) string) map[string]bool {
	set := make(map[string]bool, len(items))
	for i, item := range items {
		v := id(item)
		if v == "" {
			c.add(fmt.Sprintf("%s #%d", kind, i+1), "missing id")
			continue
		}
		if set[v] {
			c.add(fmt.Sprintf("%s %q", kind, v), "duplicate id")
		}
		set[v] = true
	}
	return set
}

func (c *checker) add(subject, message string) {
	c.issues = append(c.issues, Issue{subject, message})
}

// oneOf wraps an optional scalar reference as a 0-or-1 element slice, so a
// single-valued field flows through resolveRefs like a list one. "" means unset.
func oneOf(s string) []string {
	if s == "" {
		return nil
	}
	return []string{s}
}

// resolveRefs reports any value in each item's reference field that does not
// resolve to a declared id in target. noun labels the field in the message
// ("member", "mitigation"); targetLabel names what it must resolve to
// ("component"). The single place the resolve-or-report rule lives — every
// cross-reference edge in Check is one call.
func resolveRefs[E any](c *checker, kind string, items []E, id func(E) string, noun, targetLabel string, refs func(E) []string, target map[string]bool) {
	for _, it := range items {
		subject := fmt.Sprintf("%s %q", kind, id(it))
		for _, r := range refs(it) {
			if !target[r] {
				c.add(subject, fmt.Sprintf("%s %q does not match any %s", noun, r, targetLabel))
			}
		}
	}
}

// Check returns all referential-integrity issues in the project.
// A clean project yields no issues.
func Check(p *model.Project) []Issue {
	c := &checker{}

	assetID := func(a model.Asset) string { return a.ID }
	compID := func(cp model.Component) string { return cp.ID }
	flowID := func(f model.DataFlow) string { return f.ID }
	threatID := func(t model.Threat) string { return t.ID }
	controlID := func(ct model.Control) string { return ct.ID }
	objID := func(o model.Objective) string { return o.ID }

	assets := idSet(c, "asset", p.Assets, assetID)
	objectives := idSet(c, "objective", p.Objectives, objID)
	controls := idSet(c, "control", p.Controls, controlID)
	components := idSet(c, "component", p.Components, compID)
	flows := idSet(c, "data flow", p.DataFlows, flowID)
	idSet(c, "trust zone", p.TrustZones, func(z model.TrustZone) string { return z.ID })
	idSet(c, "threat", p.Threats, threatID)
	idSet(c, "reference", p.References, func(r model.Reference) string { return r.ID })

	patterns := make(map[string]bool)
	for _, cat := range p.ThreatCatalogs {
		for _, pat := range cat.Patterns {
			patterns[cat.ID+"::"+pat.ID] = true
		}
	}
	requirements := make(map[string]bool)
	for _, cat := range p.Catalogs {
		for _, req := range cat.Requirements {
			requirements[cat.ID+"::"+req.ID] = true
		}
	}

	// A threat target resolves against components and data flows alike.
	targetable := make(map[string]bool, len(components)+len(flows))
	for id := range components {
		targetable[id] = true
	}
	for id := range flows {
		targetable[id] = true
	}

	// Every cross-reference edge in the model, one line each — mirrors the
	// "Cross-reference rules" table in docs/MODEL.md. resolveRefs holds the
	// single resolve-or-report rule.
	resolveRefs(c, "asset", p.Assets, assetID, "objective", "objective",
		func(a model.Asset) []string { return a.Objectives }, objectives)
	resolveRefs(c, "component", p.Components, compID, "asset", "asset",
		func(cp model.Component) []string { return cp.Assets }, assets)
	resolveRefs(c, "component", p.Components, compID, "control", "control",
		func(cp model.Component) []string { return cp.Controls }, controls)
	resolveRefs(c, "trust zone", p.TrustZones, func(z model.TrustZone) string { return z.ID }, "member", "component",
		func(z model.TrustZone) []string { return z.Members }, components)
	resolveRefs(c, "data flow", p.DataFlows, flowID, "connects", "component",
		func(f model.DataFlow) []string { return f.Connects }, components)
	resolveRefs(c, "data flow", p.DataFlows, flowID, "asset", "asset",
		func(f model.DataFlow) []string { return f.Assets }, assets)
	resolveRefs(c, "control", p.Controls, controlID, "ref", "catalog requirement",
		func(ct model.Control) []string { return oneOf(ct.Ref) }, requirements)
	resolveRefs(c, "threat", p.Threats, threatID, "ref", "threat catalog pattern",
		func(t model.Threat) []string { return oneOf(t.Ref) }, patterns)
	resolveRefs(c, "threat", p.Threats, threatID, "target", "component or data flow",
		func(t model.Threat) []string { return oneOf(t.Target) }, targetable)
	resolveRefs(c, "threat", p.Threats, threatID, "asset", "asset",
		func(t model.Threat) []string { return oneOf(t.Asset) }, assets)
	resolveRefs(c, "threat", p.Threats, threatID, "mitigation", "control",
		func(t model.Threat) []string { return t.Mitigations }, controls)
	resolveRefs(c, "threat", p.Threats, threatID, "violates", "objective",
		func(t model.Threat) []string { return t.Violates }, objectives)
	resolveRefs(c, "threat", p.Threats, threatID, "backed-by", "threat catalog pattern",
		func(t model.Threat) []string { return t.BackedBy }, patterns)

	checkRisk(c, p)

	return c.issues
}

// checkRisk enforces the CRA gate: every risk whose computed level is not
// accepted must carry a treatment and an owner. The gate only applies when a
// risk-policy is declared, so models that don't use the risk layer are
// unaffected. (Field vocabularies — treatment, scales, attack factors,
// residual level — are the schema's job; see the package doc.)
func checkRisk(c *checker, p *model.Project) {
	if !p.RiskPolicy.Set {
		return // no policy declared → no CRA gate
	}
	eval := risk.Evaluate(p)
	for _, t := range p.Threats {
		subject := fmt.Sprintf("threat %q", t.ID)
		e := eval[t.ID]
		if e.Open() {
			// A mitigate with no control records an intention, not a reduction —
			// it stays open. Name that case so the fix is obvious.
			if t.Treatment == "mitigate" && len(t.Mitigations) == 0 {
				c.add(subject, fmt.Sprintf("%s risk: mitigate treatment needs a mitigations control — a mitigate with no control is a plan, not a current reduction (use accept with an owner to acknowledge it instead)", e.Level))
			} else {
				c.add(subject, fmt.Sprintf("%s risk is not accepted and needs a treatment + owner", e.Level))
			}
		}
		// A residual below the computed level must be earned by a control.
		if t.ResidualRisk != "" && risk.LevelRank(t.ResidualRisk) < risk.LevelRank(e.Level) && len(t.Mitigations) == 0 {
			c.add(subject, fmt.Sprintf("residualRisk %q is below the computed %s risk but no mitigations control is recorded to justify the reduction", t.ResidualRisk, e.Level))
		}
	}
}
