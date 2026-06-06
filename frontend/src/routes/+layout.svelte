<script lang="ts">
  import '../app.css';
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { restoreToken, logout } from '$lib/api/auth';
  import { user } from '$lib/stores/auth';
  import { disconnectClock } from '$lib/stores/clock';
  import { notify } from '$lib/stores/notify';
  import { subjectsApi } from '$lib/api/subjects';
  import { subjects, currentSubject } from '$lib/stores/subject';
  import { activeSession } from '$lib/stores/session';
  import { sessionsApi } from '$lib/api/sessions';

  const PUBLIC = ['/login'];

  onMount(async () => {
    const token = restoreToken();
    if (!token && !PUBLIC.includes($page.url.pathname)) {
      goto('/login');
      return;
    }
    if (token) {
      try {
        const list = await subjectsApi.list();
        subjects.set(list);
        if (list.length && !$currentSubject) currentSubject.set(list[0]);
        // Restaurar nombre del usuario desde el token almacenado
        if (!$user) {
          try {
            const payload = JSON.parse(atob(token.split('.')[1]));
            user.set({ id: payload.user_id, email: payload.email, role: payload.role, full_name: payload.email.split('@')[0] });
          } catch {}
        }

        // Verificar si la sesión persistida en localStorage sigue activa
        if ($activeSession) {
          try {
            const fresh = await sessionsApi.getActive($activeSession.subject_id);
            if (fresh && fresh.id === $activeSession.id) {
              activeSession.set(fresh);
            } else {
              activeSession.set(null);
            }
          } catch {
            activeSession.set(null);
          }
        }
      } catch { goto('/login'); }
    }
  });

  function tryChangeSubject(s: typeof $currentSubject) {
    if ($activeSession) return;
    currentSubject.set(s);
  }

  function handleLogout() {
    disconnectClock();
    activeSession.set(null);
    subjects.set([]);
    currentSubject.set(null);
    user.set(null);
    logout();
    goto('/login');
  }

  $: locked = $activeSession !== null;
</script>

