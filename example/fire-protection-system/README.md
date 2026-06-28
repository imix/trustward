# Fire Protection System

A complete baseline trustward model — the place to start. A standalone fire
protection system for a single building: a smoke/heat sensor detects fire, a
central unit evaluates the signal, and it notifies occupants and the fire
service.

## What it demonstrates

- **`imports`** — the model is split across files merged from `system.yaml`:
  - `company-catalog.yaml` — a reusable company control catalog (requirements).
  - `stride-catalog.yaml` — a threat catalog of STRIDE patterns.
  - `company.yaml` — shared company-level definitions.
  - `threat-model.yaml` — the threats and controls for this system.
- **Catalogs** — threats inherit from catalog patterns via `ref`, and controls
  map onto catalog requirements, producing the compliance-evidence section.
- **Controls** — recorded as current-state measures (no implemented/planned
  status; absence is the gap).

It deliberately does **not** turn on the risk layer — it's the lightweight
posture. For the gated, scored posture see [../access-control](../access-control).

## Run it

```
cd example/fire-protection-system
../../trustward.sh validate      # referential integrity → "model is consistent"
../../trustward.sh diagram dataflow
../../trustward.sh render        # → out/report.html
```
