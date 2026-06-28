# A mitigate treatment needs a control to count as treated

The CRA gate closes a non-accepted risk when it carries a `treatment` and an `owner`. A risk-process review found this passed on recorded *intent* rather than *reduction*: most `mitigate` decisions had no control, residual equalled the inherent level, yet the model validated clean — the gate proved a decision was written down, not that risk went down.

So a `treatment: mitigate` now counts as treated only with at least one `mitigations:` control; `accept` / `transfer` / `avoid` still need only an owner, because they are not reductions. A mitigate with no control is an intention to fix — a plan — and the model is current-state, not plans ([0003](0003-no-control-status.md)), so it stays **open** and fails the gate. To acknowledge a risk you are not currently reducing, use `accept` with an owner.

The same principle guards `residualRisk`: a residual below the computed level must be earned by a control, or it is an unsupported paper reduction, so `validate` flags it. This sharpens, not contradicts, [0006](0006-risk-analysis-precedes-controls.md) — controls still follow risk; this only stops a treatment from claiming credit a control hasn't delivered.
