# Rendering is delegated to Quarto; the binary emits `.qmd`, not HTML

trustward generates a Quarto `.qmd` document and leaves rendering to Quarto, which the Docker image bundles. The bare `go install` binary therefore validates and emits diagrams and `.qmd`, but **cannot** produce HTML/PDF — that needs the Docker image (or a local Quarto). Chosen so trustward owns only the model→document mapping and inherits Quarto's theming, TOC, math, and multi-format output for free instead of building a renderer.

## Consequences

- HTML/PDF requires Docker or a local Quarto install.
- CI typically uses the bare binary for `validate` and the image for rendering.
