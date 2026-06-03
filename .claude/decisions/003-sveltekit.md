# ADR-003: SvelteKit sobre React/Vue/Angular

**Estado:** Aceptado  
**Fecha:** 2025-01-15  
**Autor:** Johan Sebastian Giraldo Hurtado

## Contexto

El frontend necesita actualizaciones en tiempo real (WebSocket), un reloj grande visible durante la clase, y debe ser mantenible a largo plazo sin riesgo de que el framework se deprece.

## Decisión

Usar SvelteKit en lugar de Next.js (React), Nuxt (Vue) o Angular.

## Justificación

1. Svelte compila a vanilla JS en build time — sin runtime de framework en producción.
2. Si Svelte desaparece, el código compilado sigue funcionando.
3. Bundle mínimo — importante en redes lentas de aula de clase.
4. WebSocket nativo sin librerías adicionales.
5. Curva de aprendizaje menor que React para contribuidores nuevos.

## Consecuencias

- Ecosistema más pequeño que React
- Svelte 5 introdujo Runes — documentar la migración si aplica
- Menos recursos de aprendizaje disponibles que React/Vue
