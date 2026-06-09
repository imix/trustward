## YAML Schema Reference

### Top-level keys

Any `.yaml` file in a sectrack project can contain any combination of the following top-level keys. All files are linked via `imports:`. The loader starts at `system.yaml` and follows the import graph depth-first.

#### `version:`
- `semver` — SemVer string, e.g. `"0.1.0"` (string)
- `releasedate` — ISO date, e.g. `"2026-06-09"` (string)

File-level metadata only; not part of the domain model.

#### `imports:` — list of file references
List of objects with:
- `path` — relative path to a YAML file (string)
- `version` — expected SemVer of the imported file (string)

#### `system:` — system metadata (first occurrence wins)
- `id` — unique identifier (string, kebab-case)
- `title` — human-readable name (string)
- `description` — free-text description used in reports (string)

#### `assets:` — list of assets
List of objects with:
- `id` — unique identifier (string, kebab-case)
- `type` — e.g. `user-data`, `config`, `firmware`, `function` (string)
- `classification` — e.g. `public`, `internal`, `confidential`, `restricted` (string, optional)
- `description` — asset purpose and sensitivity context (string)

#### `components:` — list of system components
List of objects with:
- `id` — unique identifier (string, kebab-case)
- `type` — e.g. `server`, `embedded-device`, `hmi`, `plc` (string)
- `assets` — list of asset IDs hosted on this component (list of strings)
- `description` — component role and technical details (string)

#### `trust-zones:` — logical security boundaries
List of objects with:
- `id` — unique identifier (string, kebab-case)
- `title` — human-readable name shown in diagrams (string)
- `description` — zone characteristics and access model (string)
- `members` — list of component IDs in this zone (list of strings)

#### `data-flows:` — communication paths between components
List of objects with:
- `id` — unique identifier (string, kebab-case)
- `title` — edge label in diagrams (string)
- `connects` — exactly two component IDs being connected (list of two strings)
- `assets` — list of asset IDs carried by this flow (list of strings)
- `description` — protocol, encryption, or technology details (string)

#### `threat-catalog:` — threat pattern catalog (one per file)
A single object defining reusable threat patterns. Threat entries reference patterns via `ref:` and inherit their fields.
- `id` — unique identifier (string, kebab-case)
- `title` — human-readable catalog name (string)
- `patterns` — list of threat pattern objects:
  - `id` — unique identifier within this catalog (string, kebab-case)
  - `title` — threat pattern name (string)
  - `type` — e.g. `spoofing`, `tampering`, `repudiation`, `disclosure`, `denial`, `elevation` (string)
  - `severity` — default severity level (string)
  - `notes` — description of the attack and generic mitigation guidance (string)

#### `threats:` — list of threats
Only treated as threat list when value is a YAML sequence (not a mapping). List of objects with:
- `id` — unique identifier (string, kebab-case)
- `ref` — optional reference to a threat catalog pattern in `catalog-id::pattern-id` form; inherited fields (`title`, `type`, `severity`, `notes`) are used when the instance field is empty (string, optional)
- `title` — threat name; overrides catalog if set (string)
- `type` — e.g. `spoofing`, `tampering`, `repudiation`, `disclosure`, `denial`, `elevation`; overrides catalog if set (string)
- `target` — component ID or data-flow ID being attacked (string)
- `asset` — asset ID at risk (string, optional)
- `severity` — e.g. `low`, `medium`, `high`, `critical` (string)
- `mitigations` — list of control IDs that reduce risk (list of strings)
- `residualRisk` — severity after mitigations applied (string)
- `notes` — rationale, mitigation justification, residual risk explanation (string)

#### `catalog:` — requirement catalog (one per file)
A single object defining a named set of requirements used for gap analysis and compliance mapping:
- `id` — unique identifier (string, kebab-case)
- `title` — human-readable catalog name (string)
- `requirements` — list of requirement objects:
  - `id` — unique identifier within this catalog (string, kebab-case)
  - `title` — requirement name (string)
  - `description` — what must be implemented (string)
  - `satisfies` — list of requirements in other catalogs this requirement covers, in `catalog-id::req-id` form (list of strings, optional)

Multiple catalogs are loaded by importing multiple catalog files. A company baseline catalog can reference which IEC 62443, NIS2, or other standard requirements it satisfies via `satisfies:`.

