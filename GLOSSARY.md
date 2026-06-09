# Glossary — IEC 62443 terms used in this model

Terms are drawn from **ANSI/ISA-62443-3-2-2020** (Security risk assessment) and
**ANSI/ISA-62443-4-2-2018** (Technical security requirements for IACS components).

---

## Asset
Any data, function, device, or capability that has value to the organization
and that an attacker might target.

---

## Component
A physical or virtual node (server, PLC, HMI, network switch) that hosts assets
and implements controls. Mapped to zones in `system.yaml`.

---

## Conduit
A communication path connecting two zones. Per 62443-3-2, a conduit inherits
the higher of the two connected zones' SL-T values; conduit-specific controls
(e.g. firewall, encrypted tunnel) can reduce the inherited requirement.

---

## Control (company control)
A technical or organizational measure deployed by the organization to fulfill
one or more standard requirements. Defined in `definitions/company.yaml`;
capability evidence recorded in `mappings/`.

---

## Control-group
A named, version-stable set of standard controls. Used in `catalogs/` to
express the exact set of requirements mandated at a given Security Level for
a given Foundational Requirement. Membership is always **explicit** because
the standard has "Not Selected" gaps — a control absent at SL1 may appear at
SL2, so membership cannot be inferred from level alone.

---

## CR (Component Requirement)
A generic requirement that applies to all component types (SAR, EDR, HDR, NDR).
Identified by `CR <FR>.<seq>` (e.g. `CR 1.1`). May have Requirement
Enhancements (`RE`), which are separate, higher-bar permutations written as
`CR 1.1 (1)`, `CR 1.1 (1) (2)`, etc.

---

## CCSC (Common Component Security Constraint)
An overarching constraint that applies across all component types, independent
of Security Level. Structurally identical to a CR but carries the `ccsc` type.

---

## EDR / HDR / NDR / SAR
Component-type-specific requirement series:
- **EDR** — Embedded Device Requirement
- **HDR** — Host Device Requirement
- **NDR** — Network Device Requirement
- **SAR** — Software Application Requirement

The same conceptual requirement (e.g. "Mobile code") may appear as separate
numbered entries per component type (`EDR 2.4`, `SAR 2.4`, …), each with its
own clause and wording.

---

## FR (Foundational Requirement)
A top-level security requirement category. IEC 62443-4-2 defines seven:

| FR | Name |
|-----|------|
| FR1 | Identification & Authentication Control (IAC) |
| FR2 | Use Control (UC) |
| FR3 | System Integrity (SI) |
| FR4 | Data Confidentiality (DC) |
| FR5 | Restricted Data Flow (RDF) |
| FR6 | Timely Response to Events (TRE) |
| FR7 | Resource Availability (RA) |

---

## RE (Requirement Enhancement)
An additive extension to a base CR that raises the security bar. REs are
cumulative and applied in parenthetical notation: `CR 1.1 (1)` includes RE1;
`CR 1.1 (1) (2)` includes RE1 and RE2. Each permutation is a separate, flat
control entry in the catalog.

---

## Security Level (SL)
An integer (0–4) expressing the security capability or target for a zone or
component. Three related concepts:

| Term | Meaning |
|------|---------|
| **SL-T** (Target) | The SL a zone is *required* to achieve — set by the risk assessment and declared in `system.yaml` on each zone. |
| **SL-C** (Capability) | The SL a company control or component *can achieve* — evidenced by the mapping from company controls to standard requirements in `mappings/`. |
| **SL-A** (Achieved) | The SL *actually reached* given the deployed controls. SL-A = the highest SL for which all required controls are implemented (no gaps). If any gap exists at SL-T, then SL-A < SL-T. |

---

## Zone
A logical grouping of components and assets with a **uniform security target**
(SL-T). The fundamental unit of threat analysis in IEC 62443-3-2. Defined in
`system.yaml`; threats are assigned to zones in `threat-model.yaml`.
