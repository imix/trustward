# Glossary

Domain terms used in trustward models and documentation.

---

## Asset

Data, configuration, firmware, or capability that has value and could be targeted by an attacker. Assets are defined at the project level and assigned to components or data flows. Examples: sensor readings, system configuration, firmware image, alarm activation signal.

---

## Attack Potential

A measure of how hard an attack is, per ETSI TS 102 165-1 (TVRA) clause 6.6.3: the sum of four attacker factors — expertise, knowledge, opportunity, and equipment. Used by the `etsi-tvra` scoring method, where a higher attack potential maps to a *lower* [Likelihood](#likelihood) (a harder attack is less likely).

---

## Component

A physical or logical node in the system — a device, server, PLC, HMI, or application. Components host assets and are placed into trust zones. Data flows connect pairs of components.

---

## Control

A technical or organizational measure that reduces the likelihood or impact of a threat. Controls are defined once and referenced by ID from threat mitigations. A control ID that appears in `mitigations:` must resolve to a `controls:` entry somewhere in the import graph.

---

## Cybersecurity Objective

A property an [Asset](#asset) must uphold, named on a CIA scale
(`confidentiality`, `integrity`, `availability`, `authenticity`,
`accountability`). Objectives are defined at the project level and referenced
from assets (`objectives:`) and from threats that breach them (`violates:`).
They make the objective→asset→threat trace explicit, per prEN 40000-1-2 §6.5.2
("Asset and cybersecurity objective identification").

---

## Data Flow

A communication path between exactly two components. Data flows carry assets across the network. When a data flow connects components in different trust zones it crosses a trust boundary — a signal that additional scrutiny may be warranted.

---

## Import Graph

The set of YAML files reachable from `system.yaml` by following `imports:` declarations recursively. The loader merges all files in the graph into a single project model. Cycles are detected and skipped.

---

## Impact

The magnitude of harm if a threat is realised, rated qualitatively (`low`, `medium`, `high`). Combined with [Likelihood](#likelihood) to compute the risk level.

---

## Likelihood

How probable it is that a threat is realised, rated qualitatively (`low`, `medium`, `high`). Combined with [Impact](#impact) to compute the risk level.

---

## Project

The accumulated security model for a directory: all assets, components, trust zones, data flows, threats, and controls merged from the import graph. The unit of work for all `trustward` commands.

---

## Risk

The combination of [Likelihood](#likelihood) and [Impact](#impact) for a threat in its context. The risk *level* is computed by the [Risk Policy](#risk-policy)'s method (the default qualitative method is a 3×3 matrix → `low`/`medium`/`high`/`critical`). A risk above the acceptance criteria must carry a [Treatment](#treatment) decision and an owner.

---

## Risk Policy

The project-level declaration of how risk is scored (`method`) and which risk levels are acceptable without treatment (`accept`). Corresponds to the risk acceptance criteria of CRA / prEN 40000-1-2 §6.3. When present, it activates the CRA gate in `validate`: every non-accepted risk needs a treatment and an owner.

---

## Residual Risk

The severity of a threat that remains after all listed mitigations are applied. If no controls address a threat, residual risk equals the original severity. Residual risk is an assessment, not a computed value.

---

## Severity

A qualitative rating of threat impact or likelihood. Recommended values: `low`, `medium`, `high`, `critical`. Used on both the raw threat and the residual risk after mitigations.

---

## Target

The component or data flow a threat is directed at. A threat against a component attacks the assets it hosts; a threat against a data flow attacks the assets it carries.

---

## Threat

A potential attack scenario: what could go wrong, against which target, with what severity, and what controls reduce the risk. Threats are typed (spoofing, tampering, repudiation, disclosure, denial, elevation) and have an explicit residual risk assessment.

---

## Treatment

The decision on how to handle a risk: `mitigate` (apply controls), `accept` (tolerate it), `transfer` (e.g. insure or outsource), or `avoid` (remove the feature or context). Recorded with an owner and a sign-off date. Required by the CRA gate for any risk above the acceptance criteria.

---

## Trust Zone

A logical boundary grouping components that share a common security posture and access model. Represented as a subgraph in the data flow diagram. A data flow that connects components in different trust zones crosses a trust boundary.
