# Stage 1: build the trustward CLI
FROM golang:1.26-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN go build -o trustward ./cmd/trustward/

# Stage 2: Quarto runtime + trustward binary
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

COPY --from=builder /build/trustward /usr/local/bin/trustward

WORKDIR /model

# Don't run as root. ubuntu:24.04 ships a 'ubuntu' user at uid 1000; trustward.sh
# overrides this with the host uid at runtime so mounted output is owned by you.
USER ubuntu

ENTRYPOINT ["/usr/local/bin/trustward"]
CMD ["--help"]
