#!/usr/bin/env bash
# Run trustward via Docker, mounting the current directory as /model.
#
# Usage:
#   ./trustward.sh <command> [args...]          — pass any trustward command
#   ./trustward.sh render [report-type]         — generate and render a report
#
# Examples:
#   ./trustward.sh diagram dataflow
#   ./trustward.sh render threat-model          # produces threat-model.html
#   ./trustward.sh render                       # defaults to threat-model
set -euo pipefail

IMAGE=${TRUSTWARD_IMAGE:-trustward}

if [[ "${1:-}" == "render" ]]; then
    report="${2:-threat-model}"
    qmd="${report}.qmd"
    docker run --rm -v "$(pwd):/model" "$IMAGE" report "$report" "${@:3}" > "$qmd"
    docker run --rm -v "$(pwd):/model" --entrypoint quarto "$IMAGE" render "$qmd"
else
    docker run --rm -v "$(pwd):/model" "$IMAGE" "$@"
fi
