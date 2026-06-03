# ADR-001: Go sobre Node.js para el backend

**Estado:** Aceptado  
**Fecha:** 2025-01-15  
**Autor:** Johan Sebastian Giraldo Hurtado

## Contexto

NoteOPs necesita manejar múltiples conexiones WebSocket simultáneas
(un cliente por estudiante por sesión activa) más un API REST para CRUD de notas.

## Decisión

Usar Go 1.22 con Gin en lugar de Node.js con Express o Fastify.

## Justificación

1. Las goroutines son más livianas que workers Node.js para el hub de WebSocket.
2. El binario compilado produce imagen Docker de ~15MB (base `scratch`) vs ~200MB de Node.
3. Tipado estático atrapa errores en compilación que JS solo muestra en runtime.
4. Sin ORM — pgx con queries directas da control total sobre el SQL generado.

## Consecuencias

- Contribuidores necesitan conocer Go básico
- No hay ORM como Prisma — queries SQL explícitas en `repository/`
- Curva de aprendizaje mayor que Express para principiantes
