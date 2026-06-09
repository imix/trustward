#!/usr/bin/env bash
# Run sectrack via Docker, mounting the current directory as /model.
#
# Usage:
#   ./sectrack.sh <command> [args...]          — pass any sectrack command
#   ./sectrack.sh render [report-type]         — generate and render a report
#
# Examples:
#   ./sectrack.sh diagram dataflow
#   ./sectrack.sh render threat-model          # produces threat-model.html
#   ./sectrack.sh render                       # defaults to threat-model
set -euo pipefail

IMAGE=${SECTRACK_IMAGE:-sectrack}

if [[ "${1:-}" == "render" ]]; then
    report="${2:-threat-model}"
    qmd="${report}.qmd"
    docker run --rm -v "$(pwd):/model" "$IMAGE" report "$report" > "$qmd"
    docker run --rm -v "$(pwd):/model" --entrypoint quarto "$IMAGE" render "$qmd"
else
    docker run --rm -v "$(pwd):/model" "$IMAGE" "$@"
fi
