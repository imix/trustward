#!/usr/bin/env bash
# Run trustward via Docker, mounting the current directory as /model.
set -euo pipefail

IMAGE=${TRUSTWARD_IMAGE:-trustward}

run=(docker run --rm -u "$(id -u):$(id -g)" -v "$(pwd):/model")

usage() {
    cat <<'EOF'
trustward.sh — run trustward via Docker, mounting the current directory as /model.

Usage:
  ./trustward.sh <command> [args...]   run any trustward command (validate, diagram, report, template)
  ./trustward.sh render [--pdf]        generate the report and render it with Quarto into out/

render is provided by this wrapper, not the binary: it runs `trustward report`,
renders the result with Quarto, and writes out/report.html (and out/report.pdf
with --pdf). Always run from the directory containing your system.yaml.

Examples:
  ./trustward.sh validate
  ./trustward.sh diagram dataflow
  ./trustward.sh render                # writes out/report.html
  ./trustward.sh render --pdf          # also writes out/report.pdf

Binary commands (trustward --help):
EOF
    "${run[@]}" "$IMAGE" --help || true
}

# Wrapper-level help: render is shell-only, so the binary's --help can't list it.
if [[ $# -eq 0 || "${1:-}" =~ ^(-h|--help|help)$ ]]; then
    usage
    exit 0
fi

if [[ "${1:-}" == "render" ]]; then
    # render --help must not fall through to `report --help`, whose stdout would
    # otherwise be captured into report.qmd and rendered as the "report".
    for a in "${@:2}"; do
        if [[ "$a" =~ ^(-h|--help)$ ]]; then usage; exit 0; fi
    done

    wants_pdf=false
    for a in "${@:2}"; do [[ "$a" == "--pdf" ]] && wants_pdf=true; done

    mkdir -p out
    # Pre-render the data flow diagram to a static SVG with the Graphviz CLI (no
    # browser). Quarto's executable {dot}/{mermaid} cells need a headless Chrome
    # for non-HTML output; a pre-rendered image sidesteps that — Typst embeds the
    # SVG straight into the PDF.
    if ! "${run[@]}" "$IMAGE" diagram dot > diagram.gv; then
        rm -f diagram.gv
        echo "trustward.sh: could not generate the diagram; nothing rendered (see the error above)" >&2
        exit 1
    fi
    "${run[@]}" --entrypoint dot "$IMAGE" -Tsvg diagram.gv -o diagram.svg

    # Generate the .qmd at the model root so Quarto resolves relative paths (logo,
    # images) against your model dir. Fail closed: only render a real document, so
    # a failed run never becomes a garbage report.html.
    if ! "${run[@]}" "$IMAGE" report "${@:2}" > report.qmd; then
        rm -f report.qmd
        echo "trustward.sh: 'trustward report' failed; nothing rendered (see the error above)" >&2
        exit 1
    fi
    if [[ ! -s report.qmd ]] || ! head -n1 report.qmd | grep -q '^---'; then
        rm -f report.qmd
        echo "trustward.sh: 'trustward report' did not produce a Quarto document; nothing rendered" >&2
        exit 1
    fi

    # HOME=/tmp: the mapped UID has no home in the image; Quarto needs a writable one.
    "${run[@]}" -e HOME=/tmp --entrypoint quarto "$IMAGE" render report.qmd --output-dir out
    # All generated artifacts live under out/ (report.html embeds the SVG via
    # embed-resources, so the loose diagram files are just byproducts).
    mv -f report.qmd diagram.gv diagram.svg out/ 2>/dev/null || true

    if $wants_pdf && [[ ! -f out/report.pdf ]]; then
        echo "trustward.sh: --pdf was requested but out/report.pdf was not produced." >&2
        echo "  PDF renders via Quarto's Typst engine plus Graphviz for the diagram." >&2
        echo "  Rebuild the image (docker build -t trustward .) so it bundles graphviz, or" >&2
        echo "  check the Quarto output above for the rendering error." >&2
        exit 1
    fi
else
    "${run[@]}" "$IMAGE" "$@"
fi
