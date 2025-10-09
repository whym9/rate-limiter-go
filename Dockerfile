# ---------- Builder ----------
FROM golang:1.24.0-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/rate-limiter

FROM gcr.io/distroless/static:nonroot
WORKDIR /app

COPY --from=builder /out/api /app/api

USER nonroot:nonroot

ENV HTTP_ADDRESS=":1234" \
    RATE_LIMIT=100 \
    WINDOW_SEC=60 \
    REDIS_ADDRESS="redis:6379"

EXPOSE 1234

ENTRYPOINT ["/app/api"]
