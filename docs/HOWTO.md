# How to model with trustward

This is the **judgment** layer: how *far* to take each part of a model, and when the
simple form is enough versus when to reach for detail. For *what* each element means see
[GLOSSARY.md](GLOSSARY.md); for the exact YAML see [MODEL.md](MODEL.md); for a first
walkthrough see "Start your own model" in the [README](../README.md).

The model scales from a back-of-the-envelope sketch to a CRA / prEN 40000-1-2 §6
conformance artifact. The same elements serve both — you decide how much to fill in.

> **The one principle:** start minimal, add detail only on a *signal* — a decision you
> need to record, a threat that has to cite what it breaks, or a conformance section you
> have to produce. Detail that nothing reads is just maintenance cost.

> prEN 40000-1-2 is a draft; clause numbers below are how trustward maps onto its
> process, not quotations from it.

## How much model is "enough"?

The floor is `system` + `components` + `data-flows` — that already renders a data flow
diagram and validates. Add the rest when the work asks for it:

- **`assets`** — once you want to say *what's worth protecting* and reason about it.
- **`trust-zones`** — once boundary crossings matter (see below).
- **`threats` + `controls`** — the core of any actual analysis.
- **`objectives`** — when you want the objective→asset→threat trace (see below).
- **`risk-policy`** — when you want scored, gated, conformance-shaped risk (see below).

You don't need all of it. A diagram plus a dozen threats and their mitigations is a
perfectly good lightweight model.

## Objectives: "CIA and done" vs fine-grained

Objectives are optional, and how granular they should be is the most common question.

**Coarse — "CIA and done."** Define a handful of project-level objectives — even just
one per CIA-scale property you care about (`confidentiality`, `integrity`,
`availability`, `authenticity`, `accountability`) — and attach them broadly to assets.
Reach for this when the model is small or internal, the protected properties are
uniform, or you're not producing a §6.5.2 conformance section. It's cheap and still
gives you a CIA view of the assets.

**Fine-grained — named, asset-spanning, threat-linked.** Define distinct, named
objectives and point threats at them with `violates:`. Reach for this when:

- the **consequence or audience differs** for the *same* property. A smart EV charger, for
  example, separates *control-and-load integrity* (a *safety* concern — forged load
  commands) from *billing integrity* (a *commercial* concern — metering fraud). Both are
  "integrity," but a reader reasons about them completely differently.
- a **threat needs to cite what it breaks** — `threats[].violates` makes "this attack
  defeats *this* objective" explicit, which a bare per-asset CIA tag can't express.
- you need the **objective → asset → threat trace** in the report (§6.5.2 ↔ §6.5.3),
  e.g. for a CRA conformance artifact.

> **Rule of thumb:** split an objective out when *naming it changes how a reader reasons
> about the risk*. Otherwise keep it coarse. An objective can span several assets (one
> objective, many `assets[].objectives` references), so you rarely need one per asset.

## Asset granularity

Group assets by **shared protection need / shared consequence**, not by data structure.
"Billing records" is one asset even if it's three database tables; don't model every
field. If two things are always protected the same way and failing them has the same
consequence, they're one asset. Split when a threat or control applies to one but not the
other.

## Threats: inline vs catalog, and how many

Aim for **STRIDE coverage per target** — for each component or data flow, ask which of
spoofing / tampering / repudiation / disclosure / denial / elevation actually apply, and
write the ones that do. (Threat `type` is free text; STRIDE is the recommended
vocabulary, not enforced — so stay consistent yourself.)

Keep threats **inline** by default. Factor a reusable pattern into a `threat-catalog`
(and reference it with `ref: catalog-id::pattern-id`) only when the *same* threat recurs
across many components or models and you want to define its title/severity/notes once.
For a single system, inline is simpler.

## Scoring method: `qualitative` vs `etsi-tvra`

- **`qualitative`** (the default) — you state `likelihood` and `impact` (low/medium/high)
  and a 3×3 matrix yields the level. Right for almost everything, and for early work.
- **`etsi-tvra`** — likelihood is *derived* from a per-threat `attack:` block (expertise,
  knowledge, opportunity, equipment → an attack potential). Reach for it when you want a
  defensible, attacker-effort-based likelihood for regulated or contested risk. It's only
  worth filling in the `attack:` blocks if you've chosen this method; otherwise skip them.

You set the method once, in `risk-policy.method`.

## The risk-policy and the CRA gate: when to turn it on

Declaring a `risk-policy` is the switch that turns a *threat list* into a *gated risk
assessment*. With it present, `validate` enforces the **CRA gate**: every risk whose
computed level isn't in `accept:` must carry a `treatment` (mitigate/accept/transfer/avoid)
and an `owner`. It also lights up the §6.3 (criteria), §6.5.5 (evaluation) and §6.7
(monitoring) sections of the report.

Turn it on when you want conformance enforcement and sign-off discipline. Leave it off for
a lightweight model — without a `risk-policy` there's no gate, and untreated threats are
fine. Don't switch it on until you're ready to assign treatments and owners, or `validate`
will (correctly) start failing.

## Trust zones & data flows: how much segmentation

Model `trust-zones` when **boundary crossings carry meaning** — a data flow between two
zones is a trust boundary and a cue for extra scrutiny. For a handful of components on one
network, zones add noise; for a device that spans field / local / cloud, they're the point.
Don't over-segment: a zone earns its place only if something crosses into or out of it.

## Compliance mapping: when

Plain `controls` (a title, a description, maybe `evidence`) are enough to record what you
do. Add a `catalog` of requirements and point controls at them with
`ref: catalog-id::req-id` **only when you need coverage/gap evidence** against a standard
(e.g. an IEC 62443-4-2 subset) — that's what produces the Compliance Evidence section with
its `covered` / `gap` rows. A baseline catalog can map onto external standards via
`satisfies:` (those targets aren't validated — they may name requirements you don't load).
Skip all of this for an internal model.

## Splitting into files

One `system.yaml` holds an entire model and is fine until it's unwieldy. Split via
`imports:` when you want to separate by subsystem, by concern (system vs threats vs
catalogs), or to version files independently — the loader merges everything depth-first
from `system.yaml`. Structure is for *your* convenience; the merged model is identical
either way.

## Two postures

| | Lightweight sketch | CRA conformance artifact |
|---|---|---|
| Diagram (system/components/flows) | ✓ | ✓ |
| Assets | optional | ✓ |
| Objectives | skip or coarse | fine-grained, threat-linked |
| Threats + controls | ✓ | ✓ |
| `risk-policy` (CRA gate) | off | on |
| Treatments + owners | optional | required for non-accepted risk |
| Compliance catalog | skip | mapped, gaps surfaced |
| `risk-policy.review` (§6.7) | skip | filled in |

Pick the posture you actually need. Most models live near the left and move right only for
the parts that face an assessor.
