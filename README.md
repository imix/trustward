# sectrack — IEC 62443 threat modelling and controls model

A YAML-based model for conducting threat analyses and managing security controls
against the **IEC 62443** (ANSI/ISA-62443) standard for industrial cybersecurity.

All files are human-readable YAML. Tooling (validators, gap reporters, dashboards)
is not yet present but this model is designed to be straightforward to automate —
the worked example at the end of this file shows the exact manual computation a
tool would run.

---

## Repository layout

```
sectrack/
  README.md               ← you are here
  MODEL.md                ← complete schema: every entity, field, and cross-reference
  GLOSSARY.md             ← IEC 62443 terms (CR, CCSC, FR, SL-T/C/A, zone, conduit …)

  definitions/
    company.yaml          ← (1) vocabulary: asset types & classifications,
                               threat taxonomy & severity, company-owned controls

  catalogs/
    iec-62443-4-2.yaml    ← (2) standard catalog: controls and control-groups
                               for ANSI/ISA-62443-4-2-2018

  mappings/
    iec-62443-4-2.yaml    ← (3) capability evidence: which company controls
                               implement which standard requirements (SL-C)

  systems/
    example/
      system.yaml         ← (4, 5) the system under assessment: assets, components,
                               zones (with SL-T), and conduits
      threat-model.yaml   ← (6) threats, severity, mitigations, residual risk
```

Files reference each other via `imports`. Every cross-reference is a string ID
that must resolve to a defined entity in the same file or an imported one — see
MODEL.md for the full map.

---

## Methodology — the seven-step loop

These seven steps correspond directly to the files above. Iterate the loop as
the system evolves or new requirements are adopted.

```
 ┌──────────────────────────────────────────────────────────────────────┐
 │ 1  DEFINE VOCABULARY        definitions/company.yaml                 │
 │    Asset types, classifications, threat taxonomy, severity scale,    │
 │    and company-owned controls. The shared dictionary everything       │
 │    else references.                                                   │
 │                                                                       │
 │ 2  ADOPT STANDARD CATALOG   catalogs/iec-62443-4-2.yaml              │
 │    Transcribe the published standard: flat per-permutation controls   │
 │    and control-groups keyed by FR and SL. Read-only — never add       │
 │    company content here.                                              │
 │                                                                       │
 │ 3  MAP CAPABILITY (SL-C)    mappings/iec-62443-4-2.yaml              │
 │    Record which company control implements which standard             │
 │    requirement. This is your SL-C (capability) evidence. A           │
 │    requirement with no mapping is a visible, trackable gap.           │
 │                                                                       │
 │ 4  MODEL THE SYSTEM         systems/*/system.yaml                    │
 │    Instantiate assets and components, then partition them into        │
 │    zones and conduits.                                                │
 │                                                                       │
 │ 5  SET TARGETS (SL-T)       systems/*/system.yaml — zones[].targetSL │
 │    Each zone declares the Security Level it must achieve. The         │
 │    required control set is derived — not retyped — by looking up      │
 │    the catalog control-group for that FR/SL combination.              │
 │                                                                       │
 │ 6  ANALYSE THREATS          systems/*/threat-model.yaml              │
 │    For each threat: which asset/zone, what type, what severity,       │
 │    which company controls mitigate it, and what residual risk         │
 │    remains after those controls are applied.                          │
 │                                                                       │
 │ 7  COMPUTE COVERAGE (SL-A)  derived from steps 3 + 5                 │
 │    Required controls (SL-T → control-group) vs. implemented           │
 │    (SL-C mapping) = gap list. Highest SL with no gaps = SL-A.        │
 └──────────────────────────────────────────────────────────────────────┘
```

---

## Worked coverage example

This shows step 7 manually, using the example system. A future tool would
automate exactly this computation.

**Zone:** `zone-control-room` — `targetSL: 1`

**Step 1 — look up the required control-group for FR1 / SL1:**

From `catalogs/iec-62443-4-2.yaml`, group `IEC-62443-4-2-SL1-IAC`:

| Requirement | Title |
|-------------|-------|
| CR 1.1  | Human user identification and authentication |
| CR 1.3  | Account management |
| CR 1.4  | Identifier management |
| CR 1.5  | Authenticator management |
| CR 1.7  | Strength of password-based authentication |
| CR 1.10 | Authenticator feedback |
| CR 1.11 | Unsuccessful login attempts |
| CR 1.12 | System use notification |

**Step 2 — look up SL-C: which requirements have a company control mapped?**

From `mappings/iec-62443-4-2.yaml`:

| Requirement | Implemented by | Status |
|-------------|----------------|--------|
| CR 1.1  | `comp-iam`             | ✅ covered |
| CR 1.3  | `comp-iam`             | ✅ covered |
| CR 1.4  | `comp-iam`             | ✅ covered |
| CR 1.5  | `comp-password-policy` | ✅ covered |
| CR 1.7  | `comp-password-policy` | ✅ covered |
| CR 1.10 | —                      | ❌ **gap** |
| CR 1.11 | `comp-password-policy` | ✅ covered |
| CR 1.12 | —                      | ❌ **gap** |

**Step 3 — determine SL-A:**

Two open gaps (CR 1.10 and CR 1.12) mean SL1 is not yet fully achieved for FR1.

```
SL-T = 1    (target)
SL-C = 1    (comp-iam + comp-password-policy have the right capabilities)
SL-A = 0    (SL1 is NOT achieved — two requirements unimplemented)
```

**Action:** Add company controls for authenticator feedback (CR 1.10) and
system use notification (CR 1.12), add their `implements` entries in
`mappings/iec-62443-4-2.yaml`, and re-run. Once both are covered, SL-A reaches 1.

---

## Extending this model

| To add… | Edit… |
|---------|-------|
| A new asset type | `definitions/company.yaml` → `asset.types` |
| A new company control | `definitions/company.yaml` → `controls` |
| Coverage for another standard requirement | `mappings/iec-62443-4-2.yaml` → `implements` |
| More of the IEC 62443-4-2 catalog (FR2–FR7) | `catalogs/iec-62443-4-2.yaml` → add controls + control-groups |
| A new system/product | New folder under `systems/`, copy `example/` as a starting point |
| A second standard (NIST, ISO 27001, …) | New file under `catalogs/`, new mapping file under `mappings/` |
