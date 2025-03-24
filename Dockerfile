# Stage 1: Build
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o /medication-scheduler ./cmd/server/

# Stage 2: Run
FROM scratch
WORKDIR /app
COPY --from=builder /medication-scheduler /app/medication-scheduler
COPY --from=builder /app/migrations /app/migrations
CMD ["./medication-scheduler"]