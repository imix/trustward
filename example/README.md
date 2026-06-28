# Examples

Worked trustward models, smallest first. Each model directory has its own README; run any of them with the [`trustward.sh`](../trustward.sh) wrapper from inside the directory (e.g. `cd example/fire-protection-system && ../../trustward.sh render`).

| Directory | What it shows |
|---|---|
| [fire-protection-system](fire-protection-system) | A complete baseline model: imports, a company control catalog + a STRIDE threat catalog, threats, controls, and compliance mapping. Start here. |
| [access-control](access-control) | The advanced model: the opt-in risk layer (attack-potential scoring, treatment, the CRA gate, risk matrix), an EMB3D device pass via property-gating, external `references`, multiple device profiles, and a local report-template override. |
| [report-templates](report-templates) | Standard-shaped report shells (not a model) — copy one to `report.tmpl` to get a conformance report's full structure with the manual narrative sections stubbed as `_fill in_`. |
