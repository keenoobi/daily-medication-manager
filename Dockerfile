# Stage 1: Build
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /medication-scheduler ./cmd/server/

# Stage 2: Run
FROM alpine:3.18
WORKDIR /app
COPY --from=builder /medication-scheduler /app/medication-scheduler
COPY --from=builder /app/migrations /app/migrations
CMD ["./medication-scheduler"]