# Data Model

This document is the single source of truth for every field in the YAML files.
A field used in any YAML file must appear here; a field defined here must appear
in at least one YAML file (round-trip). See GLOSSARY.md for term definitions.

---

## Common conventions

### Versioning
Every file that can be imported carries a top-level `version` block:

```yaml
version:
  semver: <SemVer string>      # e.g. "1.0.0"
  releasedate: <ISO date>      # e.g. "2026-06-09"
```

### Imports
Files that reference entities from other files declare:

```yaml
imports:
  - path: <relative path>       # e.g. "../../definitions/company.yaml"
    version: <SemVer string>    # version of the imported file expected
```

Cross-references use **string IDs only**. Every ID referenced in any `members`,
`by`, `affects`, `mitigations`, `asset`, `zone`, or `connects` field **must
resolve** to a defined entity either in the same file or in an imported file.

### ID conventions
- **Company-defined entities** — kebab-case, optionally prefixed by kind:
  `comp-iam`, `zone-control-room`, `threat-steal-data`.
- **Standard catalog entries** — preserve the standard's own notation:
  `CR 1.1`, `CR 1.1 (1) (2)`, `CCSC 1`, `EDR 2.4`.

### Cardinality rule
Any field that references multiple entities is **always a list**, even when a
single value is currently sufficient. Never a scalar for multi-valued fields.

---

## Entity types

### `asset` (in `definitions/company.yaml`)

```yaml
asset:
  types:
    - name: <string>            # unique ID; referenced by assets[].type
      description: <string>
  classifications:
    - name: <string>            # unique ID; referenced by assets[].classification
      level: <integer>          # 0 = least sensitive; higher = more sensitive
      description: <string>
```

### `asset instance` (in `system.yaml`)

```yaml
assets:
  - id: <string>
    type: <string>              # → asset.types[].name
    classification: <string>    # (optional) → asset.classifications[].name
    description: <string>
```

---

### `component` (in `system.yaml`)

```yaml
components:
  - id: <string>
    type: <string>              # free text: server, embedded-device, hmi, switch, …
    description: <string>
```

---

### `zone` (in `system.yaml`)

```yaml
zones:
  - id: <string>
    title: <string>
    targetSL: <integer>         # 1–4; determines required control-group from catalog
    members:                    # list of component IDs in this zone
      - <component id>
```

### `conduit` (in `system.yaml`)

```yaml
conduits:
  - id: <string>
    title: <string>
    connects:                   # exactly two zone IDs
      - <zone id>
      - <zone id>
    description: <string>       # (optional) protocol / technology note
```

---

### `control` — company-defined (in `definitions/company.yaml`)

```yaml
controls:
  - id: <string>                # kebab-case; comp- prefix recommended
    title: <string>
    description: <string>
```

### `control` — standard catalog entry (in `catalogs/*.yaml`)

```yaml
controls:
  - id: <string>                # standard notation: "CR 1.1", "CR 1.1 (1)", etc.
    type: <string>              # ccsc | cr | edr | hdr | ndr | sar
    title: <string>
    clause: <string>            # section reference in the standard (e.g. "5.3")
    appliesTo: <string>         # (optional) component type, e.g. embedded-device
    _stub: true                 # (optional) marks entries added only for reference
                                #   integrity; full text not yet transcribed
```

---

### `control-group` (in `catalogs/*.yaml`)

```yaml
control-groups:
  - id: <string>
    title: <string>
    foundationalRequirement: <string>   # (optional) FR1–FR7
    securityLevel: <integer>            # (optional) 1–4
    members:                            # (optional) flat list of control IDs
      - <control id>
    containedGroups:                    # (optional) IDs of nested control-groups
      - <control-group id>
```

Either `members` or `containedGroups` (or both) must be present.

---

### `implements` mapping (in `mappings/*.yaml`)

```yaml
implements:
  - requirement: <string>     # → catalogs control id
    by:                       # list of company control IDs that satisfy this requirement
      - <control id>
```

An entry with an empty or missing `by` list, or a standard requirement that
appears in a zone's SL-T control-group but has **no** matching `implements`
entry, is an **open capability gap**.

---

### `threat` (in `systems/*/threat-model.yaml`)

```yaml
threats:
  - id: <string>
    title: <string>
    type: <string>              # → threats.types[].id in definitions/company.yaml
    asset: <string>             # (optional) → assets[].id
    zone: <string>              # → zones[].id (the attack surface)
    affects:                    # list of component IDs in the attack path
      - <component id>
    severity: <string>          # → threats.severity[].name
    mitigations:                # list of company control IDs that reduce risk
      - <control id>
    residualRisk: <string>      # → threats.severity[].name after mitigations
    notes: <string>             # (optional) rationale — why these mitigations,
                                #   what gap remains, and SL alignment
```

---

## Cross-reference map

```
system.yaml
  assets[].type             → definitions/company.yaml  asset.types[].name
  assets[].classification   → definitions/company.yaml  asset.classifications[].name
  zones[].members[]         → system.yaml               components[].id
  conduits[].connects[]     → system.yaml               zones[].id

mappings/iec-62443-4-2.yaml
  implements[].requirement  → catalogs/iec-62443-4-2.yaml  controls[].id
  implements[].by[]         → definitions/company.yaml      controls[].id

systems/*/threat-model.yaml
  threats[].type            → definitions/company.yaml  threats.types[].id
  threats[].severity        → definitions/company.yaml  threats.severity[].name
  threats[].residualRisk    → definitions/company.yaml  threats.severity[].name
  threats[].asset           → system.yaml               assets[].id
  threats[].zone            → system.yaml               zones[].id
  threats[].affects[]       → system.yaml               components[].id
  threats[].mitigations[]   → definitions/company.yaml  controls[].id
```
