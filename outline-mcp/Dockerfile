FROM golang:1.22 AS builder

WORKDIR /app

COPY . .
RUN CGO_ENABLED=0 go build -mod=vendor -o outline-mcp ./cmd/server

FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/outline-mcp .

EXPOSE 9292

CMD ["./outline-mcp"]