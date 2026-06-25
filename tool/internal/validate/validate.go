// Package validate checks referential integrity of a loaded project:
// every cross-reference between entities must resolve to a declared ID,
// and IDs must be unique within their entity kind.
// Requirement `satisfies` entries are exempt — they point at external
// standards that need not be loaded into the model.
package validate

import (
	"fmt"

	"sectrack/internal/model"
	"sectrack/internal/risk"
)

// validTreatments are the CRA risk treatment decisions (prEN 40000-1-2 §6.6).
var validTreatments = map[string]bool{
	"mitigate": true, "accept": true, "transfer": true, "avoid": true,
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

// Check returns all referential-integrity issues in the project.
// A clean project yields no issues.
func Check(p *model.Project) []Issue {
	c := &checker{}

	assets := idSet(c, "asset", p.Assets, func(a model.Asset) string { return a.ID })
	controls := idSet(c, "control", p.Controls, func(ct model.Control) string { return ct.ID })
	components := idSet(c, "component", p.Components, func(cp model.Component) string { return cp.ID })
	flows := idSet(c, "data flow", p.DataFlows, func(f model.DataFlow) string { return f.ID })
	idSet(c, "trust zone", p.TrustZones, func(z model.TrustZone) string { return z.ID })
	idSet(c, "threat", p.Threats, func(t model.Threat) string { return t.ID })

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

	for _, cp := range p.Components {
		subject := fmt.Sprintf("component %q", cp.ID)
		for _, a := range cp.Assets {
			if !assets[a] {
				c.add(subject, fmt.Sprintf("asset %q does not match any asset", a))
			}
		}
		for _, ctrl := range cp.Controls {
			if !controls[ctrl] {
				c.add(subject, fmt.Sprintf("control %q does not match any control", ctrl))
			}
		}
	}

	for _, z := range p.TrustZones {
		subject := fmt.Sprintf("trust zone %q", z.ID)
		for _, m := range z.Members {
			if !components[m] {
				c.add(subject, fmt.Sprintf("member %q does not match any component", m))
			}
		}
	}

	for _, f := range p.DataFlows {
		subject := fmt.Sprintf("data flow %q", f.ID)
		if len(f.Connects) != 2 {
			c.add(subject, fmt.Sprintf("must connect exactly 2 components, has %d", len(f.Connects)))
		}
		for _, cp := range f.Connects {
			if !components[cp] {
				c.add(subject, fmt.Sprintf("connects %q does not match any component", cp))
			}
		}
		for _, a := range f.Assets {
			if !assets[a] {
				c.add(subject, fmt.Sprintf("asset %q does not match any asset", a))
			}
		}
	}

	for _, ctrl := range p.Controls {
		if ctrl.Ref != "" && !requirements[ctrl.Ref] {
			c.add(fmt.Sprintf("control %q", ctrl.ID), fmt.Sprintf("ref %q does not match any catalog requirement", ctrl.Ref))
		}
	}

	for _, t := range p.Threats {
		subject := fmt.Sprintf("threat %q", t.ID)
		if t.Ref != "" && !patterns[t.Ref] {
			c.add(subject, fmt.Sprintf("ref %q does not match any threat catalog pattern", t.Ref))
		}
		if t.Target != "" && !components[t.Target] && !flows[t.Target] {
			c.add(subject, fmt.Sprintf("target %q does not match any component or data flow", t.Target))
		}
		if t.Asset != "" && !assets[t.Asset] {
			c.add(subject, fmt.Sprintf("asset %q does not match any asset", t.Asset))
		}
		for _, m := range t.Mitigations {
			if !controls[m] {
				c.add(subject, fmt.Sprintf("mitigation %q does not match any control", m))
			}
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
