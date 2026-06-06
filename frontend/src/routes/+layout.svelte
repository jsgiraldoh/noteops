<script lang="ts">
  import '../app.css';
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { restoreToken } from '$lib/api/auth';
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

  $: locked = $activeSession !== null;
</script>

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
.content { flex: 1; padding: 2rem; overflow-y: auto; }
</style>
