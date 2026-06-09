#!/usr/bin/env bash
# Run the sectrack tool via Docker, mounting the current directory as /model.
# Usage: ./sectrack.sh <command> [args...]
#
# Examples:
#   ./sectrack.sh diagram dataflow
#   ./sectrack.sh report threat-model > threat-model.qmd
#
# To render a .qmd file with Quarto:
#   docker run --rm -v "$(pwd):/model" --entrypoint quarto sectrack render threat-model.qmd
set -euo pipefail

IMAGE=${SECTRACK_IMAGE:-sectrack}

docker run --rm \
  -v "$(pwd):/model" \
  "$IMAGE" \
  "$@"
