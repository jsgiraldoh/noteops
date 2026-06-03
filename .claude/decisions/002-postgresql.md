# ADR-002: PostgreSQL sobre MongoDB

**Estado:** Aceptado  
**Fecha:** 2025-01-15  
**Autor:** Johan Sebastian Giraldo Hurtado

## Contexto

El modelo de datos de NoteOPs es claramente relacional: estudiante → materia → corte → actividad → nota. Se necesitan consultas agregadas para calcular la nota definitiva.

## Decisión

Usar PostgreSQL 16 en lugar de MongoDB.

## Justificación

1. El modelo es relacional por naturaleza — las foreign keys garantizan integridad.
2. La vista `student_final_grades` calcula la nota definitiva como SQL agregado, sin lógica en la aplicación.
3. Los agentes Python futuros (análisis de notas) funcionan igual de bien con Postgres.
4. JSONB disponible si se necesita flexibilidad en el futuro.

## Consecuencias

- Schema fijo — cambios requieren migraciones
- Mejor para reportes y cálculos agregados que MongoDB
