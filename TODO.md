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

Remaining:
- **CRA report shaping** — structure the report to prEN 40000-1-2 §6.2–6.7 and
  fold in the control→requirement coverage as compliance evidence.
- **Register polish** — the risk register's Likelihood column is blank for
  etsi-tvra threats (likelihood is computed); show the attack-potential band there.

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
