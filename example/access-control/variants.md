# Variant register — Access Control System

Version is pinned in `system.yaml` (`references[variants].version`) — that single pin is what
the report cites; bump it on any change to the table or triage log below, then reissue the report.

Maps shipped SKUs/firmware to the threat-model **profiles** in `system.yaml`. The model
assesses profiles; this register justifies that each SKU is covered by one. Columns are the
profiling `properties:` keys from the model — same tuple → same profile, so classification is
mechanical and the "why" is just the evidence + the triage log below.

Regenerate when a SKU is added or reclassified — **not** per build. The release manifest
references this file; it does not duplicate it.

## Classification

| SKU | intelligence | power | transport | fail-mode | → profile | evidence |
|---|---|---|---|---|---|---|
| ACME-4471 | dumb | mains | wired | per-door | `lock` | datasheet §3; teardown 2024-11 |
| ACME-7780 | smart | battery | offline | fail-secure | `lock-battery-standalone` | fw manifest 2.x; teardown 2025-02 |
| ACME-9012 | smart | mains | wired | per-door | `lock-connected` | datasheet §4; integration test rig |

A SKU whose measured tuple matches no profile is the substantial-modification trigger — it
needs a profile (and an EMB3D pass) before it can ship under this model.

## Triage log

One dated line per firmware/hardware change, recording whether it moved a **profiling
property** (→ reclassify + re-assess) or not (→ covered, no re-assessment). This is the
CRA substantial-modification record for each variant.

### ACME-4471 (`lock` profile)
- **fw 1.5 — 2026-06-27 — no threat change.** Patch touched the LED driver only; no profiling
  property moved (still dumb / mains / wired), introduces no firmware or comms surface. Covered
  by the `lock` assessment; security update, not a substantial modification.

### ACME-7780 (`lock-battery-standalone` profile)
- **fw 2.0 — _date_ — _classify_.** _e.g. added BLE sync → `transport` moves offline→wireless;
  pulls in a wireless EMB3D surface; re-assess before shipping._

### ACME-9012 (`lock-connected` profile)
- _no changes logged yet._
