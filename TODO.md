# TODO

## Risk management — done (Phase 1)

Threats now carry `likelihood`/`impact`/`treatment`/`owner`/`decided`; a
project-level `risk-policy` sets the scoring method and acceptance criteria.
The risk level is computed (`internal/risk`, qualitative 3×3 matrix) and
`validate` enforces the CRA gate: every non-accepted risk needs a treatment
and an owner. The report shows a risk register.

Phase 2 done: the `risk.Scorer` seam now has a second profile, `etsi-tvra`
(`internal/risk/etsi.go`) — attack-potential factors on a threat's `attack`
block sum to an attack potential that maps inversely to likelihood, then the
shared matrix. Selected via `risk-policy.method: etsi-tvra`.

Phase 3 done: the report is shaped to the prEN 40000-1-2 §6 process — a Risk
Acceptance Criteria and Methodology section (§6.3, from the risk-policy), a Risk
Register with an Evaluation column marking each risk accepted/treated/**open**
(§6.5.4–5), and the control→requirement coverage reframed as Compliance Evidence
(§6.6). The "open risk" rule is defined once in `risk.Evaluate` and shared by the
report and the validate CRA gate.

Phase 4 done: **Register polish** — `Scorer` now returns a `risk.Score`
(`{Level, Likelihood}`) instead of a bare level, so the derived likelihood is
no longer discarded. `risk.Eval` embeds `Score`; `risk.Evaluate` is the single
scoring entry point (the standalone `Score(p)` map and the report's redundant
`RiskLevels` are gone). The register's Likelihood column shows the derived
likelihood for etsi-tvra threats (was blank).

## Diagram scale

Data flow diagrams become unreadable on large systems (20+ components,
5+ trust zones). Proposed filtering options:

- `--zone <id>` — render only components in that trust zone plus cross-boundary flows
- `--component <id>` — one-hop neighbourhood view
- `--cross-zone-only` — drop intra-zone flows

## YAML schema validation

Validate YAML files against a schema on load to give actionable errors
instead of silent zero-values. Deferred until file structures stabilise.

## Threat template library

Reusable threat patterns (e.g. STRIDE per component type) that can be
instantiated with per-system overrides, so a new model doesn't start
from zero.
