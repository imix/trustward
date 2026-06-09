# Glossary

Domain terms used in sectrack models and documentation.

---

## Asset

Data, configuration, firmware, or capability that has value and could be targeted by an attacker. Assets are defined at the project level and assigned to components or data flows. Examples: sensor readings, system configuration, firmware image, alarm activation signal.

---

## Component

A physical or logical node in the system — a device, server, PLC, HMI, or application. Components host assets and are placed into trust zones. Data flows connect pairs of components.

---

## Control

A technical or organizational measure that reduces the likelihood or impact of a threat. Controls are defined once and referenced by ID from threat mitigations. A control ID that appears in `mitigations:` must resolve to a `controls:` entry somewhere in the import graph.

---

## Data Flow

A communication path between exactly two components. Data flows carry assets across the network. When a data flow connects components in different trust zones it crosses a trust boundary — a signal that additional scrutiny may be warranted.

---

## Import Graph

The set of YAML files reachable from `system.yaml` by following `imports:` declarations recursively. The loader merges all files in the graph into a single project model. Cycles are detected and skipped.

---

## Project

The accumulated security model for a directory: all assets, components, trust zones, data flows, threats, and controls merged from the import graph. The unit of work for all `sectrack` commands.

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

## Trust Zone

A logical boundary grouping components that share a common security posture and access model. Represented as a subgraph in the data flow diagram. A data flow that connects components in different trust zones crosses a trust boundary.
