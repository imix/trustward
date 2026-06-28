# Architecture Decision Records

Why the model is shaped the way it is. Each ADR records *that* a decision was made and *why* — especially the deliberate "no"s a future reader would otherwise try to "fix".

| # | Decision |
|---|---|
| [0001](0001-threat-catalog-ref-vs-backed-by.md) | Threat–catalog links: `ref` for inheritance, `backed-by` for citation |
| [0002](0002-scoring-method-named-by-method.md) | Scoring methods named by the method (`attack-potential`), not a standard |
| [0003](0003-no-control-status.md) | Controls carry no implemented/planned status — absence is the gap |
| [0004](0004-references-single-versioned-list.md) | External documents as one versioned `references:` list |
| [0005](0005-device-properties-model-local.md) | Device-property vocabulary kept model-local until reuse is proven |
| [0006](0006-risk-analysis-precedes-controls.md) | Risk analysis precedes controls |
| [0007](0007-no-yaml-schema-validation-yet.md) | No YAML schema validation yet (deferred) |
| [0008](0008-rendering-delegated-to-quarto.md) | Rendering delegated to Quarto; the binary emits `.qmd`, not HTML |
| [0009](0009-risk-layer-is-opt-in.md) | The risk layer is opt-in; `severity` is the fallback |
| [0010](0010-threats-key-sequence-only.md) | `threats:` is a threat list only when it is a YAML sequence |
| [0011](0011-external-standard-refs-unvalidated.md) | External standard references are deliberately unvalidated |

Format: `docs/adr/NNNN-slug.md`, sequential. See the project's ADR conventions for the (minimal) template.
