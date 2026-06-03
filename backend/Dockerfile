# ── Etapa dev: hot reload con air ────────────────────────────────────────────
FROM golang:1.23-alpine AS dev

WORKDIR /app

RUN apk add --no-cache git && \
    go install github.com/air-verse/air@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["air", "-c", ".air.toml"]

# ── Etapa build: compilar binario estático ────────────────────────────────────
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /noteops-backend \
    ./cmd/server

# ── Etapa producción: imagen mínima ──────────────────────────────────────────
FROM scratch AS prod

COPY --from=builder /noteops-backend /noteops-backend
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8080

ENTRYPOINT ["/noteops-backend"]
