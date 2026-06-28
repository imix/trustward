# External standard references are deliberately unvalidated

A requirement's `satisfies:` entries — and any reference to an external standard's IDs (e.g. `iec-62443-sl2::SR-1.1`) — are **not** checked by `validate`. They point at standards that need not be loaded into the model, so an unresolved target is expected, not an error. Cross-references to entities *within* the model stay strictly validated.

This lets a baseline catalog map onto standards you don't carry, without forcing you to model an entire external standard just to reference it. A reader who sees an unchecked reference and assumes it's a validation bug should read it as the deliberate inside-vs-outside boundary.
