# trustward

Threat models that live next to your code.

Define your system — its assets, cybersecurity objectives, threats, and controls — in plain text. trustward renders data flow diagrams and a risk-management report shaped to the CRA / prEN 40000-1-2 §6 process, in your editor, in CI, or on every pull request. (A *ward* is a guarded zone; trust zones are what the model maps.)

- **Reviewable in PRs.** Threats and mitigations are text diffs. Your team can discuss risk in the same place they discuss code.
- **Version-controlled history.** Git tells you when a threat was identified, when a mitigation was added, when residual risk was accepted — and by whom.
- **No proprietary tooling.** No licenses, no accounts, no vendor lock-in. A directory of text files and a Docker image.

## Quick start

```bash
# Build the Docker image once
docker build -t trustward .

# Try it on the example model
cd example/fire-protection-system
../../trustward.sh render                 # writes out/report.html
../../trustward.sh diagram dataflow       # prints a Mermaid data flow diagram
```

`trustward.sh` wraps the Docker image and mounts the **current directory** as the model directory — always run it from the directory containing your `system.yaml`. The image is the only runtime dependency.

## Start your own model

A model directory needs exactly one file to begin with: `system.yaml`. This is enough to render a diagram and a report:

```yaml
version:
  semver: 0.1.0
  releasedate: 2026-06-10

system:
  id: my-system
  title: My System
  description: One paragraph on what the system does and who uses it.

components:
  - id: api-server
    title: API Server
    type: server
    description: Serves the public API.

  - id: database
    title: Database
    type: server
    description: Stores user data.

data-flows:
  - id: flow-api-db
    title: API → Database
    connects: [api-server, database]
    description: SQL over TLS.
```

```bash
cd my-system
/path/to/trustward.sh diagram dataflow    # quick check that the model loads
/path/to/trustward.sh render              # out/report.html
```

From there, grow the model incrementally:

1. Add `assets:` and attach them to components and data flows — what is worth protecting.
2. Add `trust-zones:` to group components by security boundary — they become subgraphs in the diagram.
3. Add `threats:` (and the `controls:` that mitigate them) — they become the core of the report.
4. When `system.yaml` gets large, split it into multiple files linked via `imports:`. The loader starts at `system.yaml`, follows imports depth-first, and merges all top-level keys into a single model. Split by concern, nest by subsystem — any structure works.

