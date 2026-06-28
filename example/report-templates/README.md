# Report templates

Standard-shaped report shells — **not models**. A conformance report needs
manual narrative that isn't threat-modeling data (product description, intended
use, operational environment, assumptions). That content belongs in the
template (your owned document shell), not in the model. These shells give you
the full structure a standard expects, with the manual sections stubbed as
`_fill in_` prompts wrapping the same model includes.

| File | Shape |
|---|---|
| `en40000.tmpl` | A prEN 40000-1-2 conformance report: front matter, the manual narrative sections, and the model-driven assessment sections. |

## Use it

Copy a shell into your model directory as `report.tmpl` (trustward picks up a
local `report.tmpl` over the built-in default), fill in the `_fill in_`
sections, then render:

```
cp /path/to/trustward/example/report-templates/en40000.tmpl report.tmpl
# edit the _fill in_ sections
../../trustward.sh render
```

[../access-control](../access-control) ships its own filled-in `report.tmpl`
derived from this shell — a worked end state.

## On standards and copyright

See the repository [README](../../README.md) — these shells carry only section
structure, not a standard's normative text, and prEN 40000 is still a draft.

