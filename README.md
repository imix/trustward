# sectrack

A YAML-based security modeling tool. Define your system, its threats, and your controls in plain text files — then generate data flow diagrams and threat model reports.

---

## Quick start

```bash
# Build the Docker image
docker build -t sectrack .

# From your model directory
cd example/fire-protection-system

# Generate a Mermaid data flow diagram
../../sectrack.sh diagram dataflow

# Generate and render a threat model report to HTML
../../sectrack.sh render threat-model       # produces threat-model.html
```

---

## Project layout

A sectrack project is a directory containing YAML files. The tool always starts from `system.yaml` and follows `imports:` declarations depth-first.

```
my-system/
  system.yaml          ← entry point: system metadata, assets, components,
                          trust zones, data flows
  threat-model.yaml    ← threats (imports system.yaml)
  company.yaml         ← controls (imported by system.yaml or threat-model.yaml)
```

Files are containers for version management only. Any file can hold any combination of top-level keys. A single `system.yaml` with everything in it is valid. See [MODEL.md](MODEL.md) for the full schema.

---

## Commands

### `sectrack diagram dataflow`

Prints a [Mermaid](https://mermaid.js.org) flowchart to stdout showing components grouped by trust zone and data flows as labeled edges.

```bash
sectrack diagram dataflow
```

Pipe to a file or embed directly in Markdown.

### `sectrack report threat-model [--pdf]`

Prints a [Quarto](https://quarto.org) document (`.qmd`) to stdout. Use `sectrack.sh render` to generate and render in one step:

```bash
./sectrack.sh render threat-model    # produces threat-model.html
./sectrack.sh render threat-model --pdf
```

Or manually:

```bash
sectrack report threat-model > report.qmd
quarto render report.qmd
```

Add `--pdf` to include a PDF target in the front matter (requires Chrome headless).

### `sectrack template export threat-model`

Writes the built-in report template to `./templates/threat-model.tmpl` as a starting point for customization.

```bash
sectrack template export threat-model
# edit templates/threat-model.tmpl
# subsequent report runs will use your version automatically
```

---

## Template customization

Report templates are [Go `text/template`](https://pkg.go.dev/text/template) files. The template has access to:

| Field | Type | Description |
|-------|------|-------------|
| `.Title` | string | System title |
| `.Date` | string | Release date |
| `.Version` | string | SemVer |
| `.Description` | string | System description |
| `.Threats` | `[]Threat` | All threats |
| `.Controls` | `map[string]string` | Control ID → title |
| `.Diagram` | string | Rendered Mermaid diagram |
| `.PDF` | bool | Whether PDF output was requested |

Built-in template functions: `controlTitle <controls> <id>`, `join <sep> <list>`, `upper <string>`.

To change the Mermaid theme, edit the `format:` block in the template front matter:

```yaml
format:
  html:
    mermaid:
      theme: neutral   # default | neutral | dark | forest | base
```

---

## Building from source

```bash
cd tool
go build -o sectrack ./cmd/sectrack/
go test ./...
```

Requires Go 1.25+. The Docker image also includes [Quarto](https://quarto.org) for rendering.

---

## Reference

- [MODEL.md](MODEL.md) — complete YAML schema reference
- [GLOSSARY.md](GLOSSARY.md) — domain term definitions
