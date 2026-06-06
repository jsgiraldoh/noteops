# NoteOPs — Contexto del proyecto para Claude

## Qué es este proyecto

NoteOPs es un sistema open source de gestión de notas académicas con:
- Backend en **Go + Gin** (REST API + WebSocket)
- Frontend en **SvelteKit** (tiempo real, reloj de clase)
- Base de datos **PostgreSQL** con cálculo automático de nota definitiva
- **Redis** para estado de sesiones WebSocket
- Infraestructura **Docker + Traefik**
- CI/CD con **GitHub Actions** y registry en **GHCR**

**Autor original:** Johan Sebastian Giraldo Hurtado  
**Licencia:** Apache 2.0  
**Repositorio:** https://github.com/johansgiraldo/noteops

---

## Estructura del repositorio

```
noteops/
├── backend/          Go + Gin — API REST + WebSocket
│   ├── cmd/server/   Punto de entrada (main.go)
│   ├── internal/
│   │   ├── config/       Carga de variables de entorno
│   │   ├── handlers/     HTTP handlers + WebSocket hub
│   │   ├── middleware/   JWT auth + CORS
│   │   ├── models/       Structs de dominio y DTOs
│   │   ├── repository/   Queries SQL (pgx, sin ORM)
│   │   └── service/      Lógica de negocio
│   └── migrations/   Migraciones SQL
├── frontend/         SvelteKit
│   └── src/
│       ├── lib/      API clients, stores, componentes
│       └── routes/   Páginas
├── workers/          Python — agentes IA (futuro)
├── infra/            Traefik + PostgreSQL init.sql
├── .github/
│   └── workflows/    ci.yml · cd.yml · release.yml
└── .claude/
    └── skills/       Roles del equipo de desarrollo
```

---

## Convenciones del proyecto

- **Branches:** `main` (producción) · `develop` (integración) · `feature/*` · `fix/*`
- **Commits:** Conventional Commits — `feat:` `fix:` `docs:` `test:` `chore:` `refactor:`
- **Versioning:** SemVer — `vMAJOR.MINOR.PATCH`
- **PR:** siempre hacia `develop`, nunca directo a `main`
- **Release:** tag `v*.*.*` en `main` → GitHub Actions construye y publica imágenes en GHCR

---

## Roles disponibles

Cada skill define el comportamiento de un rol específico del equipo.
Invoca el rol adecuado según la tarea:

| Skill | Rol | Cuándo usarlo |
|---|---|---|
| `.claude/skills/dev/SKILL.md` | Developer | Crear PRs de backend o frontend |
| `.claude/skills/qa/SKILL.md` | QA Engineer | Tests unitarios e integración |
| `.claude/skills/architect/SKILL.md` | Arquitecto | Actualizar README y documentación técnica |
| `.claude/skills/docs/SKILL.md` | Ing. de Requisitos | Documentar código y módulos |
| `.claude/skills/release/SKILL.md` | Release Engineer | Versionar y publicar releases |
| `.claude/skills/ux/SKILL.md` | Diseñador UX/UI | Auditar y mejorar el frontend — contraste, estados, consistencia visual |
| `.claude/skills/security/SKILL.md` | Security Engineer | Auditar vulnerabilidades, fallos de seguridad y code smells de seguridad |
| `.claude/skills/devsecops/SKILL.md` | DevSecOps Engineer | Integrar seguridad en CI/CD — escaneo de dependencias, SAST, secretos, imágenes |
| `.claude/skills/marketing/SKILL.md` | Marketing & Community | Copys, anuncios, contenido promocional y gestión de la comunidad |
