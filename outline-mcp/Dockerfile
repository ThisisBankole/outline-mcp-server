# ── Build stage ───────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o outline-mcp ./cmd/server

# ── Run stage ─────────────────────────────────────────────────────────────────
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/outline-mcp .

EXPOSE 9292

CMD ["./outline-mcp"]