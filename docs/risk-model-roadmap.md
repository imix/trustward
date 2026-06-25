# Risk-model roadmap: remaining ETSI-completeness items

The CRA-relevant work is done: the ETSI attack-potential scoring profile, the
attack-potential-derived likelihood in the register, cybersecurity objectives
(§6.5.2), and the clause-mapped prEN 40000-1-2 §6 report format all shipped.
What remains are the **Phase C** ETSI-completeness items below — all LOW
priority, none required for CRA conformance, each built on demand.

Each is **additive** — new model types, loader merges, validate checks, and
report sections — reusing the established patterns:

- model types in `tool/internal/model/types.go`, `Project` in `project.go`
- loader merge in `tool/internal/project/project.go` (generic `mergeList` for list keys; singletons first-wins)
- validation in `tool/internal/validate/validate.go` (`idSet`, ref checks, `checkRisk`)
- scoring seam in `tool/internal/risk/` (`Scorer.Score`, `Evaluate` as the single entry point; `MethodKnown`)
- report in `tool/internal/quarto/threatmodel.go` + `templates/threat-model.tmpl`

Each item ships TDD (one RED→GREEN per slice), keeps existing models working
(new fields optional), and updates MODEL.md + GLOSSARY.md + the example.

---

## Phase C — ETSI-completeness items  ·  priority: LOW (build on demand)

Each is self-contained; do only when a workflow or obligation calls for it.

### C1 — First-class threat agents (motivation + capability)
Reusable attacker profiles instead of per-threat inline `attack` blocks.
- `ThreatAgent{ID, Title, Description, Expertise, Knowledge, Opportunity, Equipment, Motivation, Capability}`.
- `Threat.Agent string` (agent id). ETSI scorer reads factors from the referenced
  agent when set, else the inline `attack` block (back-compat).
- Validate: agent ref resolves; motivation/capability in their scales.
- `motivation`/`capability` modelled + shown; folding them into the likelihood is optional.

### C2 — Countermeasure cost-benefit analysis (ETSI Annex H)
Justify control spend.
- Extend `Control` with `Cost map[string]string` (category → impact; categories:
  standards-design/implementation/operation/regulatory/market-acceptance) and
  `Benefit []{RiskLevel, Original, Revised int}`.
- Validate categories + a `cost_impact` scale.
- Report: a cost-benefit section (reuse the `tvra` Annex H layout).

### C3 — Unwanted incidents
The consequence layer (ties to the "XSS → integrity of executed code → cascade"
discussion: threats hit a technical asset, incidents are the downstream harm).
- `UnwantedIncident{ID, Title, Description}`; link via `Threat.Incidents []string`
  or a `problems-to-avoid` list.
- Report: a consequences view.

### C4 — Attack intensity
- `Threat.Intensity` ∈ `single|moderate|high` (cumulative-impact qualifier, ETSI
  clause 6.8.1). Validate against the scale; optionally adjust impact.

### C5 — Citations (provenance)
- `Citation{ID, Publisher, Document}`; optional `citation` field on
  assets/objectives/requirements. Complements per-control `evidence`.

### C6 — ETSI scales as a report appendix
- When `method: etsi-tvra`, append the factor scales with their definitions
  (expertise/knowledge/opportunity/equipment/motivation/capability/intensity) as
  reference material, so the computed numbers are auditable. Also the home for the
  Phase B `Basis`/attack-potential-band detail if an assessor needs the raw number.
