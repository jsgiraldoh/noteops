# ═══════════════════════════════════════════════════════════════
# NoteOPs — Frontend Multi-stage Dockerfile
#
# Etapas:
#   base    → Node + dependencias npm cacheadas
#   dev     → servidor de desarrollo con Vite HMR
#   builder → build estático de SvelteKit
#   prod    → Node mínimo sirviendo el build
# ═══════════════════════════════════════════════════════════════

# ── base: dependencias npm cacheadas ─────────────────────────
# Instala node_modules una sola vez.
# Docker reutiliza esta capa mientras package.json no cambie.
FROM node:26-alpine AS base

WORKDIR /app

COPY package*.json ./

# npm install en lugar de npm ci — no requiere package-lock.json
RUN npm install

# ── dev: Vite con Hot Module Replacement ─────────────────────
# Código fuente montado como volumen en docker-compose.
# Cambios en .svelte se reflejan sin rebuild.
FROM base AS dev

COPY . .

EXPOSE 3000
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]

# ── builder: compilación SvelteKit ───────────────────────────
# svelte-kit sync genera el tsconfig y tipos internos antes del build.
# Las variables PUBLIC_* se embeben en el JS compilado en build time.
FROM base AS builder

COPY . .

ARG PUBLIC_API_URL=http://noteops.local/api
ARG PUBLIC_WS_URL=ws://noteops.local

ENV PUBLIC_API_URL=$PUBLIC_API_URL
ENV PUBLIC_WS_URL=$PUBLIC_WS_URL

# svelte-kit sync debe correr antes de vite build
# genera .svelte-kit/ con tsconfig y tipos necesarios
RUN npx svelte-kit sync && npm run build

# ── prod: Node mínimo sirviendo el build ─────────────────────
# Solo runtime de Node + output compilado.
# Sin código fuente ni dependencias de desarrollo.
FROM node:26-alpine AS prod

WORKDIR /app

COPY package*.json ./
RUN npm install --omit=dev

# Output del builder
COPY --from=builder /app/build ./build

EXPOSE 3000
ENV HOST=0.0.0.0 PORT=3000 NODE_ENV=production

CMD ["node", "build"]