#### `controls:` — list of security controls
List of objects with:
- `id` — unique identifier (string, kebab-case)
- `title` — control name (string)
- `description` — control scope and implementation approach (string)
- `ref` — single catalog requirement this control implements, in `catalog-id::req-id` form (string, optional)
- `evidence` — list of references proving implementation: commit hashes, ticket numbers, document names (list of strings, optional)

---

### Cross-reference rules

| Source | Field | Target |
|--------|-------|--------|
| `components[].assets[]` | asset IDs | `assets[].id` |
| `trust-zones[].members[]` | component IDs | `components[].id` |
| `data-flows[].connects[]` | component IDs | `components[].id` |
| `data-flows[].assets[]` | asset IDs | `assets[].id` |
| `threats[].target` | component or data-flow ID | `components[].id` or `data-flows[].id` |
| `threats[].asset` | asset ID | `assets[].id` |
| `threats[].mitigations[]` | control IDs | `controls[].id` |
| `controls[].ref` | `catalog-id::req-id` | `catalog.id` + `catalog.requirements[].id` |
| `threats[].ref` | `catalog-id::pattern-id` | `threat-catalog.id` + `threat-catalog.patterns[].id` |
| `catalog.requirements[].satisfies[]` | `catalog-id::req-id` | another `catalog.id` + `requirements[].id` |

---

### File splitting and merging

The loader merges content from all imported files depth-first. Behavior by key:
- `system:` — first occurrence wins; subsequent declarations ignored
- `version:`, `imports:` — file-level metadata only
- All list fields (`assets:`, `components:`, `trust-zones:`, `data-flows:`, `threats:`, `controls:`) — merged by appending
- `catalog:` — each file contributes at most one catalog; all catalogs are collected into the project
- `threat-catalog:` — same as above; threat refs are resolved after the full graph is loaded

A single file can hold the entire model; splitting is purely for version management convenience.

---

### ID conventions

- All IDs use kebab-case: `comp-iam`, `zone-control-room`, `threat-steal-data`
- IDs must be unique within their entity type across all imported files
- Use descriptive prefixes to clarify intent when appropriate

---

### Minimal complete example

**system.yaml**
```yaml
version:
  semver: "0.1.0"
  releasedate: "2026-06-09"

system:
  id: fire-protection
  title: Fire Protection System
  description: Building fire detection and suppression system

assets:
  - id: sensor-readings
    type: telemetry
    classification: internal
    description: Real-time temperature and smoke sensor data
  - id: control-config
    type: config
    classification: restricted
    description: Suppression system activation logic

components:
  - id: sensor-hub
    type: embedded-device
    assets: [sensor-readings]
    description: Central sensor aggregator
  - id: control-unit
    type: plc
    assets: [control-config]
    description: Automated suppression logic controller

trust-zones:
  - id: zone-industrial
    title: Industrial Floor
    description: Manufacturing area with fire hazard
    members: [sensor-hub]
  - id: zone-control
    title: Control Room
    description: Operator station and decision center
    members: [control-unit]

data-flows:
  - id: flow-sensor-control
    title: Sensor Alerts
    connects: [sensor-hub, control-unit]
    assets: [sensor-readings]
    description: Encrypted sensor data over hardened network
```

**threat-model.yaml**
```yaml
version:
  semver: "0.1.0"
  releasedate: "2026-06-09"

imports:
  - path: system.yaml
    version: "0.1.0"

controls:
  - id: comp-sensor-encryption
    title: Sensor Data Encryption
    description: AES-256 encryption in transit
  - id: comp-access-control
    title: Role-Based Access Control
    description: Operator authentication and privilege separation

threats:
  - id: threat-sensor-spoof
    title: Sensor Data Spoofing
    type: spoofing
    target: flow-sensor-control
    asset: sensor-readings
    severity: critical
    mitigations: [comp-sensor-encryption]
    residualRisk: medium
    notes: Encrypted channel prevents active injection; monitoring logs for anomalies reduce residual risk to medium

  - id: threat-config-tampering
    title: Control Logic Tampering
    type: tampering
    target: control-unit
    asset: control-config
    severity: critical
    mitigations: [comp-access-control]
    residualRisk: low
    notes: Role-based access and audit logging ensure only authorized operators modify suppression logic
```

---

### Notes

- The `threats:` key disambiguation: a YAML sequence (list) is treated as a threat list; a YAML mapping (dict) is treated as vocabulary and ignored by the loader.
- `version:` and `imports:` are file-level metadata; they do not appear in the merged domain model.
- All ID references must resolve within the import graph; unresolved references are validation errors.
