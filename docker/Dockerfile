# Build stage (runs natively on build machine)
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

WORKDIR /build

# Install dependencies (runs natively)
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Pass build target arguments from buildx
ARG TARGETOS
ARG TARGETARCH

# Build binary for the target architecture
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-arm64} go build \
    -ldflags="-w -s" \
    -o /build/api \
    ./cmd/api

# Runtime stage (Target architecture)
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
# NOTE: This requires qemu-user-static installed on the host system if building for arm64 on amd64
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /build/api /app/api

# Copy migrations (embedded, but keeping for reference)
COPY --from=builder /build/migrations /app/migrations

# Create non-root user
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run
ENTRYPOINT ["/app/api"]
