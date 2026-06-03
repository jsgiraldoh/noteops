.DEFAULT_GOAL := help

REPO := johansgiraldo/noteops
TAG  ?= latest

.PHONY: help up up-registry dev down logs test migrate seed build push release deploy shell-db ps

help: ## Mostrar esta ayuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

# ── Entorno ──────────────────────────────────────────────────
up: ## Levantar con build local desde código fuente
	docker compose --profile local up -d

up-registry: ## Levantar usando imágenes de GHCR (TAG=latest por defecto)
	TAG=$(TAG) docker compose --profile registry up -d

dev: ## Desarrollo: solo infra (DB, Redis, MinIO) — corre backend/frontend en local
	docker compose up -d postgres redis minio

down: ## Apagar todos los servicios
	docker compose --profile local --profile registry down

logs: ## Logs en tiempo real
	docker compose logs -f

ps: ## Estado de los contenedores
	docker compose ps

shell-db: ## Abrir psql
	docker compose exec postgres psql -U noteops -d noteops

# ── Calidad ──────────────────────────────────────────────────
test: ## Correr tests backend + check frontend
	cd backend && go test ./... -race -cover
	cd frontend && npm run check

migrate: ## Aplicar migraciones
	docker compose exec noteops_backend ./migrate up

seed: ## Cargar datos de ejemplo
	docker compose exec noteops_backend ./seed

# ── Imágenes ─────────────────────────────────────────────────
build: ## Build local de imágenes
	docker build -t ghcr.io/$(REPO)/backend:$(TAG) ./backend
	docker build -t ghcr.io/$(REPO)/frontend:$(TAG) ./frontend

push: ## Push imágenes al registry (requiere docker login ghcr.io)
	docker push ghcr.io/$(REPO)/backend:$(TAG)
	docker push ghcr.io/$(REPO)/frontend:$(TAG)

pull: ## Descargar imágenes del registry
	docker pull ghcr.io/$(REPO)/backend:$(TAG)
	docker pull ghcr.io/$(REPO)/frontend:$(TAG)

# ── Release ──────────────────────────────────────────────────
release: ## Crear release — uso: make release VERSION=v1.0.0
	@test -n "$(VERSION)" || (echo "❌  Uso: make release VERSION=v1.0.0" && exit 1)
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)
	@echo "✅  Tag $(VERSION) publicado — GitHub Actions construirá las imágenes"

deploy: ## Desplegar versión — uso: make deploy TAG=v1.0.0
	@test -n "$(TAG)" || (echo "❌  Uso: make deploy TAG=v1.0.0" && exit 1)
	TAG=$(TAG) docker compose --profile registry pull
	TAG=$(TAG) docker compose --profile registry up -d --no-deps --force-recreate \
		backend-registry frontend-registry
	@echo "✅  Desplegado TAG=$(TAG)"