Run `trustward.sh validate` as you go — it catches typos in cross-references (a threat mitigated by a control that doesn't exist, a flow connecting a renamed component) that would otherwise silently produce wrong reports.

[docs/MODEL.md](docs/MODEL.md) documents every key. The [example/](example) directory has worked models: [fire-protection-system](example/fire-protection-system) is the complete baseline (imports, catalogs, threats, controls), and [access-control](example/access-control) is the advanced one (the opt-in risk layer with attack-potential scoring and the CRA gate, an EMB3D device pass, and external references). For *how far* to take each part — coarse vs detailed, and when — see [docs/HOWTO.md](docs/HOWTO.md).

## Commands

Run all commands from your model directory.

### `trustward.sh render [flags]`

Generates and renders the report in one step:

```bash
trustward.sh render                       # writes out/report.html
trustward.sh render --pdf                 # also writes out/report.pdf
```

Under the hood this runs `trustward report` (which prints a Quarto `.qmd` document to stdout) and then renders it with Quarto. All generated files land in `out/` (`report.qmd`, `report.html`, `report_files/`) — that whole directory is a by-product; it's `.gitignore`d by default, or upload it as a CI artifact.

### `trustward.sh diagram dataflow`

Prints a [Mermaid](https://mermaid.js.org) flowchart to stdout. Components are grouped by trust zone; data flows appear as labeled edges. Paste it into anything that renders Mermaid — Markdown files, [mermaid.live](https://mermaid.live), wikis.

```bash
trustward.sh diagram dataflow
```

### `trustward.sh validate`

Checks the referential integrity of the model and exits non-zero if anything is broken — made for CI and pre-commit hooks:

```bash
trustward.sh validate
```

It verifies that every cross-reference resolves to a declared ID:

- threat `target` → a component or data flow; threat `asset` → an asset; threat `mitigations` → controls; threat `ref` and `backed-by` → threat catalog patterns
- component `assets` → assets; component `controls` → controls
- trust zone `members` → components
- data flow `connects` → exactly two components; data flow `assets` → assets
- control `ref` → a control catalog requirement
- reference `location` (when a local path) → an existing file; and every reference declares a `version`
- every entity has an `id`, and IDs are unique within each entity kind

Requirement `satisfies` entries are deliberately **not** checked — they may point at external standards (e.g. `iec-62443-sl2::SR-1.1`) that are not part of the model.

### `trustward.sh report --format json`

Prints the **risk register** as JSON instead of a Quarto document — each threat joined with its computed evaluation (level, likelihood, accepted/treated/open) and treatment decision. Pure Go (no Quarto/Docker needed; works with the bare binary too), so it runs in a minimal CI container. Blocking CI on open risks is `validate`'s job; use the JSON when a pipeline needs its own predicate over the numbers — a residual ceiling, an alarm count, a cross-version diff.

```bash
trustward.sh report --format json > register.json
```

### `trustward.sh template export`

Writes the built-in report template to `report.tmpl` in your model directory, as a starting point for customisation (see below). Refuses to overwrite an existing file.

## Customising the report

For anything beyond a quick look, **export the template and own it** — it's where your branding and document framing live, and most real deployments need both. The built-in template renders out of the box (and prints a reminder pointing you here), but exporting is the recommended first step for a model you'll keep:

```bash
trustward.sh template export
# edit report.tmpl
trustward.sh render
```

If `report.tmpl` exists in your model directory, trustward uses it instead of the built-in. It's a [Go `text/template`](https://pkg.go.dev/text/template) file. Customise it for:

- **Branding** — theme, fonts, logo treatment, and title-block styling (all in the Quarto front matter).
- **Document framing** — trustward owns the threat model and risk assessment; the report is one artifact in a larger conformance set. The system design, asset inventory, and other documents live elsewhere. Add a "Related documents" section that **links out** to them rather than reproducing them here — keeping a single source of truth for each and avoiding drift. For documents whose *version* matters, declare them in the model's `references:` (see [docs/MODEL.md](docs/MODEL.md)) instead of hand-listing them: the built-in template renders a References table with each pinned version, and `validate` checks they resolve — so the report can't cite a stale version.

### Two starting points: the default, or a standard-shaped shell

The built-in template (`template export`) is **model-driven only**: every section is populated from your YAML, nothing manual, and the headings are standard-agnostic (Pandoc numbers them). It renders for any model regardless of the standard you're targeting.

But a conformance report also needs **manual narrative that isn't threat-modeling data** — product description, intended use and users, operational environment, assumptions. That content belongs in the template (your owned doc shell), not in the model. To save you scaffolding it, [`example/report-templates/`](example/report-templates) ships **standard-shaped shells**: the full structure a given standard expects, with those manual sections as `_fill in_` prompts wrapping the same model includes.

```bash
cp /path/to/trustward/example/report-templates/en40000.tmpl report.tmpl
# fill in the _fill in_ prompts, then render
trustward.sh render
```

`en40000.tmpl` is the prEN 40000-1-2 §6 shell — explicit clause numbers (6.2 Product Context … 6.7 Monitoring) and the §6.2 manual context subsections. Copy it, or use the minimal default and add only what you need.

> **On standards and copyright.** A shell reproduces only a standard's *structure* — clause numbers and section headings — as a scaffold for your own content; it ships **none** of the standard's normative text, and prEN 40000 is still a *draft* whose numbering may change before publication. Standards are copyrighted works sold by their publishers (CEN, IEC, …): obtain the standard itself from its publisher, and note that whether reproducing its section structure is permitted (fair use, fair dealing, or a local exception) depends on your jurisdiction and is yours to assess. Nothing here is legal advice.

The template receives:

| Field | Type | Description |
|-------|------|-------------|
| `.Title` | string | System title |
| `.Date` | string | Release date |
| `.Version` | string | SemVer |
| `.Description` | string | System description |
| `.Logo` | string | Logo path from the system metadata |
| `.References` | `[]Reference` | External versioned docs — variant register, requirements, standards, SBOM |
| `.AssetList` | `[]Asset` | All assets |
| `.AssetComponents` | `map[string][]string` | Asset ID → component IDs that hold it |
| `.ObjectiveList` | `[]Objective` | Cybersecurity objectives (§6.5.2) |
| `.ObjectiveAssets` | `map[string][]string` | Objective ID → asset IDs that uphold it |
| `.ThreatGroups` | `[]ThreatGroup` | Threats grouped by target (`.TargetID`, `.TargetTitle`, `.Threats`), in encounter order |
| `.ThreatList` | `[]Threat` | Flat threat list, for the risk register |
| `.RiskEval` | `map[string]Eval` | Threat ID → computed evaluation (`.Level`, `.Likelihood`, `.Accepted`, `.Treated`, `.Open`) |
| `.RiskMatrix` | `Matrix` | Likelihood×impact heatmap (`.Impacts`, `.Rows[].Cells[]`, `.Unplaced`) |
| `.RiskMethod` | string | Scoring method from the risk-policy |
| `.RiskAccept` | `[]string` | Accepted risk levels |
| `.RiskReview` | string | Monitoring & review cadence (§6.7) |
| `.RiskPolicySet` | bool | Whether a risk-policy is declared — gate the risk sections on this |
| `.Controls` | `map[string]string` | Control ID → title (for the `controlTitle` helper) |
| `.ControlList` | `[]Control` | Full control objects, for a controls section |
| `.ControlComponents` | `map[string][]string` | Control ID → component IDs that implement it |
| `.ComponentList` | `[]Component` | All components |
| `.CatalogList` | `[]ControlCatalog` | Requirement catalogs, for compliance mapping |
| `.RequirementControls` | `map[string][]string` | `catalog-id::req-id` → control IDs that satisfy it |
| `.Diagram` | string | Rendered Mermaid diagram source |
| `.PDF` | bool | Whether PDF output was requested |

Built-in template functions: `controlTitle <controls> <id>`, `join <list> <sep>`, `upper <string>`, `trim <string>`, `riskColor <level>` (a heatmap fill colour for the risk matrix).

The template's front matter is regular Quarto config — theme, table of contents, Mermaid theme, and output formats are all controlled there.

## Using in CI

A minimal GitHub Actions step, assuming your model lives in `my-system/` and `trustward.sh` at the repo root:

```yaml
- name: Validate and render threat model
  run: |
    docker build -t trustward .
    cd my-system
    ../trustward.sh validate
    ../trustward.sh render
- uses: actions/upload-artifact@v4
  with:
    name: threat-model-report
    path: my-system/out/report.html
```

## Installing the binary

```bash
go install github.com/imix/trustward/cmd/trustward@latest
```

This puts the `trustward` binary on your `PATH` (under `$(go env GOPATH)/bin`). Like any bare binary it generates diagrams and `.qmd` documents but does **not** render — rendering needs Quarto, which only the Docker image bundles. Use it for `validate` in CI and for diagram generation; use `trustward.sh` (Docker) when you want rendered HTML/PDF.

## Building from source

```bash
go build -o trustward ./cmd/trustward/
go test ./...
```

Requires Go 1.25+. The Docker image also bundles [Quarto](https://quarto.org) for rendering; the bare binary only generates diagrams and `.qmd` documents.

## Reference

- [docs/HOWTO.md](docs/HOWTO.md) — how to model: how far to take each part, and when
- [docs/MODEL.md](docs/MODEL.md) — complete YAML schema reference
- [docs/GLOSSARY.md](docs/GLOSSARY.md) — domain term definitions
- [docs/adr/](docs/adr) — architecture decision records: why the model is shaped the way it is

## AI assistance

trustward — its code, documentation, and example models — was built with substantial help from AI coding tools. A human directed and reviewed the work, but treat it accordingly: it is a **modelling aid, not a certified compliance tool**. It does not by itself guarantee conformance with the CRA, prEN 40000-1-2, or any standard — the output is only as good as the model you write and the review you give it. Validate the results and have a qualified person sign off any assessment you rely on.
