# PDF via Typst + Graphviz, no LaTeX and no browser

`render --pdf` must work inside the Docker image in a CI/CD pipeline — self-contained, no host dependencies, no silent no-op. The naive way to render this report to PDF needs two heavy engines: a LaTeX distribution (the `pdf:` format's engine, ~700MB–1GB) and headless Chromium. Both bloat the image every CI job pulls.

Decision: render PDF the lean, browser-free way.

- **Typst, not LaTeX.** The report's PDF format is `typst:`, not `pdf:`. Quarto bundles the Typst compiler, so the PDF engine costs **zero** added image size — the entire LaTeX layer is gone.
- **Graphviz CLI, not a browser.** This is the subtle part. Quarto's *executable* diagram cells — ` ```{mermaid} ` **and** ` ```{dot} ` alike — render via a headless Chrome for any non-HTML output; Graphviz the language does not avoid the browser, Graphviz the *binary* does. So the report does **not** use a `{dot}` cell. Instead the render pipeline pre-renders the diagram with the `dot` CLI (`dot -Tsvg`, no browser) into a static `diagram.svg`, and the template embeds it as a plain image (`![](diagram.svg)`). Typst embeds the SVG natively into the PDF; HTML inlines it via `embed-resources`. The image bundles only `graphviz` (~50MB) — **no** Chromium, **no** TeX.

How the pieces fit:

- `internal/dot` emits Graphviz DOT from the model; `trustward diagram dot` prints it.
- `trustward.sh` runs `diagram dot | dot -Tsvg` to produce `diagram.svg`, then renders the report, which references that file.
- The standalone `trustward diagram dataflow` command stays **Mermaid** — it renders client-side in any Markdown viewer (GitHub, mermaid.live), where a browser is always present. The two renderers share a colour palette for parity.

Consequence, accepted deliberately: the rendered report's diagram is a pre-rendered Graphviz SVG, not a live diagram cell, so it is produced by the render pipeline rather than embedded in the bare `report` output. A bare-binary user who renders the `.qmd` themselves also produces `diagram.svg` themselves (`trustward diagram dot | dot -Tsvg -o diagram.svg`).

Failure is surfaced, not swallowed: `trustward.sh` checks that `out/report.pdf` was actually produced when `--pdf` was requested and exits non-zero otherwise, so a pipeline never reports success with a missing artifact.

Status: accepted.
