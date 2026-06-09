# sectrack

Threat models that live next to your code.

Define your system, its threats, and your security controls in plain text files. Generate data flow diagrams and threat model reports — in your editor, in CI, or on every pull request.

---

## Why

- **Reviewable in PRs.** Threats and mitigations are text diffs. Your team can discuss risk in the same place they discuss code.
- **Version-controlled history.** Git tells you when a threat was identified, when a mitigation was added, when residual risk was accepted — and by whom.
- **No proprietary tooling.** No licenses, no accounts, no vendor lock-in. A directory of text files and a Docker image.
- **CI/CD friendly.** The Docker image is the only runtime dependency. Run it in any pipeline that can pull a container.

---

## Quick start

```bash
# Build the Docker image once
docker build -t sectrack .

# From your model directory
cd example/fire-protection-system

# Render the threat model report to HTML
../../sectrack.sh render threat-model       # produces threat-model.html
```

---

## Project layout

The only required file is `system.yaml` — the entry point. Everything else is optional and can be split across as many files as you like, or kept in one. The tool starts from `system.yaml` and follows `imports:` declarations depth-first, merging all top-level keys into a single model.

Split by concern, keep it flat, nest by subsystem — structure it however fits your repo. See [MODEL.md](MODEL.md) for the full schema.

---

## Commands

### `sectrack diagram dataflow`

Prints a [Mermaid](https://mermaid.js.org) flowchart to stdout. Components are grouped by trust zone; data flows appear as labeled edges.

```bash
../../sectrack.sh diagram dataflow
```

### `sectrack report threat-model`

Generates and renders a threat model report in one step:

```bash
../../sectrack.sh render threat-model    # produces threat-model.html
../../sectrack.sh render threat-model --pdf
```

### `sectrack template export threat-model`

Writes the built-in report template to `./templates/threat-model.tmpl`. Edit it to customise the report — sectrack picks it up automatically on the next render.

```bash
../../sectrack.sh template export threat-model
# edit templates/threat-model.tmpl
../../sectrack.sh render threat-model
```

---

## Customising the report

Place a [Go `text/template`](https://pkg.go.dev/text/template) file at `templates/threat-model.tmpl` in your model directory and sectrack will use it instead of the built-in one. The built-in template is at `tool/internal/quarto/templates/threat-model.tmpl` — or export it as a starting point:

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

To change the Mermaid theme, edit the `format:` block in the template front matter:

```yaml
format:
  html:
    mermaid:
      theme: neutral   # default | neutral | dark | forest | base
```

---

## Using in CI

The Docker image is self-contained — it's the only runtime dependency. A minimal GitHub Actions step:

```yaml
- name: Render threat model
  run: |
    docker build -t sectrack .
    cd my-system && ../../sectrack.sh render threat-model
- uses: actions/upload-artifact@v4
  with:
    name: threat-model
    path: my-system/threat-model.html
```

---

## Building from source

```bash
cd tool
go build -o sectrack ./cmd/sectrack/
go test ./...
```

Requires Go 1.21+. The Docker image also bundles [Quarto](https://quarto.org) for rendering.

---

## Reference

- [MODEL.md](MODEL.md) — complete YAML schema reference
- [GLOSSARY.md](GLOSSARY.md) — domain term definitions
