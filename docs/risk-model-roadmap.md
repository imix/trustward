# Risk-model roadmap: ETSI-parity items

sectrack was chosen as the base over the ETSI-based `tvra` prototype, and the
ETSI attack-potential method was ported as a scoring profile (`internal/risk/etsi.go`).
This roadmap covers the remaining concepts that exist in `tvra` but not yet in
sectrack. Each item is **additive** — new model types, loader merges, validate
checks, and report sections — reusing the established patterns:

- model types in `tool/internal/model/types.go`, `Project` in `project.go`
- loader merge in `tool/internal/project/project.go` (generic `mergeList` for list keys; singletons first-wins)
- validation in `tool/internal/validate/validate.go` (`idSet`, ref checks, `checkRisk`)
- scoring seam in `tool/internal/risk/` (`Scorer.Score`, `Evaluate` as the single entry point; `MethodKnown`)
- report in `tool/internal/quarto/threatmodel.go` + `templates/threat-model.tmpl`

Priority reflects CRA relevance, not ETSI completeness for its own sake.

---

## Phase A — Cybersecurity objectives  ·  **DONE**

prEN 40000-1-2 §6.5.2 is "Asset **and cybersecurity objective** identification".
sectrack now models objectives: the `Objective` type (CIA-scale `type`),
`Asset.objectives[]`, and `Threat.violates[]`. Shipped across three slices —
model+loader, validate (ref resolution + CIA scale), and an "Assets and
Cybersecurity Objectives" report section (§6.5.2). MODEL.md, GLOSSARY.md, and
the fire-protection example were updated.

**Design**
- New `Objective{ID, Title, Type, Description}` where `Type` ∈ CIA scale
  (`confidentiality|integrity|availability|authenticity|accountability`).
- `Asset.Objectives []string` — the objectives an asset must uphold.
- Optional `Threat.Violates []string` — objectives a threat violates (enables the
  objective→asset→threat trace and ties to §6.5.3).
- Loader: append `objectives:`.
- Validate: `asset.objectives[]` resolve to objectives; `objective.type` in the CIA scale;
  `threat.violates[]` resolve.
- Report: an "Assets and Cybersecurity Objectives" section (§6.5.2) — assets with
  their objectives and CIA types.

**Slices**: (1) model+loader; (2) validate refs+type; (3) report section.
**Files**: types.go, project.go, validate.go, threatmodel.go + template, MODEL/GLOSSARY, example.

---

## Phase B — Attack-potential band in the register  ·  **DONE** (`f35b5aa`)

For `etsi-tvra` threats the register's Likelihood column was blank — likelihood is
computed, not stated. The seam was widened and the derived likelihood now shows.

**Shipped**
- The `Scorer` seam is `Score(t) Score` where `Score{Level, Likelihood}`. `risk.Eval`
  embeds `Score`; the standalone `Score(p)` map folded into `Evaluate` (single entry
  point). The matrix is unchanged.
- Report: the register's Likelihood column shows the derived likelihood for ETSI rows
  (high/medium/low), consistent with qualitative rows.

**Deliberately omitted**: the `Basis` field (e.g. `"attack potential: Moderate (10)"`).
"Derived likelihood only" was chosen over "likelihood + AP band" — the band is the
ETSI-specific extra. Add it back as a `Basis`/band column if an assessor needs the
attack-potential number surfaced; otherwise the appendix in C6 already covers
auditability of the scales.

---

## Phase D — Full EN 40000 report format  ·  priority: HIGH (depends on A)

Phase 3 added the missing §6 *sections* into sectrack's existing layout (System
Overview → Data Flow → Assets → Threats → Acceptance Criteria → Register →
Compliance). This item adopts the prEN 40000-1-2 **format**: restructure the
report so its headings mirror the standard's clause structure and numbering, so
the rendered document reads as an EN 40000 risk-management record an assessor can
map clause-by-clause.

**Target structure (clause → sectrack content)**
- **6.2 Product context** — system overview, components, trust zones, data flow diagram
- **6.3 Risk acceptance criteria and methodology** — the risk-policy (from Phase 3)
- **6.5 Risk assessment**
  - **6.5.2 Asset and cybersecurity objective identification** — assets + objectives (needs **Phase A**)
  - **6.5.3 Threat identification** — threats grouped by target
  - **6.5.4 Risk estimation** — the risk register (likelihood/impact → risk)
  - **6.5.5 Risk evaluation** — evaluation status + an open-risks summary
- **6.6 Risk treatment** — mitigations/controls + the compliance-evidence coverage
- **6.6 Risk communication** / **6.7 Risk monitoring and review** — review cadence
  and open-risk tracking; driven by a small `monitoring`/`review` field on the
  system or risk-policy, or a documented placeholder when absent

**Design**
- Restructure `templates/threat-model.tmpl` with clause-numbered headings; keep the
  existing sub-content (the `### target` threat grouping, register table,
  mitigation tables) so behaviour is preserved.
- Offer it as the default report structure, or as a selectable `en-40000` report
  type alongside `threat-model` (sectrack already supports multiple report types
  and `template export`); recommend making it the default since Phase 3 already
  moved the report in this direction.
- Add an optional `monitoring`/`review` field (system- or policy-level) to feed §6.7.

**Caveat**: existing report tests assert sub-strings (`### comp-a`, `| critical |`,
mitigation rows) — preserve those or update the tests alongside the restructure.

**Slices**: (1) reorder + clause-number headings (keep sub-content, update only
header-asserting tests); (2) §6.5.2 objectives section (after Phase A);
(3) §6.7 monitoring/review field + section.
**Files**: threatmodel.go + template, types.go (monitoring field), MODEL.md, example.

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
  reference material, so the computed numbers are auditable.

---

## Recommended order

1. ~~**B — Attack-potential band**~~ — **done** (`f35b5aa`); `Basis`/band note deferred.
2. ~~**A — Objectives**~~ — **done**; closes the CRA §6.5.2 gap and unblocks D's §6.5.2.
3. **D — Full EN 40000 report format** (turns the output into a clause-mapped
   conformance artifact; the strongest CRA-presentation win).
4. **C1–C6** as specific needs arise — none are required for CRA conformance on
   their own; they round out ETSI fidelity.

Each phase ships TDD (one RED→GREEN per slice), keeps existing models working
(new fields optional), and updates MODEL.md + GLOSSARY.md + the example.
