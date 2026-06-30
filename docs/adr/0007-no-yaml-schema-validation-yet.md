# YAML schema validation: structural at load, referential in Go

This decision **supersedes** the earlier one recorded here ("no YAML schema validation yet — deferred until the file structures stabilise"). The structures have stabilised, and CI needs a single structural gate, so `trustward validate` now validates every file in the import graph against [`schema/trustward.schema.json`](../../schema/trustward.schema.json) — the *same* schema editors consume via yaml-language-server — before the referential checks run.

## Division of labour

- **Schema (structural, per-file).** The shape of each top-level key and the closed vocabularies the scoring code keys on: `severity`, `treatment`, objective `type`, the `attack` factors, `risk-policy.method`, residual/accept levels, data-flow arity. One source, shared with the editor, so a vocabulary can't drift between the editor hint and the CI gate.
- **`internal/validate` (relational, whole-model).** Cross-references resolve, IDs are unique within a kind, and the CRA gate (treatment + owner + the mitigate-needs-a-control logic) — things a per-file JSON Schema cannot express.

## Constraints kept

- The schema stays deliberately **permissive** — unknown keys allowed (the loader ignores them; models carry local extensions like component `properties:`), IDs unconstrained (catalog IDs mirror external taxonomies: `TID-119`, `SR-1.1`). It only *adds* checks; it never rejects what the loader accepts.
- Validation runs in `validate` (the CI gate). `render`/`report` stay tolerant — a malformed model fails them at Decode, as before.
- The schema is **embedded** in the binary (`//go:embed`), so the bare `trustward` binary validates without the repo present.

## One owner per check

Structure and vocabulary live **only** in the schema; relations and the gate live **only** in `internal/validate`. The vocabulary checks that once existed in both (`treatment`, `method`, the scales, attack factors, residual level, data-flow arity, reference version) were removed from `internal/validate` — so a malformed model reports each mistake once, not twice, and a vocabulary has a single definition. Their coverage moved to `schema`'s test suite.

`internal/validate` no longer enforces field vocabularies; the schema pass in `validate` is what catches them. The two run together in the `validate` command, so the CI gate is unchanged.

Status: accepted (supersedes the prior deferral).
