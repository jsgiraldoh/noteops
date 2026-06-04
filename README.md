# NoteOPs

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![CI](https://github.com/jsgiraldoh/noteops/actions/workflows/ci.yml/badge.svg)](../../actions/workflows/ci.yml)
[![Release](https://github.com/jsgiraldoh/noteops/actions/workflows/release.yml/badge.svg)](../../actions/workflows/release.yml)

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
- [Base de datos](#base-de-datos)
- [API Reference](#api-reference)
- [WebSocket — Reloj en tiempo real](#websocket--reloj-en-tiempo-real)
- [CI/CD y releases](#cicd-y-releases)
- [Equipo de desarrollo (Claude Code)](#equipo-de-desarrollo-claude-code)
- [Comandos útiles](#comandos-útiles)
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
| **Backend** | Go + Gin | 1.23 / v1.10 | Concurrencia nativa para WebSocket, binario estático de ~15 MB en imagen `scratch`, tipado fuerte en DTOs |
| **Frontend** | SvelteKit | 2.x | Compila a vanilla JS sin runtime — bundle mínimo, variables de entorno embebidas en build time |
| **Base de datos** | PostgreSQL | 16 | Modelo relacional, vista SQL para nota definitiva calculada automáticamente, `pgcrypto` para hashing de contraseñas |
| **Cache / WS** | Redis | 7 | Estado de sesiones WebSocket entre instancias del backend |
| **Archivos** | MinIO | latest | Exportes de planillas, compatible con S3, self-hosted |
| **Proxy** | Traefik | v3 | SSL automático con Let's Encrypt, routing por hostname y path, zero-config con Docker labels |
| **Contenedores** | Docker + Compose | latest | Un comando levanta todo el stack — local y producción idénticos |

---

## Arquitectura

```
┌─────────────────────────────────────────────────────────────┐
│                      Red local / Internet                    │
│           Navegador / App móvil (HTTP o HTTPS)               │
└───────────────────────┬─────────────────────────────────────┘
                        │ :80 / :443
                ┌───────▼────────┐
                │    Traefik v3   │  ← SSL, routing por path y hostname
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
```

**Flujo de una nota registrada:**

```
Docente en el navegador
  → SvelteKit (POST /api/grades)
    → Traefik (routing por path /api)
      → Go/Gin Handler (valida JWT, binding JSON)
        → Repository (UPSERT en PostgreSQL)
          → Vista student_final_grades (recalcula definitiva automáticamente)
        ← Grade struct como JSON
      ← 200 OK
    ← store de Svelte actualizado
  ← tabla de notas re-renderizada en pantalla
```

### Estructura del repositorio

```
noteops/
├── backend/                 Go + Gin
│   ├── cmd/server/          Punto de entrada (main.go)
│   ├── go.mod               Módulo Go con dependencias declaradas
│   └── internal/
│       ├── config/          Carga de variables de entorno (.env + OS)
│       ├── handlers/        HTTP handlers + WebSocket hub
│       ├── middleware/       JWT auth + request logger
│       ├── models/          Structs de dominio y DTOs
│       ├── repository/      Queries SQL con pgx (sin ORM)
│       └── service/         Lógica de negocio (slots, notas agregadas)
├── frontend/                SvelteKit + TypeScript
│   └── src/
│       ├── lib/api/         Clientes HTTP tipados (subjects, grades, sessions)
│       ├── lib/stores/      Estado reactivo (auth, clock, subject)
│       ├── lib/components/  Clock, SlotGrid, GradeCell, modales
│       └── routes/          / (notas) · /session · /students · /login
├── workers/                 Python — agentes IA (futuro)
├── infra/
│   ├── traefik/             traefik.yml — configuración del proxy
│   └── postgres/
│       ├── init.sql         Schema completo: tablas, índices, vista de nota definitiva, usuario admin
│       └── 02_seed_data.sql Datos académicos de ejemplo (excluido de git — privado)
├── .claude/                 Skills del equipo de desarrollo (Claude Code)
│   ├── CLAUDE.md
│   └── skills/              dev · qa · architect · docs · release
└── .github/workflows/       ci.yml · cd.yml · release.yml
```

---

## Inicio rápido — Local (build desde código)

**Prerequisitos:** Docker 24+ y Docker Compose v2 instalados.

### Linux / macOS

```bash
# 1. Clonar el repositorio
git clone https://github.com/jsgiraldoh/noteops.git
cd noteops

# 2. Configurar variables de entorno
cp .env.example .env
# Editar .env — al menos cambiar JWT_SECRET y DB_PASSWORD

# 3. Agregar hostname local (solo una vez)
echo "127.0.0.1  noteops.local" | sudo tee -a /etc/hosts

# 4. Levantar con build desde código fuente
docker compose --profile local up -d --build

# 5. Verificar que todo está corriendo
docker compose ps
```

### Windows

```powershell
# 1. Clonar el repositorio
git clone https://github.com/jsgiraldoh/noteops.git
cd noteops

# 2. Configurar variables de entorno
copy .env.example .env
# Editar .env — al menos cambiar JWT_SECRET y DB_PASSWORD

# 3. Agregar hostname local (abrir Notepad como administrador y editar):
#    C:\Windows\System32\drivers\etc\hosts
#    Agregar al final:  127.0.0.1  noteops.local
#
#    O desde PowerShell como administrador:
Add-Content -Path "C:\Windows\System32\drivers\etc\hosts" -Value "127.0.0.1  noteops.local"

# 4. Levantar con build desde código fuente
docker compose --profile local up -d --build
```

La aplicación estará disponible en **http://noteops.local**

**Credenciales por defecto:**

| Campo | Valor |
|---|---|
| Email | `admin@noteops.local` |
| Contraseña | `admin123` |

> Cambiar la contraseña en producción accediendo directamente a la base de datos.

> El primer build toma 2–3 minutos mientras descarga dependencias Go y Node.

### Cargar datos de un periodo académico

Si tienes un archivo `02_seed_data.sql` con datos reales (generado desde una planilla Excel), colócalo en `infra/postgres/` antes de levantar el stack. PostgreSQL lo ejecutará automáticamente en el primer inicio:

```bash
# Con datos de seed ya en infra/postgres/02_seed_data.sql
docker compose --profile local down -v   # elimina el volumen anterior
docker compose --profile local up -d --build
```

El flag `-v` es necesario para que PostgreSQL vuelva a ejecutar los scripts de inicialización desde cero.

---

## Inicio rápido — Registry (imágenes de GitHub)

Usa las imágenes pre-construidas publicadas en GHCR. No necesitas el código fuente.

```bash
# 1. Descargar solo los archivos necesarios
curl -O https://raw.githubusercontent.com/jsgiraldoh/noteops/main/docker-compose.yml
curl -O https://raw.githubusercontent.com/jsgiraldoh/noteops/main/.env.example
mkdir -p infra/postgres infra/traefik
curl -o infra/postgres/init.sql \
  https://raw.githubusercontent.com/jsgiraldoh/noteops/main/infra/postgres/init.sql
curl -o infra/traefik/traefik.yml \
  https://raw.githubusercontent.com/jsgiraldoh/noteops/main/infra/traefik/traefik.yml

# 2. Configurar entorno
cp .env.example .env   # editar .env

# 3. Hostname local (ver sección anterior)

# 4. Levantar con imágenes del registry
docker compose --profile registry up -d

# Para una versión específica:
TAG=v1.0.0 docker compose --profile registry up -d
```

### Imágenes disponibles en GHCR

```
ghcr.io/jsgiraldoh/noteops/backend:latest
ghcr.io/jsgiraldoh/noteops/frontend:latest
ghcr.io/jsgiraldoh/noteops/backend:v1.0.0
ghcr.io/jsgiraldoh/noteops/frontend:v1.0.0
```

---

## Variables de entorno

Copia `.env.example` a `.env` y ajusta los valores marcados como requeridos:

| Variable | Requerida | Default | Descripción |
|---|---|---|---|
| `DATABASE_URL` | ✅ | — | URL completa de conexión a PostgreSQL. Formato: `postgres://user:pass@host:5432/db?sslmode=disable` |
| `DB_USER` | ✅ | `noteops` | Usuario de la base de datos |
| `DB_PASSWORD` | ✅ | `secret` | **Cambiar en producción** |
| `DB_NAME` | ❌ | `noteops` | Nombre de la base de datos |
| `JWT_SECRET` | ✅ | — | Secreto para firmar tokens JWT. Mínimo 32 caracteres. Generar: `openssl rand -hex 32` |
| `REDIS_URL` | ❌ | `redis://redis:6379` | URL de conexión a Redis |
| `APP_ENV` | ❌ | `development` | `development` activa logs detallados. `production` activa modo release de Gin |
| `APP_PORT` | ❌ | `8080` | Puerto interno del backend |
| `APP_DOMAIN` | ❌ | `noteops.local` | Dominio principal usado por Traefik para el routing |
| `GITHUB_REPOSITORY` | ❌ | `jsgiraldoh/noteops` | Ruta del repositorio para construir las URLs de imágenes GHCR |
| `TAG` | ❌ | `latest` | Versión de imagen a desplegar con el perfil `registry` |
| `MINIO_ROOT_USER` | ❌ | `minioadmin` | Usuario administrador de MinIO |
| `MINIO_ROOT_PASSWORD` | ❌ | `minioadmin` | **Cambiar en producción** |
| `PUBLIC_API_URL` | ❌ | `http://noteops.local/api` | URL del API REST consumida por el frontend. **Variable de build time**: se embebe en el JS compilado en el Dockerfile — no se puede cambiar en runtime sin reconstruir la imagen |
| `PUBLIC_WS_URL` | ❌ | `ws://noteops.local` | URL del WebSocket consumida por el frontend. Misma restricción de build time que `PUBLIC_API_URL` |

> **Nota sobre `PUBLIC_*`:** SvelteKit embebe estas variables en el JavaScript compilado durante `npm run build`. Si necesitas cambiar la URL del API después del build, debes reconstruir la imagen frontend pasando los `ARG` correspondientes al `docker build`.

---

## Base de datos

El schema se aplica automáticamente al primer `docker compose up` ejecutando los archivos en `infra/postgres/` en orden alfabético:

| Archivo | Propósito |
|---|---|
| `init.sql` → montado como `01_schema.sql` | Crea todas las tablas, índices, la vista `student_final_grades` e inserta el usuario administrador por defecto |
| `02_seed_data.sql` *(opcional, excluido de git)* | Datos de un periodo académico real — estudiantes, materias, cortes, actividades y notas |

### Entidades principales

```
users → subjects → cuts → activities
students → enrollments ─┐
                         └→ grades (valor + comentario por actividad)
sessions → slots (espacios de tiempo reservables)
```

### Vista de nota definitiva

PostgreSQL calcula la nota definitiva automáticamente sin lógica en el backend:

```sql
SELECT * FROM student_final_grades WHERE subject_id = 'uuid';
-- → { enrollment_id, student_id, subject_id, final_grade: 4.75 }
```

La fórmula: `ROUND( Σ (nota × peso_actividad × peso_corte), 2 )`.

### Usuario administrador por defecto

El `init.sql` crea un usuario administrador usando `pgcrypto` para el hash de la contraseña (bcrypt, compatible con el backend Go):

```sql
INSERT INTO users (full_name, email, password, role)
VALUES ('Admin', 'admin@noteops.local', crypt('admin123', gen_salt('bf')), 'admin')
ON CONFLICT (email) DO NOTHING;
```

---

## API Reference

Todos los endpoints salvo `/api/health` y `/api/auth/login` requieren `Authorization: Bearer <token>`.

### Autenticación

```bash
# Login — devuelve JWT válido por 24 horas
curl -X POST http://noteops.local/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@noteops.local","password":"admin123"}'
# → { "token": "eyJ...", "user": { "id": "uuid", "role": "admin" } }

# Health check (sin autenticación)
curl http://noteops.local/api/health
# → { "status": "ok" }
```

### Materias

```bash
# Listar materias del docente autenticado
GET /api/subjects
# → [ { "id": "uuid", "name": "Sistemas Operativos", "period": "2025-1", ... } ]
```

### Estudiantes

```bash
# Registrar estudiante
POST /api/students
{ "full_name": "ARCE PAREJA SEBASTIAN", "email": "s.arce@uni.edu.co", "code": "240220211012" }

# Inscribir en una materia
POST /api/subjects/:id/enroll
{ "student_id": "uuid" }

# Listar estudiantes de una materia
GET /api/subjects/:id/students
```

### Notas

```bash
# Registrar o actualizar nota (upsert por enrollment_id + activity_id)
POST /api/grades
{
  "enrollment_id": "uuid",
  "activity_id": "uuid",
  "value": 4.5,
  "comment": "Buena entrega"
}

# Agregar o editar comentario de retroalimentación
PATCH /api/grades/:id/comment
{ "comment": "Mejorar documentación del código" }

# Notas completas de una materia (cortes + actividades + estudiantes)
GET /api/subjects/:id/grades
# → { "cuts": [...], "students": [...], "final_grades": [...] }

# Nota definitiva calculada por estudiante
GET /api/subjects/:id/final-grades
# → [ { "student_id": "uuid", "final_grade": 4.75 } ]
```

### Sesiones y espacios

```bash
# Crear sesión con espacios automáticos
POST /api/sessions
{
  "subject_id": "uuid",
  "starts_at": "2025-03-15T08:00:00Z",
  "duration_min": 120,
  "slot_min": 20,
  "room": "Sala 201"
}
# → { "session": {...}, "slots": [ { "number": 1, "starts_at": "..." }, ... ] }

# Activar sesión (inicia el reloj WebSocket)
POST /api/sessions/:id/activate

# Ver espacios disponibles y reservados
GET /api/sessions/:id/slots

# Reservar un espacio para un estudiante
POST /api/sessions/:id/slots/:slotId/reserve
{ "student_id": "uuid" }
```

---

## WebSocket — Reloj en tiempo real

Una vez activada una sesión, cada cliente conectado recibe un tick por segundo:

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

El componente `Clock.svelte` consume este store y muestra el reloj en pantalla grande. Cambia a amarillo en los últimos 5 minutos y a rojo cuando llega a cero.

---

## CI/CD y releases

### Flujo de ramas

```
feature/* ──PR──▶ develop ──PR──▶ main ──tag v*.*.*──▶ GHCR + GitHub Release
                    │
                 CI automático
              (lint + tests en cada PR)
```

### Workflows de GitHub Actions

| Archivo | Disparo | Qué hace |
|---|---|---|
| `ci.yml` | Pull Request | Tests Go con `-race` + type-check TypeScript + build frontend |
| `cd.yml` | Push a `main` | Build imágenes Docker + deploy SSH en servidor de producción |
| `release.yml` | Tag `v*.*.*` | Build + push a GHCR con tags semver + crea GitHub Release |

### Crear un release

```bash
# Asegurarse de estar en main con todo mergeado
git checkout main && git pull

# Crear tag y publicar — GitHub Actions construye y publica las imágenes
make release VERSION=v1.0.0
```

Publica automáticamente:
- `ghcr.io/jsgiraldoh/noteops/backend:v1.0.0` y `:latest`
- `ghcr.io/jsgiraldoh/noteops/frontend:v1.0.0` y `:latest`

### Secrets necesarios en GitHub

En `Settings → Secrets and variables → Actions`:

| Secret | Para qué |
|---|---|
| `SERVER_HOST` | IP del servidor para deploy SSH |
| `SERVER_USER` | Usuario SSH del servidor |
| `SSH_PRIVATE_KEY` | Llave privada SSH para el deploy |

El `GITHUB_TOKEN` para GHCR es automático — no necesita configuración adicional.

---

## Equipo de desarrollo (Claude Code)

El directorio `.claude/` contiene skills para Claude Code que definen el comportamiento de cada rol:

| Skill | Rol | Cuándo invocarlo |
|---|---|---|
| `skills/dev` | Developer | Crear PRs, implementar endpoints, componentes Svelte |
| `skills/qa` | QA Engineer | Tests unitarios e integración del backend Go |
| `skills/architect` | Arquitecto | Actualizar README, documentar decisiones técnicas |
| `skills/docs` | Ing. de Requisitos | Godoc, JSDoc, ADRs de arquitectura |
| `skills/release` | Release Engineer | Versionar, publicar releases, actualizar CHANGELOG |

---

## Comandos útiles

```bash
# Desarrollo
make up               # Levantar stack completo con build local
make dev              # Solo infra (DB, Redis, MinIO) — útil para desarrollo sin Docker
make logs             # Logs en tiempo real de todos los servicios
make shell-db         # Abrir psql en el contenedor de postgres

# Testing
make test             # go test ./... -race -cover + npm run check

# Producción
make deploy TAG=v1.0.0      # Desplegar versión específica desde registry
make release VERSION=v1.0.0  # Crear tag y disparar release en GitHub Actions
```

---

## Guía de contribución

1. Fork del repositorio
2. Crear rama desde `develop`: `git checkout -b feature/mi-funcionalidad`
3. Implementar cambios siguiendo las convenciones (Conventional Commits, ver `.claude/CLAUDE.md`)
4. `make test` debe pasar sin errores
5. Abrir Pull Request hacia `develop` con descripción completa

Ver [CONTRIBUTING.md](CONTRIBUTING.md) para el flujo detallado y [CHANGELOG.md](CHANGELOG.md) para el historial de versiones.

---

## Hoja de ruta

| Versión | Estado | Funcionalidad |
|---|---|---|
| **v0.1** | ✅ | Backend Go: estudiantes, materias, cortes, notas, sesiones, slots |
| **v0.2** | ✅ | Frontend SvelteKit: notas, sesión con reloj, estudiantes |
| **v0.3** | ✅ | CI/CD GitHub Actions + registry GHCR + skills Claude Code |
| **v0.4** | ✅ | Fix build Docker (GONOSUMDB), variables `PUBLIC_*` en build time, carga de materias post-login |
| **v1.0** | Pendiente | Exportar planilla Excel compatible con formato de la institución |
| **v1.1** | Pendiente | Auth completa: registro de docentes, cambio de contraseña desde el frontend |
| **v1.2** | Pendiente | Historial de cambios por nota, trazabilidad completa |
| **v2.0** | Pendiente | Workers Python: análisis de notas con agentes IA |

---

## Licencia

```
Copyright 2025 Johan Sebastian Giraldo Hurtado

Licensed under the Apache License, Version 2.0.
See LICENSE file for full terms.
```
