# ── Build stage ──────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Dipendenze
COPY go.mod go.sum ./
RUN go mod download

# Codice sorgente
COPY . .

# Compila
RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go

# ── Run stage ─────────────────────────────────────────────────────
FROM alpine:3.19

WORKDIR /app

# Certificati SSL (per future chiamate HTTPS outbound)
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/server .
COPY .env.dist ./.env

EXPOSE 8080

CMD ["./server"]
