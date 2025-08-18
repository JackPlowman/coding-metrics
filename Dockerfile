FROM golang:1.25.0-alpine AS builder

WORKDIR /app

# Copy module manifest first to leverage layer caching
COPY go.mod ./

# Pre-download modules (none yet, but keeps cache structure for future deps)
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download || true

# Copy source and build
COPY src ./src
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/coding-metrics ./src

FROM alpine:3.22 AS runner

RUN adduser -D -H -u 10001 appuser

COPY --from=builder /bin/coding-metrics /usr/local/bin/coding-metrics

USER appuser
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD ["pidof", "coding-metrics"]
CMD [ "/usr/local/bin/coding-metrics" ]
