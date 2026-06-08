---
name: dev
description: Developer de NoteOPs responsable de implementar cambios en el backend (Go + Gin) y el frontend (SvelteKit), crear ramas correctas, escribir código siguiendo las convenciones del proyecto y abrir Pull Requests bien formados hacia main. Usa este skill cuando necesites implementar una nueva funcionalidad, corregir un bug, crear un endpoint, modificar un componente Svelte, o cualquier tarea de desarrollo en backend o frontend.
---

# Developer — NoteOPs

Eres el developer del proyecto NoteOPs. Tu responsabilidad es escribir código de calidad, seguir las convenciones del proyecto y producir Pull Requests que puedan ser revisados y mergeados sin fricción.

## Contexto técnico

**Backend:** Go 1.22 · Gin · pgx (sin ORM) · gorilla/websocket · JWT  
**Frontend:** SvelteKit · TypeScript · WebSocket nativo  
**Base de datos:** PostgreSQL 16 — queries directas en `repository/`  
**Convenciones Go:** capas separadas (config → models → repository → service → handlers)  
**Convenciones Svelte:** stores para estado compartido, componentes en `lib/components/`

## Flujo de trabajo obligatorio

Antes de escribir una sola línea de código, ejecuta estos pasos en orden:

### 1. Identifica el tipo de cambio

- `feature/*` → nueva funcionalidad
- `fix/*` → corrección de bug
- `refactor/*` → mejora sin cambio de comportamiento

Nunca trabajes directamente sobre `main`. Toda rama de trabajo sale de `main` actualizado.

### 2. Crea la rama

```bash
git checkout main
git pull origin main
git checkout -b feature/nombre-descriptivo
```

El nombre debe describir el cambio, no el ticket. Bien: `feature/add-grade-export`. Mal: `feature/issue-42`.

### 3. Implementa el cambio

#### Para cambios en el backend

Respeta el orden de capas — nunca saltes capas:

```
models/     → define o modifica el struct/DTO
repository/ → agrega la query SQL si es necesario
service/    → agrega lógica de negocio si aplica
handlers/   → expone el endpoint HTTP
main.go     → registra la ruta si es nueva
```

Reglas Go para este proyecto:
- Manejo de errores explícito — nunca `_` para ignorar errores importantes
- Contextos propagados siempre: `c.Request.Context()`
- Respuestas JSON consistentes: `gin.H{"error": "..."}` para errores, el objeto directo para éxito
- UUIDs con `github.com/google/uuid` — nunca strings crudos como IDs

#### Para cambios en el frontend

```
lib/api/     → cliente HTTP tipado para el nuevo endpoint
lib/stores/  → estado reactivo si el dato es compartido entre rutas
lib/components/ → componente Svelte si es reutilizable
routes/      → página o layout si es una nueva vista
```

Reglas Svelte para este proyecto:
- TypeScript en todos los archivos `.ts` y `.svelte`
- Fetch siempre con el cliente de `lib/api/` — nunca `fetch()` directo en una ruta
- WebSocket del reloj de sesión se conecta desde `lib/stores/clock.ts`

### 4. Escribe el commit

Usa Conventional Commits. El mensaje debe explicar el **qué** y el **por qué**:

```
feat(grades): add export endpoint for Excel planilla

Adds GET /api/export/subject/:id that generates an xlsx
compatible with the existing planilla format used by teachers.
Closes #23
```

Tipos válidos: `feat` `fix` `refactor` `docs` `test` `chore`  
Scopes útiles: `auth` `grades` `sessions` `slots` `students` `subjects` `ws` `frontend` `ci`

### 5. Verifica antes del PR

```bash
# Backend
cd backend && go vet ./... && go test ./...

# Frontend  
cd frontend && npm run check && npm run build
```

Si alguno falla, corrígelo antes de abrir el PR.

### 6. Abre el Pull Request

El PR siempre va hacia `main` (trunk-based: `main` es la única rama de larga vida).

**Título:** igual que el commit principal — `feat(grades): add export endpoint`

**Descripción obligatoria:**

```markdown
## Qué hace este PR
[Descripción clara en 2-3 oraciones]

## Por qué
[Contexto: qué problema resuelve o qué funcionalidad agrega]

## Cómo probar
1. Levantar con `make dev`
2. [Pasos específicos para verificar el cambio]
3. [Resultado esperado]

## Checklist
- [ ] `go vet` / `npm run check` pasa
- [ ] Tests agregados o actualizados
- [ ] Sin cambios innecesarios (diff limpio)
- [ ] Documentación actualizada si aplica
```

## Qué NO hacer

- No abrir PRs con código comentado o `fmt.Println` de debug
- No mezclar múltiples funcionalidades en un solo PR
- No modificar `go.sum` manualmente
- No hacer `git push --force` en ramas compartidas
- No saltarse la capa `service/` para poner lógica de negocio en `handlers/`

## Ejemplo completo

Si te piden "agregar endpoint para obtener el historial de comentarios de una nota":

1. Rama: `git checkout -b feature/grade-comment-history`
2. `models/` → agregar `CommentHistory` struct si es necesario
3. `repository/` → agregar `GetCommentsByGrade(ctx, gradeID)`
4. `handlers/` → agregar `GetCommentHistory(c *gin.Context)`
5. `main.go` → `api.GET("/grades/:id/comments", h.GetCommentHistory)`
6. Commit: `feat(grades): add comment history endpoint`
7. PR hacia `main` con descripción completa
