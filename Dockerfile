# syntax=docker/dockerfile:1
# Targets linux/amd64. For other architectures adjust the lib copy paths in the runtime stage.

# ── Stage 1: build static Go binary ──────────────────────────────────────────
FROM golang:1.26-bookworm AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o gtrmnl .

# ── Stage 2: install Chromium and fonts ───────────────────────────────────────
FROM debian:bookworm-slim AS chromium
RUN apt-get update && apt-get install -y --no-install-recommends \
        chromium \
        fonts-liberation \
        fonts-noto-core \
    && rm -rf /var/lib/apt/lists/* \
    && fc-cache -fv

# ── Stage 3: distroless runtime ───────────────────────────────────────────────
FROM gcr.io/distroless/base-debian12

# Shared libraries Chromium depends on (not already provided by distroless/base)
COPY --from=chromium /lib/x86_64-linux-gnu     /lib/x86_64-linux-gnu
COPY --from=chromium /usr/lib/x86_64-linux-gnu /usr/lib/x86_64-linux-gnu

# Chromium resources and binary.
# The binary lives at /usr/lib/chromium/chromium; /usr/bin/chromium in Debian
# is a shell wrapper that distroless cannot execute, so we use ExecPath instead.
COPY --from=chromium /usr/lib/chromium /usr/lib/chromium

# Fonts + pre-built fontconfig cache
COPY --from=chromium /usr/share/fonts      /usr/share/fonts
COPY --from=chromium /etc/fonts            /etc/fonts
COPY --from=chromium /var/cache/fontconfig /var/cache/fontconfig

# App binary
COPY --from=builder /build/gtrmnl /gtrmnl

USER nonroot:nonroot
ENV HOME=/tmp
ENTRYPOINT ["/gtrmnl"]
