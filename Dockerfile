# Stage 1: Build
FROM golang:1.24.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build the application
RUN go build -ldflags="-s -w" -o main ./cmd/app

RUN mkdir -p /app/config-dist && \
    ([ -f .env ] && cp .env /app/config-dist/ || true)

# Stage 2: Final Image
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/views ./views
COPY --from=builder /app/static ./static
COPY --from=builder /app/proto ./proto
COPY --from=builder /app/config-dist/ .

EXPOSE 80
CMD ["./main"]