<!-- Notificaciones globales -->
<div class="notify-stack">
  {#each $notify as n (n.id)}
    <div class="notif notif-{n.type}" role="alert">
      <span class="notif-icon">{n.type === 'success' ? '✓' : n.type === 'error' ? '✕' : 'ℹ'}</span>
      <span class="notif-msg">{n.message}</span>
      <button class="notif-close" on:click={() => notify.dismiss(n.id)}>×</button>
    </div>
  {/each}
</div>

{#if $page.url.pathname === '/login'}
  <slot />
{:else}
  <div class="layout">
    <nav class="sidebar">
      <div class="brand">NoteOPs</div>
      <div class="subject-selector">
        {#each $subjects as s}
          <button
            class="sub-btn"
            class:active={$currentSubject?.id === s.id}
            class:locked={locked && $currentSubject?.id !== s.id}
            disabled={locked && $currentSubject?.id !== s.id}
            title={locked && $currentSubject?.id !== s.id ? 'Finaliza la sesión activa para cambiar de materia' : ''}
            on:click={() => tryChangeSubject(s)}
          >{s.name}</button>
        {/each}
        {#if locked}
          <div class="session-badge">Sesión en curso</div>
        {/if}
      </div>
      <div class="nav-links">
        <a href="/" class:active={$page.url.pathname === '/'}>📊 Notas</a>
        <a href="/session" class:active={$page.url.pathname.startsWith('/session')}>⏱ Sesión</a>
        <a href="/students" class:active={$page.url.pathname === '/students'}>👥 Estudiantes</a>
        <a href="/subjects" class:active={$page.url.pathname === '/subjects'}>📚 Materias</a>
        <a href="/import" class:active={$page.url.pathname === '/import'}>📥 Importar</a>
      </div>
      <div class="logout-section">
        {#if $user}<span class="username">{$user.full_name}</span>{/if}
        <button class="btn-logout" on:click={handleLogout}>↩ Cerrar sesión</button>
      </div>
    </nav>
    <main class="content">
      <slot />
    </main>
  </div>
{/if}

<style>
.layout { display: flex; height: 100vh; overflow: hidden; }
.sidebar {
  width: 220px;
  background: var(--bg2);
  border-right: 1px solid var(--border);
  padding: 1.25rem 1rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  flex-shrink: 0;
  height: 100vh;
  overflow-y: auto;
}
.brand { font-size: 1.3rem; font-weight: 700; color: var(--accent); flex-shrink: 0; }
.subject-selector { display: flex; flex-direction: column; gap: 0.3rem; flex: 1; min-height: 0; overflow-y: auto; }
.sub-btn { background: transparent; border: 1px solid transparent; color: var(--text2); border-radius: 8px; padding: 0.4rem 0.7rem; text-align: left; font-size: 0.85rem; }
.sub-btn.active { background: var(--bg3); border-color: var(--border); color: var(--text); }
.sub-btn.locked { opacity: 0.35; cursor: not-allowed; }
.session-badge { margin-top: 0.4rem; font-size: 0.72rem; color: #16a34a; background: #dcfce7; border: 1px solid #bbf7d0; border-radius: 6px; padding: 0.2rem 0.5rem; text-align: center; flex-shrink: 0; }
.nav-links {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  flex-shrink: 0;
  padding-top: 0.75rem;
  border-top: 1px solid var(--border);
}
.nav-links a { color: var(--text2); padding: 0.5rem 0.7rem; border-radius: 8px; font-size: 0.9rem; transition: background 0.15s; }
.nav-links a:hover, .nav-links a.active { background: var(--bg3); color: var(--text); text-decoration: none; }
.logout-section { flex-shrink: 0; padding-top: 0.75rem; border-top: 1px solid var(--border); display: flex; flex-direction: column; gap: 0.35rem; }
.username { font-size: 0.75rem; color: var(--text2); padding: 0 0.25rem; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.btn-logout { background: transparent; border: 1px solid var(--border); color: var(--text2); border-radius: 8px; padding: 0.45rem 0.7rem; font-size: 0.85rem; text-align: left; cursor: pointer; transition: border-color 0.15s, color 0.15s; }
.btn-logout:hover { border-color: var(--danger); color: var(--danger); }
.content { flex: 1; padding: 2rem; overflow-y: auto; }

/* ── Notificaciones globales ────────────────────────────────── */
.notify-stack {
  position: fixed; bottom: 1.5rem; right: 1.5rem;
  display: flex; flex-direction: column; gap: 0.5rem;
  z-index: 9999; pointer-events: none;
}
.notif {
  display: flex; align-items: center; gap: 0.6rem;
  padding: 0.65rem 1rem; border-radius: 8px;
  font-size: 0.85rem; font-weight: 500;
  box-shadow: 0 4px 16px rgba(0,0,0,0.4);
  pointer-events: all;
  animation: slideIn 0.2s ease;
  max-width: 360px;
}
.notif-success { background: #14401e; color: #4ade80; border: 1px solid #166534; }
.notif-error   { background: #450a0a; color: #fca5a5; border: 1px solid #7f1d1d; }
.notif-info    { background: #1e2a6e; color: #93aaff; border: 1px solid #1d4ed8; }
.notif-icon    { font-size: 0.9rem; flex-shrink: 0; }
.notif-msg     { flex: 1; line-height: 1.4; }
.notif-close   {
  background: transparent; border: none; color: inherit;
  opacity: 0.6; font-size: 1.1rem; cursor: pointer; flex-shrink: 0;
  padding: 0; line-height: 1;
}
.notif-close:hover { opacity: 1; }
@keyframes slideIn {
  from { opacity: 0; transform: translateY(8px); }
  to   { opacity: 1; transform: translateY(0); }
}
</style>
