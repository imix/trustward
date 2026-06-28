# The risk layer is opt-in; `severity` is the fallback

A model without a `risk-policy` has no computed risk level and no CRA gate: `validate` checks only referential integrity, and untreated high-severity threats are fine. Declaring a `risk-policy` turns the threat list into a *gated* assessment — every non-accepted risk must carry a `treatment` and `owner`. When a scored threat's inputs are absent or invalid, the computed level falls back to the threat's `severity`.

This keeps lightweight sketches simple and pre-risk models working, while letting a conformance model opt into enforcement. The cost is that "does `validate` gate?" depends on whether a `risk-policy` is present — surprising until you know the layer is opt-in.
