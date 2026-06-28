# External documents are one versioned `references:` list

External versioned artifacts a model depends on — variant register, requirements spec, standard, SBOM — are modelled as a single `references:` list of `{id, title, version, location}`, not a field per kind. Each pins a version that the report renders *from the model*, so a cited version can't drift from a hand-typed copy; local `location`s must resolve at load, URLs are recorded but not fetched.

## Considered Options

- **A bespoke `variants:` field** (and one more per future document kind) — rejected once several such documents existed; it is the wrong shape for a recurring need.
- **Version pinned as prose in the template** — rejected: a template can't validate itself, so the cited version silently drifts. Rendering from model data removes the second copy.
