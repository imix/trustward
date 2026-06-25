# sectrack

Threat models that live next to your code.

Define your system, its threats, and your security controls in plain text files. Generate data flow diagrams and threat model reports — in your editor, in CI, or on every pull request.

- **Reviewable in PRs.** Threats and mitigations are text diffs. Your team can discuss risk in the same place they discuss code.
- **Version-controlled history.** Git tells you when a threat was identified, when a mitigation was added, when residual risk was accepted — and by whom.
- **No proprietary tooling.** No licenses, no accounts, no vendor lock-in. A directory of text files and a Docker image.

## Quick start

```bash
# Build the Docker image once
docker build -t sectrack .

# Try it on the example model
cd example/fire-protection-system
../../sectrack.sh render                 # writes threat-model.html
../../sectrack.sh diagram dataflow       # prints a Mermaid data flow diagram
```

`sectrack.sh` wraps the Docker image and mounts the **current directory** as the model directory — always run it from the directory containing your `system.yaml`. The image is the only runtime dependency.

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
/path/to/sectrack.sh diagram dataflow    # quick check that the model loads
/path/to/sectrack.sh render              # threat-model.html
```

From there, grow the model incrementally:

1. Add `assets:` and attach them to components and data flows — what is worth protecting.
2. Add `trust-zones:` to group components by security boundary — they become subgraphs in the diagram.
3. Add `threats:` (and the `controls:` that mitigate them) — they become the core of the report.
4. When `system.yaml` gets large, split it into multiple files linked via `imports:`. The loader starts at `system.yaml`, follows imports depth-first, and merges all top-level keys into a single model. Split by concern, nest by subsystem — any structure works.

Run `sectrack.sh validate` as you go — it catches typos in cross-references (a threat mitigated by a control that doesn't exist, a flow connecting a renamed component) that would otherwise silently produce wrong reports.

[MODEL.md](MODEL.md) documents every key. [example/fire-protection-system](example/fire-protection-system) is a complete model using imports, catalogs, threats, and controls.

## Commands

Run all commands from your model directory.

### `sectrack.sh render [report] [flags]`

Generates and renders a report in one step. The report type defaults to `threat-model`:

```bash
sectrack.sh render                       # writes threat-model.html
sectrack.sh render threat-model --pdf    # also writes threat-model.pdf
```

Under the hood this runs `sectrack report threat-model` (which prints a Quarto `.qmd` document to stdout) and then renders it with Quarto. The intermediate `threat-model.qmd` and `threat-model_files/` directory are by-products — add them and the rendered output to `.gitignore` if you only want them as CI artifacts.

### `sectrack.sh diagram dataflow`

Prints a [Mermaid](https://mermaid.js.org) flowchart to stdout. Components are grouped by trust zone; data flows appear as labeled edges. Paste it into anything that renders Mermaid — Markdown files, [mermaid.live](https://mermaid.live), wikis.

```bash
sectrack.sh diagram dataflow
```

### `sectrack.sh validate`

Checks the referential integrity of the model and exits non-zero if anything is broken — made for CI and pre-commit hooks:

```bash
sectrack.sh validate
```

It verifies that every cross-reference resolves to a declared ID:

- threat `target` → a component or data flow; threat `asset` → an asset; threat `mitigations` → controls; threat `ref` → a threat catalog pattern
- component `assets` → assets; component `controls` → controls
- trust zone `members` → components
- data flow `connects` → exactly two components; data flow `assets` → assets
- control `ref` → a control catalog requirement
- every entity has an `id`, and IDs are unique within each entity kind

Requirement `satisfies` entries are deliberately **not** checked — they may point at external standards (e.g. `iec-62443-sl2::SR-1.1`) that are not part of the model.

### `sectrack.sh template export threat-model`

Writes the built-in report template to `templates/threat-model.tmpl` in your model directory, as a starting point for customisation (see below). Refuses to overwrite an existing file.

## Customising the report

For anything beyond a quick look, **export the template and own it** — it's where your branding and document framing live, and most real deployments need both. The built-in template renders out of the box (and prints a reminder pointing you here), but exporting is the recommended first step for a model you'll keep:

```bash
sectrack.sh template export threat-model
# edit templates/threat-model.tmpl
sectrack.sh render
```

If `templates/threat-model.tmpl` exists in your model directory, sectrack uses it instead of the built-in. It's a [Go `text/template`](https://pkg.go.dev/text/template) file. Customise it for:

- **Branding** — theme, fonts, logo treatment, and title-block styling (all in the Quarto front matter).
- **Document framing** — sectrack owns the threat model and risk assessment; the report is one artifact in a larger conformance set. The system design, asset inventory, and other documents live elsewhere. Add a "Related documents" section that **links out** to them rather than reproducing them here — keeping a single source of truth for each and avoiding drift.

The template receives:

| Field | Type | Description |
|-------|------|-------------|
| `.Title` | string | System title |
| `.Date` | string | Release date |
| `.Version` | string | SemVer |
| `.Description` | string | System description |
| `.Threats` | `[]Threat` | All threats |
| `.Controls` | `map[string]string` | Control ID → title (for inline references) |
| `.ControlList` | `[]Control` | Full control objects (for a controls section) |
| `.Diagram` | string | Rendered Mermaid diagram source |
| `.PDF` | bool | Whether PDF output was requested |

Built-in template functions: `controlTitle <controls> <id>`, `join <sep> <list>`, `upper <string>`, `trim <string>`.

The template's front matter is regular Quarto config — theme, table of contents, Mermaid theme, and output formats are all controlled there.

## Using in CI

A minimal GitHub Actions step, assuming your model lives in `my-system/` and `sectrack.sh` at the repo root:

```yaml
- name: Validate and render threat model
  run: |
    docker build -t sectrack .
    cd my-system
    ../sectrack.sh validate
    ../sectrack.sh render
- uses: actions/upload-artifact@v4
  with:
    name: threat-model
    path: my-system/threat-model.html
```

## Building from source

```bash
cd tool
go build -o sectrack ./cmd/sectrack/
go test ./...
```

Requires Go 1.25+. The Docker image also bundles [Quarto](https://quarto.org) for rendering; the bare binary only generates diagrams and `.qmd` documents.

## Reference

- [MODEL.md](MODEL.md) — complete YAML schema reference
- [GLOSSARY.md](GLOSSARY.md) — domain term definitions
