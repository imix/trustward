# Stage 1: build the sectrack CLI
FROM golang:1.26-alpine AS builder
WORKDIR /build
COPY tool/ .
RUN go build -o sectrack ./cmd/sectrack/

# Stage 2: Quarto runtime + sectrack binary
FROM ubuntu:24.04

ARG QUARTO_VERSION=1.9.38

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL "https://github.com/quarto-dev/quarto-cli/releases/download/v${QUARTO_VERSION}/quarto-${QUARTO_VERSION}-linux-amd64.deb" \
        -o /tmp/quarto.deb \
    && dpkg -i /tmp/quarto.deb \
    && rm /tmp/quarto.deb

COPY --from=builder /build/sectrack /usr/local/bin/sectrack

WORKDIR /model

ENTRYPOINT ["/usr/local/bin/sectrack"]
CMD ["--help"]
