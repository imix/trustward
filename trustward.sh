#!/usr/bin/env bash
# Run trustward via Docker, mounting the current directory as /model.
#
# Usage:
#   ./trustward.sh <command> [args...]          — pass any trustward command
#   ./trustward.sh render [--pdf]               — generate and render the report
#
# Examples:
#   ./trustward.sh diagram dataflow
#   ./trustward.sh render                       # produces out/report.html
#   ./trustward.sh render --pdf                 # also produces out/report.pdf
set -euo pipefail

IMAGE=${TRUSTWARD_IMAGE:-trustward}

# Run as the host user so files written into the mount (exported templates,
# rendered .qmd/.html) are owned by you, not root.
run=(docker run --rm -u "$(id -u):$(id -g)" -v "$(pwd):/model")

if [[ "${1:-}" == "render" ]]; then
    mkdir -p out
    # Generate the .qmd at the model root so Quarto resolves relative paths (logo,
    # images) against your model dir, then direct all output into out/.
    "${run[@]}" "$IMAGE" report "${@:2}" > report.qmd
    # HOME=/tmp: the mapped UID has no home in the image; Quarto needs a writable one.
    "${run[@]}" -e HOME=/tmp --entrypoint quarto "$IMAGE" render report.qmd --output-dir out
    mv -f report.qmd out/
else
    "${run[@]}" "$IMAGE" "$@"
fi
