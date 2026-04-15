# ── Build stage ──
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod ./
COPY go.sum* ./
RUN go mod download || true

COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

# ── Runtime stage ──
FROM alpine:3.20

RUN apk add --no-cache ca-certificates curl postgresql-client dos2unix

# Install golang-migrate at build time (no outbound network needed at startup)
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.1/migrate.linux-amd64.tar.gz | tar xvz -C /usr/local/bin/

WORKDIR /app

COPY --from=builder /server /app/server
COPY migrations/ /app/migrations/
COPY seed.sql /app/seed.sql
COPY entrypoint.sh /app/entrypoint.sh

RUN dos2unix /app/entrypoint.sh && chmod +x /app/entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/app/entrypoint.sh"]
