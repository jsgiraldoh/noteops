# Changelog

Todos los cambios notables de NoteOPs se documentan aquí.

El formato está basado en [Keep a Changelog](https://keepachangelog.com/es/1.0.0/).
El versionado sigue [Semantic Versioning](https://semver.org/lang/es/).

## [Unreleased]

### Added
- Backend Go + Gin: REST API completa (estudiantes, materias, cortes, notas, sesiones, slots)
- WebSocket hub para reloj de sesión en tiempo real
- Schema PostgreSQL con vista `student_final_grades` para cálculo automático de nota definitiva
- Docker Compose con Traefik, PostgreSQL, Redis, MinIO
- GitHub Actions: CI (lint + test), CD (build + deploy), Release (imágenes GHCR)
- Licencia Apache 2.0 — Johan Sebastian Giraldo Hurtado
- Skills de equipo: dev, qa, architect, docs, release
