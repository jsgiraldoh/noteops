# NoteOPs

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![CI](https://github.com/johansgiraldo/noteops/actions/workflows/ci.yml/badge.svg)](../../actions/workflows/ci.yml)
[![Release](https://github.com/johansgiraldo/noteops/actions/workflows/release.yml/badge.svg)](../../actions/workflows/release.yml)

> Sistema open source de gestión de notas académicas con reloj de clase en tiempo real, reserva de espacios y cálculo automático de nota definitiva.

**Autor:** Johan Sebastian Giraldo Hurtado · **Licencia:** Apache 2.0

---

## Tabla de contenidos

- [¿Qué es NoteOPs?](#qué-es-noteops)
- [Stack técnico](#stack-técnico)
- [Arquitectura](#arquitectura)
- [Inicio rápido — Local (build desde código)](#inicio-rápido--local-build-desde-código)
- [Inicio rápido — Registry (imágenes de GitHub)](#inicio-rápido--registry-imágenes-de-github)
- [Variables de entorno](#variables-de-entorno)
- [API Reference](#api-reference)
- [WebSocket — Reloj en tiempo real](#websocket--reloj-en-tiempo-real)
- [Base de datos](#base-de-datos)
- [CI/CD y releases](#cicd-y-releases)
- [Equipo de desarrollo (Claude Code)](#equipo-de-desarrollo-claude-code)
- [Guía de contribución](#guía-de-contribución)
- [Hoja de ruta](#hoja-de-ruta)

---

## ¿Qué es NoteOPs?

NoteOPs digitaliza el proceso de registro y seguimiento de notas académicas universitarias. Reemplaza las planillas Excel con una interfaz web colaborativa que permite:

- **Registrar notas** por corte y actividad, con pesos configurables por materia
- **Calcular automáticamente** la nota definitiva usando una vista SQL en tiempo real
- **Ver un reloj grande** durante la clase que muestra el tiempo restante de la sesión
- **Reservar espacios** de 5, 10, 20 minutos o la sesión completa para grupos de estudiantes
- **Agregar reflexiones** y comentarios de retroalimentación por cada nota
- **Ejecutarse en red local** (aula de clase) o en un servidor en la nube

---

## Stack técnico

| Capa | Tecnología | Versión | Por qué |
|---|---|---|---|
| **Backend** | Go + Gin | 1.22 / v1.10 | Concurrencia nativa para WebSocket, binario de ~15MB, tipado fuerte |
| **Frontend** | SvelteKit | 2.x | Compila a vanilla JS — sin runtime que se deprece, bundle mínimo |
| **Base de datos** | PostgreSQL | 16 | Modelo relacional, vista SQL para nota definitiva, soporte JSONB futuro |
| **Cache / WS** | Redis | 7 | Estado de sesiones WebSocket entre instancias |
| **Archivos** | MinIO | latest | Exportes de planillas, compatible con S3, self-hosted |
| **Proxy** | Traefik | v3 | SSL automático, routing por hostname, zero-config |
| **Contenedores** | Docker + Compose | latest | Un comando levanta todo — local y producción |

---

## Arquitectura

```
┌─────────────────────────────────────────────────────────────┐
│                      Red local / Internet                    │
│           Navegador / App móvil (HTTP o HTTPS)               │
└───────────────────────┬─────────────────────────────────────┘
                        │
                ┌───────▼────────┐
                │    Traefik v3   │  ← SSL, routing por path
                │  Reverse Proxy  │
                └──┬─────────┬───┘
                   │         │
          ┌────────▼──┐  ┌───▼────────┐
          │ Frontend   │  │  Backend   │
          │ SvelteKit  │  │  Go + Gin  │
          │ :3000      │  │  :8080     │
          └────────────┘  └──┬──────┬─┘
                             │      │ WebSocket
                    ┌────────▼─┐  ┌─▼──────────┐
                    │PostgreSQL│  │   Redis     │
                    │  :5432   │  │   :6379     │
                    └──────────┘  └────────────┘

Flujo de una nota registrada:
  Docente → SvelteKit (POST /api/grades)
    → Traefik → Go Handler (valida JWT)
      → Repository (INSERT PostgreSQL)
        → Vista student_final_grades (recalcula definitiva)
      ← Grade JSON
    ← 200 OK → tabla actualizada en pantalla
```

### Estructura del repositorio

```
noteops/
├── backend/                 Go + Gin
│   ├── cmd/server/          Punto de entrada (main.go)
│   └── internal/
│       ├── config/          Carga de variables de entorno
│       ├── handlers/        HTTP handlers + WebSocket hub
│       ├── middleware/       JWT auth
│       ├── models/          Structs y DTOs
│       ├── repository/      Queries SQL (pgx, sin ORM)
│       └── service/         Lógica de negocio
├── frontend/                SvelteKit + TypeScript
│   └── src/
│       ├── lib/api/         Clientes HTTP tipados
│       ├── lib/stores/      Estado reactivo (auth, clock, subject)
│       ├── lib/components/  Clock, SlotGrid, GradeCell, modales
│       └── routes/          / (notas) · /session · /students · /login
├── workers/                 Python — agentes IA (futuro)
├── infra/
│   ├── traefik/             traefik.yml
│   └── postgres/            init.sql — schema completo
├── .claude/                 Skills del equipo de desarrollo
│   ├── CLAUDE.md
│   ├── decisions/           ADRs de arquitectura
│   └── skills/              dev · qa · architect · docs · release
└── .github/workflows/       ci.yml · cd.yml · release.yml
```

---

## Inicio rápido — Local (build desde código)

**Prerequisitos:** Docker 24+ y Docker Compose v2 instalados.

```bash
# 1. Clonar el repositorio
git clone https://github.com/johansgiraldo/noteops.git
cd noteops

# 2. Configurar variables de entorno
cp .env.example .env
# Editar .env — al menos cambiar JWT_SECRET y DB_PASSWORD

# 3. Agregar hostname local (solo una vez)
echo "127.0.0.1  noteops.local" | sudo tee -a /etc/hosts

# 4. Levantar con build desde código fuente
docker compose --profile local up -d

# 5. Verificar que todo está corriendo
docker compose ps
```

La aplicación estará disponible en **http://noteops.local**

> El primer build toma 2-3 minutos mientras descarga dependencias Go y Node.

---

## Inicio rápido — Registry (imágenes de GitHub)

Usa las imágenes pre-construidas publicadas en GHCR. No necesitas el código fuente — solo el `docker-compose.yml` y el `.env`.

```bash
# 1. Descargar solo los archivos necesarios
curl -O https://raw.githubusercontent.com/johansgiraldo/noteops/main/docker-compose.yml
curl -O https://raw.githubusercontent.com/johansgiraldo/noteops/main/.env.example
curl -O https://raw.githubusercontent.com/johansgiraldo/noteops/main/infra/postgres/init.sql

# Crear la carpeta que espera Docker Compose
mkdir -p infra/postgres
mv init.sql infra/postgres/
mkdir -p infra/traefik
curl -o infra/traefik/traefik.yml \
  https://raw.githubusercontent.com/johansgiraldo/noteops/main/infra/traefik/traefik.yml

# 2. Configurar entorno
cp .env.example .env
# Editar .env

# 3. Hostname local
echo "127.0.0.1  noteops.local" | sudo tee -a /etc/hosts

# 4. Autenticarse en GHCR (necesario para imágenes privadas)
echo $GITHUB_TOKEN | docker login ghcr.io -u TU_USUARIO --password-stdin

# 5. Levantar con imágenes del registry
docker compose --profile registry up -d

# Para una versión específica:
TAG=v1.0.0 docker compose --profile registry up -d
```

### Imágenes disponibles en GHCR

```bash
# Siempre la última versión estable
ghcr.io/johansgiraldo/noteops/backend:latest
ghcr.io/johansgiraldo/noteops/frontend:latest

# Versión específica
ghcr.io/johansgiraldo/noteops/backend:v1.0.0
ghcr.io/johansgiraldo/noteops/frontend:v1.0.0
```

---

## Variables de entorno

Copia `.env.example` a `.env` y ajusta los valores:

| Variable | Requerida | Default | Descripción |
|---|---|---|---|
| `DATABASE_URL` | ✅ | — | URL completa de PostgreSQL |
| `DB_USER` | ✅ | `noteops` | Usuario de la base de datos |
| `DB_PASSWORD` | ✅ | `secret` | **Cambiar en producción** |
| `DB_NAME` | ❌ | `noteops` | Nombre de la base de datos |
| `JWT_SECRET` | ✅ | — | Mínimo 32 caracteres aleatorios. Generar con: `openssl rand -hex 32` |
| `REDIS_URL` | ❌ | `redis://redis:6379` | URL de Redis |
| `APP_ENV` | ❌ | `development` | `development` o `production` |
| `APP_PORT` | ❌ | `8080` | Puerto del backend |
| `APP_DOMAIN` | ❌ | `noteops.local` | Dominio principal |
| `GITHUB_REPOSITORY` | ❌ | `johansgiraldo/noteops` | Para imágenes GHCR |
| `TAG` | ❌ | `latest` | Versión de imagen a desplegar |
| `MINIO_ROOT_USER` | ❌ | `minioadmin` | Usuario de MinIO |
| `MINIO_ROOT_PASSWORD` | ❌ | `minioadmin` | **Cambiar en producción** |
| `PUBLIC_API_URL` | ❌ | `http://noteops.local/api` | URL del API (usada por el frontend) |
| `PUBLIC_WS_URL` | ❌ | `ws://noteops.local` | URL WebSocket (usada por el frontend) |

---

## API Reference

Todos los endpoints protegidos requieren `Authorization: Bearer <token>`.

### Autenticación

```bash
# Login
curl -X POST http://noteops.local/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"docente@uni.edu.co","password":"secreto"}'
# → { "token": "eyJ...", "user": { "id": "...", "role": "teacher" } }
```

### Estudiantes

```bash
# Registrar estudiante
POST /api/students
{ "full_name": "ARCE PAREJA SEBASTIAN", "email": "s.arce@uni.edu.co", "code": "240220211012" }

# Inscribir en materia
POST /api/subjects/:id/enroll
{ "student_id": "uuid" }

# Listar por materia
GET /api/subjects/:id/students
```

### Notas

```bash
# Registrar o actualizar nota (upsert)
POST /api/grades
{ "enrollment_id": "uuid", "activity_id": "uuid", "value": 4.5, "comment": "Buena entrega" }

# Agregar reflexión a una nota
PATCH /api/grades/:id/comment
{ "comment": "Mejorar documentación del código" }

# Notas completas de una materia (cortes + actividades + estudiantes)
GET /api/subjects/:id/grades

# Nota definitiva calculada por estudiante
GET /api/subjects/:id/final-grades
```

### Sesiones y espacios

```bash
# Crear sesión con espacios automáticos
POST /api/sessions
{ "subject_id": "uuid", "starts_at": "2025-03-15T08:00:00Z",
  "duration_min": 120, "slot_min": 20, "room": "Sala 201" }
# → { "session": {...}, "slots": [ {number:1, starts_at:...}, ... ] }

# Activar sesión (inicia el reloj)
POST /api/sessions/:id/activate

# Ver espacios disponibles
GET /api/sessions/:id/slots

# Reservar un espacio
POST /api/sessions/:id/slots/:slotId/reserve
{ "student_id": "uuid" }
```

### Health check

```bash
GET /api/health
# → { "status": "ok" }
```

---

## WebSocket — Reloj en tiempo real

Cada cliente conectado a una sesión activa recibe un tick por segundo:

```
ws://noteops.local/ws/session/:session_id
```

**Payload JSON:**

```json
{
  "session_id": "uuid",
  "elapsed_sec": 1800,
  "remaining_sec": 5400,
  "duration_min": 120,
  "is_active": true
}
```

El componente `Clock.svelte` consume este store y muestra el reloj grande. Cambia a amarillo en los últimos 5 minutos y a rojo al llegar a cero.

---

## Base de datos

El schema completo está en `infra/postgres/init.sql` y se aplica automáticamente al primer `docker compose up`.

**Entidades principales:**

```
users → subjects → cuts → activities
students → enrollments ─┐
                         └→ grades (valor + comentario por actividad)
sessions → slots (espacios de tiempo reservables)
```

**Vista automática de nota definitiva:**

```sql
-- No requiere lógica en el backend — PostgreSQL la calcula
SELECT * FROM student_final_grades WHERE subject_id = '...';
-- → { enrollment_id, student_id, subject_id, final_grade: 4.75 }
```

La fórmula: `Σ (nota × peso_actividad × peso_corte)` redondeado a 2 decimales.

---

## CI/CD y releases

### Flujo de trabajo

```
feature/* ──PR──▶ develop ──PR──▶ main ──tag v*.*.*──▶ GHCR + GitHub Release
                    │
                 CI automático
              (lint + test en cada PR)
```

### Workflows de GitHub Actions

| Archivo | Disparo | Qué hace |
|---|---|---|
| `ci.yml` | Pull Request | Tests Go + check TypeScript + build frontend |
| `cd.yml` | Push a `main` | Build imágenes Docker + deploy en servidor |
| `release.yml` | Tag `v*.*.*` | Build + push a GHCR con tags semver + crea GitHub Release |

### Crear un release

```bash
# Asegurarse de estar en main con todo mergeado
git checkout main && git pull

# Crear y publicar el tag — GitHub Actions hace el resto
make release VERSION=v1.0.0
```

Esto construye y publica automáticamente:
- `ghcr.io/johansgiraldo/noteops/backend:v1.0.0` y `:latest`
- `ghcr.io/johansgiraldo/noteops/frontend:v1.0.0` y `:latest`

### Secrets necesarios en GitHub

Ve a `Settings → Secrets and variables → Actions`:

| Secret | Para qué |
|---|---|
| `SERVER_HOST` | IP del servidor para deploy SSH |
| `SERVER_USER` | Usuario SSH del servidor |
| `SSH_PRIVATE_KEY` | Llave privada SSH para el deploy |

El token `GITHUB_TOKEN` para GHCR es automático — no necesita configuración.

---

## Equipo de desarrollo (Claude Code)

El directorio `.claude/` contiene skills para Claude Code que definen el comportamiento de cada rol del equipo. En la raíz del proyecto, Claude Code lee el contexto automáticamente.

| Skill | Rol | Cuándo invocarlo |
|---|---|---|
| `skills/dev` | Developer | Crear PRs, implementar endpoints, componentes Svelte |
| `skills/qa` | QA Engineer | Tests unitarios e integración del backend Go |
| `skills/architect` | Arquitecto | Actualizar README, documentar decisiones técnicas |
| `skills/docs` | Ing. de Requisitos | Godoc, JSDoc, ADRs de arquitectura |
| `skills/release` | Release Engineer | Versionar, publicar releases, actualizar CHANGELOG |

**Uso en Claude Code:**

```
"Actúa como el dev y crea el endpoint de exportación Excel"
"Actúa como el QA y escribe los tests del cálculo de nota definitiva"
"Actúa como el release engineer y prepara la versión v1.1.0"
```

Los ADRs (decisiones de arquitectura) están en `.claude/decisions/`.

---

## Comandos útiles

```bash
# Desarrollo
make dev              # Hot reload — build local con docker compose
make logs             # Logs en tiempo real de backend y frontend
make shell-db         # Abrir psql en el contenedor

# Testing
make test             # go test ./... + npm run check

# Base de datos
make migrate          # Aplicar migraciones pendientes
make seed             # Cargar datos de ejemplo

# Producción
make up               # Levantar con perfil local
make deploy TAG=v1.0.0  # Desplegar versión específica desde registry
make release VERSION=v1.0.0  # Crear tag y disparar release
```

---

## Guía de contribución

1. Fork del repositorio
2. Crear rama desde `develop`: `git checkout -b feature/mi-funcionalidad`
3. Hacer cambios siguiendo las convenciones del proyecto
4. `make test` debe pasar
5. Pull Request hacia `develop` con descripción completa

Ver [CONTRIBUTING.md](CONTRIBUTING.md) para el flujo detallado y [CHANGELOG.md](CHANGELOG.md) para el historial de versiones.

---

## Hoja de ruta

| Versión | Funcionalidad |
|---|---|
| **v0.1** | ✅ Backend Go completo: estudiantes, materias, cortes, notas, sesiones, slots |
| **v0.2** | ✅ Frontend SvelteKit: notas, sesión con reloj, estudiantes |
| **v0.3** | ✅ CI/CD GitHub Actions + registry GHCR + skills Claude Code |
| **v1.0** | Exportar planilla Excel compatible con formato original |
| **v1.1** | Auth completa: registro de docentes, cambio de contraseña |
| **v1.2** | Historial de cambios por nota, trazabilidad completa |
| **v2.0** | Workers Python: análisis de notas con agentes IA |

---

## Licencia

```
Copyright 2025 Johan Sebastian Giraldo Hurtado

Licensed under the Apache License, Version 2.0.
See LICENSE file for full terms.
```
