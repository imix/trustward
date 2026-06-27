// Package validate checks referential integrity of a loaded project:
// every cross-reference between entities must resolve to a declared ID,
// and IDs must be unique within their entity kind.
// Requirement `satisfies` entries are exempt — they point at external
// standards that need not be loaded into the model.
package validate

import (
	"fmt"

	"github.com/imix/trustward/internal/model"
	"github.com/imix/trustward/internal/risk"
)

// validTreatments are the CRA risk treatment decisions (prEN 40000-1-2 §6.6).
var validTreatments = map[string]bool{
	"mitigate": true, "accept": true, "transfer": true, "avoid": true,
}

// validObjectiveTypes are the CIA-scale properties a cybersecurity objective
// may protect (prEN 40000-1-2 §6.5.2).
var validObjectiveTypes = map[string]bool{
	"confidentiality": true, "integrity": true, "availability": true,
	"authenticity": true, "accountability": true,
}

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

	for _, o := range p.Objectives {
		if o.Type != "" && !validObjectiveTypes[o.Type] {
			c.add(fmt.Sprintf("objective %q", o.ID),
				fmt.Sprintf("type %q is not a CIA-scale property (confidentiality/integrity/availability/authenticity/accountability)", o.Type))
		}
	}

	// External references must pin a version — that pin is what the report cites.
	for _, r := range p.References {
		if r.ID != "" && r.Version == "" {
			c.add(fmt.Sprintf("reference %q", r.ID), "missing version")
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

	// Arity: a data flow connects exactly two components.
	for _, f := range p.DataFlows {
		if len(f.Connects) != 2 {
			c.add(fmt.Sprintf("data flow %q", f.ID), fmt.Sprintf("must connect exactly 2 components, has %d", len(f.Connects)))
		}
	}

	checkRisk(c, p)

	return c.issues
}

// checkRisk validates the risk-assessment fields and enforces the CRA gate:
// every risk whose computed level is not accepted must carry a treatment and
// an owner. The gate only applies when a risk-policy is declared, so models
// that don't use the risk layer are unaffected.
func checkRisk(c *checker, p *model.Project) {
	if p.RiskPolicy.Set && !risk.MethodKnown(p.RiskPolicy.Method) {
		c.add("risk-policy", fmt.Sprintf("method %q is not a known scoring profile (qualitative, etsi-tvra)", p.RiskPolicy.Method))
	}
	for _, t := range p.Threats {
		subject := fmt.Sprintf("threat %q", t.ID)
		if t.Treatment != "" && !validTreatments[t.Treatment] {
			c.add(subject, fmt.Sprintf("treatment %q is not one of mitigate/accept/transfer/avoid", t.Treatment))
		}
		if t.Likelihood != "" && !risk.InScale(t.Likelihood) {
			c.add(subject, fmt.Sprintf("likelihood %q is not a valid scale value", t.Likelihood))
		}
		if t.Impact != "" && !risk.InScale(t.Impact) {
			c.add(subject, fmt.Sprintf("impact %q is not a valid scale value", t.Impact))
		}
		if a := t.Attack; a != nil {
			for factor, value := range map[string]string{
				"expertise": a.Expertise, "knowledge": a.Knowledge,
				"opportunity": a.Opportunity, "equipment": a.Equipment,
			} {
				if value != "" && !risk.InAttackScale(factor, value) {
					c.add(subject, fmt.Sprintf("attack %s %q is not a valid scale value", factor, value))
				}
			}
		}
	}

	if !p.RiskPolicy.Set {
		return // no policy declared → no CRA gate
	}
	eval := risk.Evaluate(p)
	for _, t := range p.Threats {
		if e := eval[t.ID]; e.Open() {
			c.add(fmt.Sprintf("threat %q", t.ID),
				fmt.Sprintf("%s risk is not accepted and needs a treatment + owner", e.Level))
		}
	}
}
