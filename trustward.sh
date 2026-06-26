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

# Run as the host user so files written into the mount (exported templates,
# rendered .qmd/.html) are owned by you, not root.
run=(docker run --rm -u "$(id -u):$(id -g)" -v "$(pwd):/model")

if [[ "${1:-}" == "render" ]]; then
    report="${2:-threat-model}"
    qmd="${report}.qmd"
    "${run[@]}" "$IMAGE" report "$report" "${@:3}" > "$qmd"
    # HOME=/tmp: the mapped UID has no home in the image; Quarto needs a writable one.
    "${run[@]}" -e HOME=/tmp --entrypoint quarto "$IMAGE" render "$qmd"
else
    "${run[@]}" "$IMAGE" "$@"
fi
