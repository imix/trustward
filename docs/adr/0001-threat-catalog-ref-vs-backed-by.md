# Threat–catalog links: `ref` for inheritance, `backed-by` for citation

A threat relates to a `threat-catalog` pattern two deliberately separate ways. `ref:` instantiates **one** pattern and inherits its title/type/severity/notes — for reusing a recurring threat definition. `backed-by:` cites **many** patterns as traceability to an external taxonomy (EMB3D, ATT&CK, CWE) with no inheritance. They are kept distinct because "reuse this definition" and "this threat is evidenced by these standard IDs" are different needs; both resolve to catalog patterns, so a typo'd id fails validation either way.

## Considered Options

- **EMB3D as freeform notes** (`[TID-119]` woven into prose) — rejected: typos pass silently, and the mapping is neither queryable nor validated.
- **Overloading `ref:` for taxonomy links** — rejected: `ref:` is single-valued with inheritance; a threat legitimately cites several TIDs and wants none of their fields.
