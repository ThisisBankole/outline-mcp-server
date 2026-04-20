# ── Build stage ───────────────────────────────────────────────────────────────
FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o outline-mcp ./cmd/server

# ── Run stage ─────────────────────────────────────────────────────────────────
FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/outline-mcp .

EXPOSE 9292

CMD ["./outline-mcp"]