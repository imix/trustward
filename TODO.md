# TODO

## Risk management

A threat describes what could go wrong. A risk is a formal assessment of
likelihood × impact for a specific threat in a specific context, leading to
a treatment decision (mitigate, accept, transfer, avoid) with a named owner
and sign-off date. This needs design before implementation.

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
