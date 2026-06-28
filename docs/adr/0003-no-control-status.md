# Controls carry no implemented/planned status

Controls have no lifecycle field. "Not done" is already represented by **absence** — a requirement with no control is a `gap`, a threat with no mitigation keeps its severity. A `status: planned` flag would let a control claim coverage or mitigation it hasn't earned; the model represents what *is*, not what's intended — plans belong in the issue tracker.

A future reader will be tempted to add this field (we were). The treatment flow does not need it: the *decision* to mitigate is recorded on the threat (`treatment` + `owner`, which the CRA gate checks); the *control* enters `controls:`/`mitigations:` only once it exists, at which point the gap closes and residual risk can drop.
