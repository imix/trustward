# Access Control System

The advanced trustward model. Physical door access control for a commercial
office building: a reader validates a credential at the door, a controller
releases or holds the door and logs the attempt, and a cloud platform manages
rights and audit. It spans dumb strikes, connected locksets, and self-contained
battery locksets.

## What it demonstrates (beyond the baseline)

- **The opt-in risk layer** — a `risk-policy` (method `attack-potential`,
  `accept: [low]`) turns the threat list into a gated assessment:
  - **attack-potential scoring** — each threat carries attacker factors
    (expertise/knowledge/opportunity/equipment) that derive a likelihood, ×
    impact → a risk level. Used where probability data doesn't exist but
    attacker difficulty is design-answerable.
  - **the CRA gate** — every risk above the acceptance criteria must be treated
    (a `mitigate` with a real control, or an owned `accept`); `validate` fails
    open risks. See ADR [0012](../../docs/adr/0012-mitigate-needs-a-control.md).
  - **the risk matrix** — a likelihood×impact heatmap in the rendered report.
- **An EMB3D device pass via property-gating** — component `properties:` decide
  which device threats apply; threats cite EMB3D TIDs with `backed-by` (citation,
  not inheritance) instead of importing the whole catalog.
- **External `references`** — the assessment pins the version of an external
  document (`variants.md`, the variant register mapping product SKUs to the
  model's profiles) and the report cites it from data, so it can't drift.
- **Multiple device profiles** — `lock`, `lock-connected`, and
  `lock-battery-standalone` show how one model covers a product family.
- **A local report template** — `report.tmpl` overrides the built-in default
  (derived from [../report-templates/en40000.tmpl](../report-templates)).

## Run it

```
cd example/access-control
../../trustward.sh validate
../../trustward.sh render                 # → out/report.html (rendered report + matrix)
../../trustward.sh report --format json   # machine-readable risk register
```
