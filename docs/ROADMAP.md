# Roadmap

trustward's risk-management layer landed in four phases. CRA conformance is
fully covered; everything under **Remaining** is optional, LOW priority, and
built on demand — it adds depth, not coverage.

## Done

**Phase 1 — Risk management.** Threats carry
`likelihood`/`impact`/`treatment`/`owner`/`decided`; a project-level
`risk-policy` sets the scoring method and acceptance criteria. The risk level is
computed (`internal/risk`, qualitative 3×3 matrix) and `validate` enforces the
CRA gate: every non-accepted risk needs a treatment and an owner. The report
shows a risk register.

**Phase 2 — ETSI attack-potential profile.** The `risk.Scorer` seam gained a
second profile, `etsi-tvra` (`internal/risk/etsi.go`) — attack-potential factors
on a threat's `attack` block sum to an attack potential that maps inversely to
likelihood, then the shared matrix. Selected via `risk-policy.method: etsi-tvra`.

**Phase 3 — prEN 40000-1-2 §6 report shape.** A Risk Acceptance Criteria and
Methodology section (§6.3, from the risk-policy), a Risk Register with an
Evaluation column marking each risk accepted/treated/**open** (§6.5.4–5), and the
control→requirement coverage reframed as Compliance Evidence (§6.6). The "open
risk" rule is defined once in `risk.Evaluate` and shared by the report and the
validate CRA gate.

**Phase 4 — Register polish.** `Scorer` returns a `risk.Score` (`{Level,
Likelihood}`) instead of a bare level, so the derived likelihood is no longer
discarded. `risk.Eval` embeds `Score`; `risk.Evaluate` is the single scoring
entry point. The register's Likelihood column shows the derived likelihood for
etsi-tvra threats (was blank).

**Cybersecurity objectives (§6.5.2).** `asset.objectives[]` and
`threat.violates[]` give the objective → asset → threat trace, on a CIA-extended
scale (confidentiality/integrity/availability/authenticity/accountability).

**Threat template library → threat catalogs** (commit 3807e8c). Reusable threat
patterns shipped as catalogs with ref-based inheritance
(`ref: catalog-id::pattern-id`), rather than a separate templating mechanism.

## Remaining — all LOW priority, build on demand

Each is **additive** — new optional model types, loader merges, validate checks,
and report sections — reusing the established patterns:

- model types in `internal/model/types.go`, `Project` in `project.go`
- loader merge in `internal/project/project.go` (generic `mergeList` for list keys; singletons first-wins)
- validation in `internal/validate/validate.go` (`idSet`, ref checks, `checkRisk`)
- scoring seam in `internal/risk/` (`Scorer.Score`, `Evaluate` as the single entry point; `MethodKnown`)
- report in `internal/quarto/report.go` + `internal/quarto/templates/report.tmpl` (built-in); exported to `report.tmpl`

Each ships TDD (one RED→GREEN per slice), keeps existing models working (new
fields optional), and updates MODEL.md + GLOSSARY.md + the example.

### ETSI-completeness (Phase C)

Each is self-contained; do only when a workflow or obligation calls for it.

**C1 — First-class threat agents (motivation + capability).** Reusable attacker
profiles instead of per-threat inline `attack` blocks.
- `ThreatAgent{ID, Title, Description, Expertise, Knowledge, Opportunity, Equipment, Motivation, Capability}`.
- `Threat.Agent string` (agent id). ETSI scorer reads factors from the referenced
  agent when set, else the inline `attack` block (back-compat).
- Validate: agent ref resolves; motivation/capability in their scales.
- `motivation`/`capability` modelled + shown; folding them into the likelihood is optional.

**C2 — Countermeasure cost-benefit analysis (ETSI Annex H).** Justify control spend.
- Extend `Control` with `Cost map[string]string` (category → impact; categories:
  standards-design/implementation/operation/regulatory/market-acceptance) and
  `Benefit []{RiskLevel, Original, Revised int}`.
- Validate categories + a `cost_impact` scale.
- Report: a cost-benefit section (reuse the `tvra` Annex H layout).

**C3 — Unwanted incidents.** The consequence layer (threats hit a technical
asset; incidents are the downstream harm).
- `UnwantedIncident{ID, Title, Description}`; link via `Threat.Incidents []string`
  or a `problems-to-avoid` list.
- Report: a consequences view.

**C4 — Attack intensity.**
- `Threat.Intensity` ∈ `single|moderate|high` (cumulative-impact qualifier, ETSI
  clause 6.8.1). Validate against the scale; optionally adjust impact.

**C5 — Citations (provenance).**
- `Citation{ID, Publisher, Document}`; optional `citation` field on
  assets/objectives/requirements. Complements per-control `evidence`.

**C6 — ETSI scales as a report appendix.**
- When `method: etsi-tvra`, append the factor scales with their definitions
  (expertise/knowledge/opportunity/equipment/motivation/capability/intensity) as
  reference material, so the computed numbers are auditable. Also the home for the
  attack-potential-band detail if an assessor needs the raw number.

### Tooling / ergonomics

**Diagram scale.** Data flow diagrams become unreadable on large systems (20+
components, 5+ trust zones). Proposed filtering options:
- `--zone <id>` — render only components in that trust zone plus cross-boundary flows
- `--component <id>` — one-hop neighbourhood view
- `--cross-zone-only` — drop intra-zone flows

**YAML schema validation.** Validate YAML files against a schema on load to give
actionable errors instead of silent zero-values. Deferred until file structures
stabilise.
