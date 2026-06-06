<div align="center">

# NoteOPs

### El sistema de notas donde el estudiante demuestra que entiende — antes de sustentar.

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![CI](https://github.com/jsgiraldoh/noteops/actions/workflows/ci.yml/badge.svg)](../../actions/workflows/ci.yml)
[![Release](https://github.com/jsgiraldoh/noteops/actions/workflows/release.yml/badge.svg)](../../actions/workflows/release.yml)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](#guía-de-contribución)

[![Go](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go&logoColor=white)](#stack-técnico)
[![SvelteKit](https://img.shields.io/badge/SvelteKit-2.x-FF3E00?logo=svelte&logoColor=white)](#stack-técnico)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?logo=postgresql&logoColor=white)](#stack-técnico)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker&logoColor=white)](#inicio-rápido)

**Gestión de notas académicas con reloj de clase en tiempo real y reserva de turnos por API.**

Hecho con propósito educativo por **Johan Sebastian Giraldo Hurtado** · Licencia **Apache 2.0**

</div>

---

## La aplicación de un vistazo

Durante la clase, el reloj corre en pantalla grande y los estudiantes reservan su turno para exponer. El docente administra todo desde una interfaz limpia, en tiempo real.

<div align="center">

### ⏱️ El reloj de clase, en vivo

![Sesión activa con el reloj en cuenta regresiva](images/6.png)
*Una sesión activa: el tiempo restante de la clase, grande y visible para todos.*

</div>

<table>
<tr>
<td width="50%">

![Turno reservado](images/7.png)
**Turno reservado.** Cuando un estudiante separa su espacio, el turno se marca como *Ocupado* al instante.

</td>
<td width="50%">

![Tabla de notas con cálculo automático](images/2.png)
**Notas que se calculan solas.** Registra por corte y actividad; la nota definitiva se recalcula automáticamente.

</td>
</tr>
<tr>
<td width="50%">

![Configurar sesión](images/4.png)
**Configura tu clase.** Define la duración total y el tamaño de cada turno: 5, 10, 20, 30, 60 o 120 minutos.

</td>
<td width="50%">

![Espacios de la sesión](images/5.png)
**Turnos listos para reservar.** NoteOPs genera los espacios automáticamente según tu configuración.

</td>
</tr>
</table>

<details>
<summary><b>Ver más capturas</b> — login, estudiantes, materias e importación de planillas</summary>

<br>

<table>
<tr>
<td width="50%">

![Login](images/1.png)
**Acceso del docente.**

</td>
<td width="50%">

![Sesión de clase](images/3.png)
**Tu espacio de sesión**, listo para arrancar.

</td>
</tr>
<tr>
<td width="50%">

![Gestión de estudiantes](images/8.png)
**Estudiantes** — registra e inscribe en segundos.

</td>
<td width="50%">

![Gestión de materias](images/9.png)
**Materias** — cada una con su ID para compartir con la clase.

</td>
</tr>
<tr>
<td width="50%">

![Importar planilla](images/10.png)
**Importa tu Excel** con arrastrar y soltar.

</td>
<td width="50%">

![Preview de importación](images/11.png)
**Detección automática** de materias, estudiantes, actividades y notas.

</td>
</tr>
</table>

</details>

---

## ¿Por qué NoteOPs?

Hoy existe un problema silencioso en las aulas: muchos estudiantes ya **no repasan ni interiorizan los conceptos**. Le piden a la inteligencia artificial que haga el trabajo por ellos, lo entregan, y el aprendizaje nunca ocurre. El conocimiento se delega — no se adquiere.

**NoteOPs le da la vuelta a eso.**

La idea es simple y poderosa: para poder sustentar su trabajo, el estudiante primero tiene que **separar su turno haciendo él mismo una petición HTTP a la API**. Con un `curl` desde la terminal o desde la web, pero hecho por él. Ese pequeño gesto cambia todo:

- 🧠 **Antes de exponer, ya está aprendiendo.** Entiende qué es una API, un endpoint, una petición, un identificador. Conocimiento técnico real, no delegado.
- 🎯 **La sustentación deja de ser un trámite** y se vuelve un reto práctico, técnico y hasta divertido.
- 👩‍🏫 **El docente administra; el estudiante demuestra.** Reservar el turno es la primera prueba de que entiende; sustentar el trabajo es la segunda.

> **Aprender haciendo.** En vez de pedirle a la IA que resuelva, el estudiante ejecuta, observa y comprende. NoteOPs convierte un requisito administrativo —apartar un turno— en una oportunidad de aprendizaje genuina.

Y para el docente, además, es la herramienta que reemplaza la frágil planilla de Excel: cálculo automático de la nota definitiva, reloj de clase en tiempo real e importación de su planilla existente en segundos.

---

## Empieza ahora

<div align="center">

**¿Eres docente?** Levanta NoteOPs en tu máquina en menos de 5 minutos → [Inicio rápido](#inicio-rápido)

**¿Eres estudiante?** Anímate al reto: aprende a separar tu turno con una petición real → [Cómo un estudiante separa su turno](#cómo-un-estudiante-separa-su-turno)

**¿Te gusta el open source?** Es Go + SvelteKit + PostgreSQL, todo en Docker. Las contribuciones son bienvenidas → [Guía de contribución](#guía-de-contribución)

⭐ **Si la idea te gusta, deja una estrella** en [github.com/jsgiraldoh/noteops](https://github.com/jsgiraldoh/noteops)

</div>

---

## Tabla de contenidos

- [La aplicación de un vistazo](#la-aplicación-de-un-vistazo)
- [¿Por qué NoteOPs?](#por-qué-noteops)
- [Cómo un estudiante separa su turno](#cómo-un-estudiante-separa-su-turno)
- [API Reference](#api-reference)
- [Arquitectura](#arquitectura)
- [Stack técnico](#stack-técnico)
- [Inicio rápido](#inicio-rápido)
- [Importar tu planilla](#importar-tu-planilla)
- [Comandos útiles](#comandos-útiles)
- [CI/CD y seguridad](#cicd-y-seguridad)
- [Hoja de ruta](#hoja-de-ruta)
- [Equipo de desarrollo (Claude Code)](#equipo-de-desarrollo-claude-code)
- [Guía de contribución](#guía-de-contribución)
- [Licencia](#licencia)

---

## Cómo un estudiante separa su turno

> 🎓 **Esta es la sección estrella de NoteOPs.** Aquí el estudiante deja de ser espectador: para apartar su turno de sustentación tiene que hacer **su propia petición a la API**. No te preocupes si nunca has usado una terminal — sigue los pasos y, sin darte cuenta, habrás aprendido cómo funciona una API real.

Todo el flujo de reserva es **público**: no necesitas usuario ni contraseña. Solo necesitas dos datos que te entrega tu docente:

| Dato | Qué es | Cambia |
|---|---|---|
| `SUBJECT_ID` | Identificador de la materia | Fijo durante todo el semestre |
| `STUDENT_ID` | Tu identificador como estudiante | Fijo durante todo el semestre |

Vas a usar `curl`, una herramienta de línea de comandos que envía peticiones HTTP. Viene incluida en macOS y Linux, y en Windows 10/11 modernos. Reemplaza los valores entre llaves `{...}` por los tuyos.

### Paso 0 — Encuentra la sesión activa

Tu docente activa la sesión al iniciar la clase. Con el `SUBJECT_ID` averiguas cuál es la sesión de hoy:

```bash
curl http://noteops.local/api/sessions/active?subject_id={SUBJECT_ID}
```

Respuesta (HTTP 200):

```json
{
  "id": "f1a2b3c4-0000-1111-2222-333344445555",
  "subject_id": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
  "starts_at": "2025-04-10T14:00:00Z",
  "duration_min": 120,
  "slot_min": 20,
  "room": "Sala 201",
  "active": true
}
```

El campo **`id`** es tu `SESSION_ID` — guárdalo para los siguientes pasos.

> Si recibes **HTTP 404**, la sesión todavía no está activa. Espera a que tu docente la inicie.

### Paso 1 — Mira los turnos disponibles

Con el `SESSION_ID` listas todos los turnos (slots) de la clase:

```bash
curl http://noteops.local/api/sessions/{SESSION_ID}/slots
```

Respuesta (HTTP 200) — un arreglo de turnos:

```json
[
  {
    "id": "11111111-aaaa-bbbb-cccc-000000000001",
    "number": 1,
    "starts_at": "2025-04-10T14:00:00Z",
    "duration_min": 20,
    "student_id": "99999999-...-student-ocupado",
    "reserved_at": "2025-04-10T13:45:10Z"
  },
  {
    "id": "22222222-aaaa-bbbb-cccc-000000000002",
    "number": 2,
    "starts_at": "2025-04-10T14:20:00Z",
    "duration_min": 20,
    "student_id": null,
    "reserved_at": null
  }
]
```

Un turno está **libre** cuando su `student_id` es `null`. Elige uno libre y copia su **`id`** — ese será tu `SLOT_ID`.

### Paso 2 — Reserva tu turno

```bash
curl -X POST http://noteops.local/api/sessions/{SESSION_ID}/slots/{SLOT_ID}/reserve \
  -H "Content-Type: application/json" \
  -d '{"student_id": "{TU_STUDENT_ID}"}'
```

Respuesta exitosa (HTTP 200):

```json
{
  "id": "22222222-aaaa-bbbb-cccc-000000000002",
  "number": 2,
  "starts_at": "2025-04-10T14:20:00Z",
  "duration_min": 20,
  "student_id": "tu-student-id-aqui",
  "reserved_at": "2025-04-10T13:52:30Z"
}
```

🎉 **¡Listo! Tu turno quedó reservado.** El campo `reserved_at` confirma la hora exacta en que lo apartaste.

### Si algo sale mal

| Código | Significado | Qué hacer |
|---|---|---|
| `409 Conflict` | El turno ya fue reservado por otro, o la sesión ya no existe | Vuelve al Paso 1 y elige otro turno libre |
| `400 Bad Request` | Tu `STUDENT_ID` no existe o tiene formato inválido | Verifica el dato con tu docente |
| `404 Not Found` | La sesión no está activa | Espera a que el docente la inicie |

> 💡 **¿Prefieres la interfaz gráfica?** También puedes reservar desde la web. Pero te animamos a hacerlo con `curl`: acabas de ejecutar una petición `GET` y un `POST` a una API REST real, leer una respuesta JSON e interpretar códigos HTTP. Eso es exactamente lo que hace una aplicación por dentro — y ahora lo hiciste tú.

---

## API Reference

La URL base local es `http://noteops.local/api`. Los endpoints marcados como 🔓 **públicos** no requieren autenticación; el resto requieren la cabecera `Authorization: Bearer <token>` obtenida en el login.

### Autenticación

| Método | Ruta | Descripción |
|---|---|---|
| 🔓 `GET` | `/api/health` | Verifica que el servicio está arriba |
| 🔓 `POST` | `/api/auth/login` | Inicia sesión y devuelve un JWT |

```bash
# Login
curl -X POST http://noteops.local/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "docente@universidad.edu.co", "password": "••••••••"}'
```

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": { "id": "uuid", "full_name": "Docente Ejemplo", "role": "teacher" }
}
```

Usa el `token` en las siguientes peticiones protegidas:

```bash
curl http://noteops.local/api/subjects \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

### Materias

| Método | Ruta | Descripción |
|---|---|---|
| `GET` | `/api/subjects` | Lista las materias del docente |
| `POST` | `/api/subjects` | Crea una materia |
| `PATCH` | `/api/subjects/:id` | Actualiza una materia |
| `DELETE` | `/api/subjects/:id` | Elimina una materia (en cascada) |
| `GET` | `/api/subjects/:id/students` | Estudiantes inscritos en la materia |
| `GET` | `/api/subjects/:id/grades` | Notas completas: cortes, actividades y estudiantes |
| `GET` | `/api/subjects/:id/final-grades` | Nota definitiva por estudiante |
| `POST` | `/api/subjects/:id/enroll` | Inscribe un estudiante en la materia |
| `POST` | `/api/subjects/:id/import` | Importa una planilla completa (ver [Importar tu planilla](#importar-tu-planilla)) |

```bash
# Crear materia
curl -X POST http://noteops.local/api/subjects \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "Sistemas Operativos", "period": "2025-1", "group_name": "A", "faculty": "Ingeniería"}'
```

### Estudiantes

| Método | Ruta | Descripción |
|---|---|---|
| `POST` | `/api/students` | Crea un estudiante |
| `PATCH` | `/api/students/:id` | Actualiza nombre, correo o código |

### Notas

| Método | Ruta | Descripción |
|---|---|---|
| `POST` | `/api/grades` | Registra o actualiza una nota (upsert) |
| `PATCH` | `/api/grades/:id/comment` | Edita el comentario de una nota |

```bash
# Registrar una nota (0.0 a 5.0)
curl -X POST http://noteops.local/api/grades \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"enrollment_id": "uuid", "activity_id": "uuid", "value": 4.5, "comment": "Buen trabajo"}'
```

La **nota definitiva** se recalcula automáticamente en la base de datos (vista `student_final_grades`) a partir de los pesos de cada corte y actividad. No hay que calcularla a mano.

### Sesiones y turnos

| Método | Ruta | Acceso | Descripción |
|---|---|---|---|
| `POST` | `/api/sessions` | 🔒 | Crea una sesión y genera sus turnos |
| `POST` | `/api/sessions/:id/activate` | 🔒 | Activa la sesión (arranca el reloj) |
| `POST` | `/api/sessions/:id/deactivate` | 🔒 | Finaliza la sesión (la elimina junto con sus turnos) |
| `GET` | `/api/sessions/active?subject_id=uuid` | 🔓 | Sesión activa de una materia |
| `GET` | `/api/sessions/:id/slots` | 🔓 | Lista los turnos de una sesión |
| `POST` | `/api/sessions/:id/slots/:slotID/reserve` | 🔓 | Reserva un turno (ver [Cómo un estudiante separa su turno](#cómo-un-estudiante-separa-su-turno)) |

```bash
# Crear una sesión de 120 min con turnos de 20 min → genera 6 turnos
curl -X POST http://noteops.local/api/sessions \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"subject_id": "uuid", "starts_at": "2025-04-10T14:00:00Z", "duration_min": 120, "slot_min": 20, "room": "Sala 201"}'
```

```json
{
  "session": { "id": "uuid", "duration_min": 120, "slot_min": 20, "active": false },
  "slots": [
    { "number": 1, "starts_at": "2025-04-10T14:00:00Z", "duration_min": 20, "student_id": null },
    { "number": 2, "starts_at": "2025-04-10T14:20:00Z", "duration_min": 20, "student_id": null }
  ]
}
```

### WebSocket — el reloj en tiempo real

| Protocolo | Ruta | Acceso | Descripción |
|---|---|---|---|
| 🔓 `WS` | `/ws/session/:id` | 🔓 | Emite el estado del reloj de la sesión, **un tick por segundo** |

Cada tick es un mensaje JSON con esta forma:

```json
{
  "session_id": "uuid",
  "elapsed_sec": 642,
  "remaining_sec": 6558,
  "duration_min": 120,
  "is_active": true
}
```

| Campo | Significado |
|---|---|
| `elapsed_sec` | Segundos transcurridos desde el inicio |
| `remaining_sec` | Segundos restantes (0 si la sesión expiró) |
| `duration_min` | Duración total configurada |
| `is_active` | `false` cuando la sesión expiró o fue finalizada |

El componente `Clock.svelte` consume este flujo para pintar el reloj grande de la clase.

---

## Arquitectura

NoteOPs es un conjunto de servicios orquestados con Docker Compose detrás de un único reverse proxy (Traefik). El navegador solo habla con Traefik, que enruta por path y hostname hacia el frontend o el backend.

```
                         Navegador / curl
                               │  :80 / :443
                    ┌──────────▼───────────┐
                    │      Traefik v3       │  routing por host y path, TLS
                    └────┬─────────────┬────┘
                /api,/ws │             │ /
              ┌──────────▼──┐     ┌────▼─────────┐
              │  Backend     │     │  Frontend    │
              │  Go + Gin    │     │  SvelteKit   │
              │  :8080       │     │  :3000       │
              └───┬──────┬───┘     └──────────────┘
        SQL (pgx)│      │ estado WS
          ┌──────▼──┐ ┌─▼────────┐
          │PostgreSQL│ │  Redis   │
          │  :5432   │ │  :6379   │
          └──────────┘ └──────────┘

  Servicios auxiliares:
    • MinIO    :9001  almacenamiento de archivos (compatible S3)
    • Adminer  :8081  administrador web de la base de datos (desarrollo)
```

### Flujo de una request — registrar una nota

```
Docente en el navegador
  └─▶ SvelteKit  (POST /api/grades)
       └─▶ Traefik  (enruta /api → backend)
            └─▶ Go/Gin handler  (valida JWT, hace binding del JSON)
                 └─▶ Repository  (UPSERT con pgx en PostgreSQL)
                      └─▶ Vista student_final_grades  (recalcula la definitiva)
                 ◀── Grade como JSON
            ◀── 200 OK
       ◀── el store de Svelte se actualiza
  ◀── la tabla de notas se re-renderiza
```

El cálculo de la nota definitiva **no vive en el backend**: es una vista SQL (`student_final_grades`) que suma `valor × peso_actividad × peso_corte`. La base de datos es la única fuente de verdad del resultado.

### Estructura del repositorio

```
noteops/
├── backend/                  Go + Gin — API REST + WebSocket
│   ├── cmd/server/           Punto de entrada (main.go, rutas)
│   └── internal/
│       ├── config/           Carga de variables de entorno
│       ├── handlers/         HTTP handlers, hub WebSocket, errores seguros
│       ├── middleware/       Autenticación JWT, logger
│       ├── models/           Structs de dominio y DTOs
│       ├── repository/       Queries SQL con pgx (interfaz Repo + tests)
│       └── service/          Lógica de negocio (slots, reloj, agregados)
├── frontend/                 SvelteKit + TypeScript
│   └── src/
│       ├── lib/api/          Clientes HTTP tipados
│       ├── lib/stores/       Estado reactivo (auth, clock, session, notify)
│       ├── lib/components/   Clock, SlotGrid, GradeCell, modales
│       └── routes/           /, /session, /students, /subjects, /import, /login
├── workers/                  Python — agentes IA (futuro)
├── infra/
│   ├── traefik/              traefik.yml
│   └── postgres/             init.sql · rollback_to_admin.sql · 02_seed_data.sql*
├── .github/workflows/        ci.yml · release.yml
├── .claude/                  CLAUDE.md + skills del equipo (roles)
├── docker-compose.yml        Stack completo (perfiles local / registry)
├── docker-compose.seed.yml   Override para cargar datos de ejemplo
└── Makefile                  Atajos de desarrollo y operación

* 02_seed_data.sql es opcional y está excluido de git (datos privados).
```

---

## Stack técnico

| Capa | Tecnología | Versión | Por qué |
|---|---|---|---|
| **Backend** | Go + Gin | 1.23 | Concurrencia nativa para el WebSocket, binario estático de ~15 MB en imagen `scratch`, tipado fuerte |
| **Acceso a datos** | pgx | v5 | Driver PostgreSQL de alto rendimiento, sin ORM — control total del SQL |
| **Frontend** | SvelteKit | 2.x | Compila a JS sin runtime; bundle mínimo y reactividad simple |
| **Base de datos** | PostgreSQL | 16 | Vista SQL para la nota definitiva, `pgcrypto` para el hash de contraseñas |
| **Cache / WS** | Redis | 7 | Estado de sesiones WebSocket entre instancias |
| **Archivos** | MinIO | latest | Compatible con S3, self-hosted |
| **Proxy** | Traefik | v3 | TLS automático con Let's Encrypt, routing por labels de Docker |
| **DB Admin** | Adminer | 4 | UI web ligera (~10 MB) para inspeccionar la base en desarrollo |
| **Contenedores** | Docker + Compose | 24+ / v2 | Un comando levanta todo; entornos local y producción idénticos |

### Decisiones técnicas clave

**Go + Gin para el backend.** El hub de WebSocket mantiene una goroutine por cliente conectado durante la clase; Go maneja esa concurrencia sin bloquear. Además, el binario compilado produce una imagen Docker de ~15 MB (base `scratch`) y el tipado estático atrapa en compilación errores que en JavaScript solo aparecerían en runtime.

**pgx sin ORM.** Las queries se escriben a mano con parámetros posicionales (`$1, $2`). El precio es más código; la ganancia es transparencia total sobre qué SQL se ejecuta, cero magia, y poder apoyar el cálculo de notas en una vista de PostgreSQL en lugar de en el código.

**SvelteKit para el frontend.** Compila a vanilla JS sin enviar un runtime al navegador, lo que produce un bundle pequeño ideal para el aula. Su modelo de stores reactivos encaja naturalmente con el reloj en tiempo real alimentado por WebSocket.

---

## Inicio rápido

**Prerequisitos:** Docker 24+ y Docker Compose v2.

### Linux / macOS

```bash
# 1. Clonar
git clone https://github.com/jsgiraldoh/noteops.git
cd noteops

# 2. Configurar entorno (al menos cambia JWT_SECRET y DB_PASSWORD)
cp .env.example .env

# 3. Hostname local (una sola vez)
echo "127.0.0.1  noteops.local" | sudo tee -a /etc/hosts

# 4. Levantar
make up
```

### Windows (PowerShell como administrador)

```powershell
git clone https://github.com/jsgiraldoh/noteops.git
cd noteops
copy .env.example .env
Add-Content -Path "C:\Windows\System32\drivers\etc\hosts" -Value "127.0.0.1  noteops.local"
docker compose --profile local up -d --build
```

La aplicación queda en **http://noteops.local** y Adminer en **http://localhost:8081**.

**Credenciales por defecto:**

| Campo | Valor |
|---|---|
| Email | `admin@noteops.local` |
| Contraseña | `admin123` |

> Cambia la contraseña del admin y el `JWT_SECRET` antes de exponer el sistema.

### Modos de arranque

| Comando | Qué hace |
|---|---|
| `make up` | Build local **conservando** los datos existentes |
| `make fresh` | Arranque **limpio**: borra el volumen y deja solo el schema + el usuario admin |
| `make fresh-seed` | Arranque con datos de ejemplo desde `infra/postgres/02_seed_data.sql` |

### Variables de entorno

| Variable | Requerida | Default | Descripción |
|---|---|---|---|
| `DATABASE_URL` | ✅ | — | Conexión a PostgreSQL: `postgres://user:pass@host:5432/db?sslmode=disable` |
| `JWT_SECRET` | ✅ | — | Secreto para firmar JWT. Mínimo 32 caracteres. Genera con `openssl rand -hex 32` |
| `APP_ENV` | ❌ | `development` | `production` activa el modo release de Gin |
| `APP_PORT` | ❌ | `8080` | Puerto interno del backend |
| `REDIS_URL` | ❌ | `redis://redis:6379` | Conexión a Redis |
| `PUBLIC_API_URL` | ❌ | `http://noteops.local/api` | URL del API embebida en el frontend (build time) |
| `PUBLIC_WS_URL` | ❌ | `ws://noteops.local` | URL del WebSocket embebida en el frontend (build time) |

---

## Importar tu planilla

NoteOPs lee la planilla de Excel que ya usas y crea todo automáticamente. Sube el archivo desde la sección **Importar** de la interfaz (endpoint `POST /api/subjects/:id/import`).

**Formato esperado** — una hoja por materia, con la estructura institucional:

| Ubicación | Contenido |
|---|---|
| Hoja | Una por materia (el nombre de la hoja es la materia) |
| Fila 6, col E | Nombre del espacio académico |
| Fila 8 | Período y grupo |
| Fila 12 | Porcentajes (pesos) de actividades y cortes |
| Fila 14 en adelante | Listado de estudiantes (código y nombre) |

El importador detecta y crea:

- **Materias** con su período, grupo y facultad
- **Estudiantes** e inscripciones
- **Estructura de notas**: cortes y actividades con sus pesos
- **Notas existentes** (opcional)

Al cargar el archivo verás un **preview por materia** antes de confirmar. Puedes elegir entre dos modos:

1. **Solo estructura** — crea materias, estudiantes, cortes y actividades, sin notas.
2. **Estructura + notas** — además importa las calificaciones ya registradas en el Excel.

> El cálculo de la nota definitiva se valida contra la planilla original: el caso real del estudiante de referencia da exactamente `4.75`, igual que en Excel.

---

## Comandos útiles

```bash
# ── Entorno ──────────────────────────────────────────────
make up             # Build local conservando datos
make fresh          # Arranque limpio (schema + admin)
make fresh-seed     # Arranque con datos de ejemplo
make dev            # Solo infra (DB, Redis, MinIO)
make down           # Apagar todos los servicios
make logs           # Logs en tiempo real
make ps             # Estado de los contenedores

# ── Base de datos ────────────────────────────────────────
make shell-db       # Abrir psql en el contenedor de postgres
make rollback       # Limpiar la BD dejando solo el usuario admin

# ── Calidad ──────────────────────────────────────────────
make test             # Tests unitarios backend + check del frontend
make test-integration # Tests de integración con testcontainers (requiere Docker)

# ── Imágenes y release ───────────────────────────────────
make build          # Build local de las imágenes Docker
make push           # Push de imágenes a GHCR (requiere login)
make release VERSION=v1.0.0   # Crea el tag y dispara el workflow de release
```

---

## CI/CD y seguridad

El pipeline de integración continua (`.github/workflows/ci.yml`) corre en cada Pull Request y en cada push a `main`/`develop`, con la seguridad integrada en fases ordenadas:

```
Fase 0   secrets         gitleaks (escaneo de secretos en el historial)
Fase 1   backend-test    go vet → unit → integration → govulncheck
         frontend-check  npm ci → check → build
Fase 2   codeql          SAST de Go y TypeScript
                              │
Fase 3-5 image  ◀─────────────┘   build (una vez) → Trivy → push :latest (solo main)
```

El job `image` espera a que pasen **todos** los gates anteriores. En un Pull Request las imágenes se construyen y escanean pero **no se publican**; solo un push a `main` publica `:latest` en GHCR. Las versiones etiquetadas (`vX.Y.Z`) las publica `release.yml` al crear un tag.

> **No hay deployment automático.** El despliegue continuo se retiró del pipeline; las imágenes quedan disponibles en GHCR para desplegarlas donde quieras.

### Gates de seguridad

| Gate | Herramienta | Acción |
|---|---|---|
| Secretos | gitleaks | **Bloquea** |
| Pruebas unitarias e integración | go test | **Bloquea** |
| Dependencias Go | govulncheck | **Bloquea** (solo vulns en código realmente usado) |
| Imágenes Docker | Trivy (CRITICAL/HIGH) | **Bloquea**, ignora CVEs sin parche disponible |
| Análisis de código | CodeQL | Reporta a *Security → Code scanning* |

Todas las herramientas son open source y no requieren cuentas ni tokens de pago — CodeQL es gratuito en repositorios públicos.

---

## Hoja de ruta

| Estado | Funcionalidad |
|---|---|
| ✅ | Backend Go + Gin: materias, estudiantes, cortes, actividades, notas |
| ✅ | Cálculo automático de la nota definitiva (vista SQL) |
| ✅ | Frontend SvelteKit: notas, estudiantes, materias |
| ✅ | Sesiones de clase con reloj en tiempo real (WebSocket) |
| ✅ | Reserva pública de turnos por API (el reto del estudiante) |
| ✅ | Importación de planillas Excel (estructura + notas) |
| ✅ | CI con seguridad: gitleaks, govulncheck, CodeQL, Trivy |
| ✅ | Tests unitarios y de integración (testcontainers) |
| ⏳ | Exportar notas a Excel compatible con el formato institucional |
| ⏳ | Autenticación completa: registro de docentes y cambio de contraseña |
| ⏳ | Endurecer producción: restringir CORS y validar el origen del WebSocket |
| ⏳ | Workers Python: análisis de notas con agentes de IA |

---

## Equipo de desarrollo (Claude Code)

NoteOPs se desarrolla con un equipo de roles definidos como *skills* de Claude Code en `.claude/skills/`. Cada rol tiene un propósito y se invoca según la tarea:

| Rol | Skill | Responsabilidad |
|---|---|---|
| Developer | `dev` | Implementa cambios en backend y frontend |
| QA Engineer | `qa` | Tests unitarios y de integración |
| Arquitecto | `architect` | Mantiene el README y las decisiones técnicas |
| Ing. de Documentación | `docs` | Documenta código, endpoints y ADRs |
| Release Engineer | `release` | Versionado SemVer y publicación de releases |
| Diseñador UX/UI | `ux` | Consistencia visual y experiencia de uso |
| Security Engineer | `security` | Auditoría de vulnerabilidades y fallos de seguridad |
| DevSecOps Engineer | `devsecops` | Seguridad en el pipeline CI/CD |
| Marketing & Community | `marketing` | Copys, anuncios y gestión de la comunidad |

---

## Guía de contribución

Las contribuciones son bienvenidas. El proyecto trabaja con **trunk-based development** sobre `main`.

1. Haz un fork del repositorio.
2. Crea tu rama desde `main`: `git checkout -b feature/mi-aporte`.
3. Sigue las convenciones del proyecto (**Conventional Commits**: `feat:`, `fix:`, `docs:`, `test:`, `chore:`, `ci:`).
4. Asegúrate de que `make test` pase sin errores.
5. Abre un Pull Request describiendo el qué y el por qué.

Al abrir el PR, el CI ejecutará automáticamente los gates de seguridad y las pruebas. Consulta `.claude/CLAUDE.md` para las convenciones detalladas del proyecto y `CONTRIBUTING.md` para el flujo completo.

---

## Licencia

Distribuido bajo la licencia **Apache 2.0**.

```
Copyright 2025 Johan Sebastian Giraldo Hurtado

Licensed under the Apache License, Version 2.0.
Consulta el archivo LICENSE para los términos completos.
```
