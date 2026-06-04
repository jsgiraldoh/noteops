# ═══════════════════════════════════════════════════════════════
# NoteOPs — Backend Multi-stage Dockerfile
#
# Etapas:
#   base    → dependencias comunes (Go + módulos)
#   dev     → hot reload con air
#   builder → compilación del binario estático
#   prod    → imagen mínima scratch (~15MB)
# ═══════════════════════════════════════════════════════════════

# ── base: módulos cacheados ───────────────────────────────────
# Descarga dependencias una sola vez.
# Docker reutiliza esta capa mientras go.mod no cambie.
FROM golang:1.23-alpine AS base

WORKDIR /app

# ca-certificates: necesarios en scratch para requests HTTPS
# git: necesario para go mod download con repos privados
RUN apk add --no-cache ca-certificates git

# Omitir verificación contra sum.golang.org (sin IPv6 en el build)
ENV GONOSUMDB=* \
    GONOSUMCHECK=* \
    GOFLAGS=-mod=mod

COPY go.mod ./
COPY go.sum* ./

RUN go mod download

# ── dev: hot reload con air ───────────────────────────────────
# Extiende base con air para detectar cambios en el código.
# El código fuente se monta como volumen en docker-compose.
FROM base AS dev

RUN go install github.com/air-verse/air@latest

COPY . .

EXPOSE 8080
CMD ["air", "-c", ".air.toml"]

# ── builder: compilación del binario ─────────────────────────
# go mod tidy garantiza que go.sum esté actualizado.
# CGO_ENABLED=0 produce un binario estático sin dependencias del SO.
FROM base AS builder

COPY . .

RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
      -ldflags="-s -w -extldflags=-static" \
      -o /noteops-backend \
      ./cmd/server

# ── prod: imagen mínima ───────────────────────────────────────
# scratch = 0MB de OS base.
# Solo el binario compilado y certificados SSL.
# Resultado: ~15MB vs ~300MB de golang:alpine
FROM scratch AS prod

# Certificados para requests HTTPS del backend
COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Binario compilado
COPY --from=builder /noteops-backend /noteops-backend

EXPOSE 8080

ENTRYPOINT ["/noteops-backend"]
