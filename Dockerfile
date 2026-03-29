# ============================================================
# Stage 1: Build
# ============================================================
# We use a "multi-stage build". Stage 1 compiles the Go binary.
# Stage 2 copies ONLY the binary into a tiny image.
#
# Why multi-stage?
#   - Go build toolchain is ~500MB. Our app binary is ~15MB.
#   - If we use a single stage, the final image contains the entire
#     Go SDK, source code, build cache — none of which are needed
#     at runtime. That's wasted disk, slower pulls, bigger attack surface.
#   - Multi-stage: final image is ~20MB instead of ~500MB.
#
# Why golang:1.25-alpine instead of golang:1.25?
#   - Alpine Linux is ~5MB vs ~100MB for Debian-based.
#   - Smaller base = faster CI builds, faster image pulls.
#   - Trade-off: Alpine uses musl libc instead of glibc.
#     For pure Go apps (no CGO), this doesn't matter.
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum FIRST, then download dependencies.
# This is a Docker layer caching trick:
#   - Docker caches each layer. If a layer's input hasn't changed,
#     Docker reuses the cached version instead of rebuilding.
#   - go.mod/go.sum rarely change (only when you add/update deps).
#   - By copying them separately, `go mod download` is cached
#     across builds where only source code changed.
#   - Without this trick, every code change would re-download
#     all dependencies — slow and wasteful.
COPY go.mod go.sum ./
RUN go mod download

# Now copy the rest of the source code.
# This layer is invalidated on every code change (expected).
COPY . .

# Build the binary.
#   CGO_ENABLED=0: compile a fully static binary (no C library dependency).
#     This is critical for Alpine/scratch images which may not have glibc.
#   -ldflags="-s -w": strip debug info and symbol table.
#     Reduces binary size by ~30% (15MB → 10MB). We don't need debug
#     symbols in production — if we need to debug, we attach a debugger
#     to the dev build, not the production image.
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/server ./cmd/api

# ============================================================
# Stage 2: Runtime
# ============================================================
# The final image contains ONLY the compiled binary.
#
# Why alpine instead of scratch?
#   - scratch is literally empty — no shell, no ls, no curl, nothing.
#     Great for security, terrible for debugging.
#   - Alpine gives us a shell and basic tools (~5MB overhead)
#     so we can exec into the container to troubleshoot if needed.
#   - For learning projects, debuggability > minimal image size.
#   - In production, consider distroless or scratch for maximum security.
FROM alpine:3.22

# Install CA certificates so the app can make HTTPS requests
# (e.g., if you ever need to call external APIs, webhook, etc).
# Without this, TLS connections fail with "x509: certificate not found".
RUN apk add --no-cache ca-certificates

# Run as non-root user.
# By default, Docker runs processes as root inside the container.
# If the app has a vulnerability that allows code execution,
# the attacker gets root access to the container.
# Running as a non-root user limits the blast radius.
RUN adduser -D -u 1000 appuser
USER appuser

WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 8080

# Use exec form (JSON array) instead of shell form.
# Shell form: CMD server        → runs as /bin/sh -c server
# Exec form:  CMD ["./server"]  → runs the binary directly
#
# Why exec form? Because the binary receives signals (SIGTERM)
# directly. With shell form, sh receives the signal, not your app,
# and your graceful shutdown code never runs.
CMD ["./server"]