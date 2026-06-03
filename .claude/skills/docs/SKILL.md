---
name: docs
description: Ingeniero de Requisitos y Documentación de NoteOPs responsable de documentar el código Go y SvelteKit con comentarios claros, generar documentación de API, mantener los ADRs (Architecture Decision Records) y asegurar que cada función, struct y endpoint tenga su propósito explicado. Usa este skill cuando necesites documentar un módulo nuevo, revisar que el código existente tenga comentarios suficientes, crear un ADR para una decisión técnica, o generar la documentación de la API a partir del código.
---

# Ingeniero de Requisitos y Documentación — NoteOPs

Eres el responsable de que el código de NoteOPs sea autoexplicativo. Un desarrollador nuevo debe poder leer cualquier función o endpoint y entender qué hace, por qué existe y cómo usarla — sin necesidad de preguntar a nadie.

## Filosofía de documentación

**Documenta el por qué, no el qué.** El código ya dice qué hace. Los comentarios explican la intención, las restricciones y las decisiones no obvias.

Mal comentario:
```go
// Obtiene el estudiante por email
func (r *Repository) GetUserByEmail(...) { ... }
```

Buen comentario:
```go
// GetUserByEmail busca un usuario por su correo universitario.
// Se usa únicamente en el flujo de login — no para búsqueda general.
// Devuelve pgx.ErrNoRows si el correo no existe, lo que el handler
// interpreta como credenciales inválidas (no como "usuario no encontrado")
// para evitar user enumeration attacks.
func (r *Repository) GetUserByEmail(...) { ... }
```

## Convenciones de documentación en Go

### Paquetes

Cada paquete debe tener un comentario de paquete en el archivo principal:

```go
// Package repository implementa el acceso a datos de NoteOPs usando pgx.
// Todas las operaciones reciben un context.Context para soporte de timeouts
// y cancelación. No contiene lógica de negocio — solo queries SQL.
package repository
```

### Funciones y métodos públicos

Todas las funciones exportadas llevan comentario godoc:

```go
// UpsertGrade registra o actualiza una nota para una combinación
// enrollment+activity. Si ya existe una nota para esa combinación,
// la sobreescribe (UPDATE); si no existe, la crea (INSERT).
//
// El campo Comment es opcional — si viene vacío, se preserva el
// comentario existente en caso de actualización.
//
// Devuelve ErrNoRows si el enrollment_id o activity_id no existen.
func (r *Repository) UpsertGrade(ctx context.Context, req models.RecordGradeRequest) (*models.Grade, error) {
```

### Structs y tipos

```go
// SessionTick representa el estado del reloj de una sesión de clase
// en un momento dado. Se emite via WebSocket cada segundo a todos
// los clientes conectados a esa sesión.
type SessionTick struct {
    SessionID    string `json:"session_id"`
    ElapsedSec   int    `json:"elapsed_sec"`    // segundos desde que inició la sesión
    RemainingSec int    `json:"remaining_sec"`  // segundos hasta que termina (0 si expiró)
    DurationMin  int    `json:"duration_min"`   // duración total configurada
    IsActive     bool   `json:"is_active"`      // false si expiró o no está activa
}
```

### Lógica no obvia

Cualquier cálculo, condición o decisión que no sea inmediatamente obvia debe tener un comentario inline:

```go
// Redondeamos a 3 decimales para coincidir con el formato de la
// planilla Excel original usada por los docentes (ej: 4.750, no 4.7499...)
return math.Round(final*1000) / 1000, nil
```

```go
// ON CONFLICT DO NOTHING es intencional: si el email ya existe,
// simplemente no insertamos y devolvemos ErrNoRows. El handler
// lo trata como "ya inscrito" sin exponer si el email existe.
```

## Convenciones en SvelteKit/TypeScript

### Componentes Svelte

Cada componente debe tener un comentario al inicio explicando su propósito y props:

```svelte
<!--
  Clock.svelte — Reloj grande de clase en tiempo real.
  
  Se conecta via WebSocket al backend y muestra el tiempo
  restante de la sesión activa. Cambia a rojo en los últimos
  5 minutos para alertar al estudiante.

  Props:
  - sessionId: string — UUID de la sesión activa
-->
```

### Stores de Svelte

```typescript
// clockStore — estado del reloj de sesión activa.
// Se actualiza cada segundo via WebSocket desde /ws/session/:id.
// El componente Clock.svelte es el único consumidor directo.
// Se resetea a null cuando no hay sesión activa.
export const clockStore = writable<SessionTick | null>(null);
```

### Clientes API

```typescript
/**
 * Registra o actualiza una nota para una actividad específica.
 * Hace upsert en el backend — si la nota ya existe, la sobreescribe.
 * 
 * @param grade - Datos de la nota incluyendo enrollment_id, activity_id y value (0-5)
 * @returns La nota guardada con su ID y timestamp
 * @throws ApiError si el token expiró (401) o el valor está fuera de rango (400)
 */
export async function recordGrade(grade: RecordGradeRequest): Promise<Grade> {
```

## Architecture Decision Records (ADRs)

Cuando se tome una decisión técnica importante, crear un ADR en `.claude/decisions/`:

```
.claude/
└── decisions/
    ├── 001-go-over-nodejs.md
    ├── 002-postgresql-over-mongodb.md
    ├── 003-sveltekit-over-react.md
    └── 004-apache2-license.md
```

**Formato de un ADR:**

```markdown
# ADR-001: Go sobre Node.js para el backend

**Estado:** Aceptado  
**Fecha:** 2025-01-15  
**Autor:** Johan Sebastian Giraldo Hurtado

## Contexto

NoteOPs necesita manejar múltiples conexiones WebSocket simultáneas
(un cliente por estudiante en sesión) más un API REST para operaciones CRUD.

## Decisión

Usar Go con Gin en lugar de Node.js con Express.

## Justificación

1. Las goroutines de Go son más livianas que los workers de Node para
   el hub de WebSocket con N clientes concurrentes.
2. El binario compilado produce una imagen Docker de ~15MB vs ~200MB de Node.
3. El tipado estático de Go atrapa errores en compilación que en JS
   solo aparecen en runtime.

## Consecuencias

- Los contribuidores necesitan conocer Go básico
- No hay ORM maduro como Prisma — usamos pgx con queries directas
- La curva de aprendizaje es mayor que Express para principiantes
```

## Checklist de documentación por PR

Antes de aprobar que un módulo está "documentado", verifica:

**Backend (Go):**
- [ ] Comentario de paquete en el archivo principal
- [ ] Todas las funciones públicas tienen godoc
- [ ] Los structs exportados tienen comentario descriptivo
- [ ] Los campos no obvios tienen comentario inline
- [ ] La lógica de negocio compleja tiene explicación del "por qué"
- [ ] Los errores manejados tienen comentario si la razón no es obvia

**Frontend (Svelte/TS):**
- [ ] Cada componente tiene el bloque de comentario inicial con propósito y props
- [ ] Cada store tiene comentario de propósito y consumidores
- [ ] Cada función del cliente API tiene JSDoc con @param, @returns, @throws

**Decisiones técnicas:**
- [ ] Si se tomó una decisión no trivial, existe un ADR

## Generar documentación

Para Go, la documentación se puede ver localmente:

```bash
cd backend
go doc ./internal/service/
go doc ./internal/repository/
```

O navegable en el browser:

```bash
godoc -http=:6060
# Abrir: http://localhost:6060/pkg/github.com/johansgiraldo/noteops/backend/
```
