<script lang="ts">
  import { subjects, currentSubject } from '$lib/stores/subject';
  import { subjectsApi } from '$lib/api/subjects';
  import { activeSession } from '$lib/stores/session';

  let name = '';
  let period = '';
  let groupName = '';
  let faculty = '';
  let creating = false;
  let createError = '';

  let confirmDeleteId: string | null = null;
  let deleting = false;

  async function createSubject() {
    if (!name.trim() || !period.trim()) return;
    creating = true;
    createError = '';
    try {
      const created = await subjectsApi.create({
        name: name.trim(),
        period: period.trim(),
        group_name: groupName.trim() || undefined,
        faculty: faculty.trim() || undefined
      });
      subjects.update(list => [created, ...list]);
      name = ''; period = ''; groupName = ''; faculty = '';
    } catch (e: any) {
      createError = e.message;
    } finally {
      creating = false;
    }
  }

  async function deleteSubject(id: string) {
    deleting = true;
    try {
      await subjectsApi.remove(id);
      subjects.update(list => list.filter(s => s.id !== id));
      if ($currentSubject?.id === id) currentSubject.set(null);
      confirmDeleteId = null;
    } catch (e: any) {
      confirmDeleteId = null;
    } finally {
      deleting = false;
    }
  }
</script>

<svelte:head><title>Materias — NoteOPs</title></svelte:head>

<div class="page-header">
  <h1>Materias</h1>
  <p class="sub">Gestiona las materias del periodo académico</p>
</div>

<div class="layout">
  <!-- Formulario de creación -->
  <div class="card form-card">
    <h2>Nueva materia</h2>
    <div class="field">
      <label class="label">Nombre *</label>
      <input bind:value={name} placeholder="Sistemas Operativos" />
    </div>
    <div class="field">
      <label class="label">Periodo *</label>
      <input bind:value={period} placeholder="2025-1" />
    </div>
    <div class="two-fields">
      <div class="field">
        <label class="label">Grupo</label>
        <input bind:value={groupName} placeholder="A" />
      </div>
      <div class="field">
        <label class="label">Facultad</label>
        <input bind:value={faculty} placeholder="Ingeniería" />
      </div>
    </div>
    {#if createError}<p class="error">{createError}</p>{/if}
    <button
      class="btn-primary"
      on:click={createSubject}
      disabled={creating || !name.trim() || !period.trim()}
    >
      {creating ? 'Creando…' : '+ Agregar materia'}
    </button>
  </div>

  <!-- Lista de materias -->
  <div class="subject-list">
    <h2>{$subjects.length} materia{$subjects.length !== 1 ? 's' : ''}</h2>
    {#if $subjects.length === 0}
      <div class="empty">No hay materias registradas. Crea la primera.</div>
    {/if}
    {#each $subjects as s (s.id)}
      <div class="card subject-card" class:current={$currentSubject?.id === s.id}>
        <div class="subject-info">
          <span class="subject-name">{s.name}</span>
          <div class="subject-meta">
            <span class="badge badge-blue">{s.period}</span>
            {#if s.group_name}<span class="meta-item">Grupo {s.group_name}</span>{/if}
            {#if s.faculty}<span class="meta-item">{s.faculty}</span>{/if}
          </div>
          <span class="subject-id">ID: <code>{s.id}</code></span>
        </div>
        <div class="subject-actions">
          {#if confirmDeleteId === s.id}
            <div class="confirm-delete">
              <span>¿Eliminar?</span>
              <button class="btn-danger-sm" on:click={() => deleteSubject(s.id)} disabled={deleting}>
                {deleting ? '…' : 'Sí'}
              </button>
              <button class="btn-secondary-sm" on:click={() => confirmDeleteId = null}>No</button>
            </div>
          {:else}
            <button
              class="btn-delete"
              title="Eliminar materia"
              disabled={$activeSession?.subject_id === s.id}
              on:click={() => confirmDeleteId = s.id}
            >✕</button>
          {/if}
        </div>
      </div>
    {/each}
  </div>
</div>

<style>
.page-header { margin-bottom: 1.5rem; }
h1 { font-size: 1.5rem; font-weight: 700; }
h2 { font-size: 1.1rem; font-weight: 600; margin-bottom: 1rem; }
.sub { color: var(--text2); font-size: 0.85rem; }

.layout { display: grid; grid-template-columns: 320px 1fr; gap: 1.5rem; align-items: start; }
@media (max-width: 900px) { .layout { grid-template-columns: 1fr; } }

.form-card { display: flex; flex-direction: column; gap: 0.75rem; }
.two-fields { display: grid; grid-template-columns: 1fr 1fr; gap: 0.75rem; }

.subject-list { display: flex; flex-direction: column; gap: 0.75rem; }

.subject-card {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
  padding: 1rem 1.25rem;
  transition: border-color 0.15s;
}
.subject-card.current { border-color: var(--accent); }

.subject-info { display: flex; flex-direction: column; gap: 0.3rem; flex: 1; min-width: 0; }
.subject-name { font-weight: 600; font-size: 0.95rem; color: var(--text); }
.subject-meta { display: flex; align-items: center; gap: 0.5rem; flex-wrap: wrap; }
.meta-item { font-size: 0.78rem; color: var(--text2); }
.subject-id { font-size: 0.72rem; color: var(--text2); }
.subject-id code { font-family: monospace; color: var(--text2); user-select: all; }

.subject-actions { flex-shrink: 0; }
.btn-delete {
  background: transparent;
  border: 1px solid var(--border);
  color: var(--text2);
  border-radius: 6px;
  padding: 0.3rem 0.6rem;
  font-size: 0.8rem;
  cursor: pointer;
  transition: border-color 0.15s, color 0.15s;
}
.btn-delete:hover:not(:disabled) { border-color: var(--danger); color: var(--danger); }
.btn-delete:disabled { opacity: 0.3; cursor: not-allowed; }

.confirm-delete { display: flex; align-items: center; gap: 0.4rem; font-size: 0.82rem; color: var(--text2); }
.btn-danger-sm { background: var(--danger); color: #fff; border: none; border-radius: 6px; padding: 0.25rem 0.6rem; font-size: 0.8rem; cursor: pointer; }
.btn-secondary-sm { background: var(--bg3); color: var(--text); border: 1px solid var(--border); border-radius: 6px; padding: 0.25rem 0.6rem; font-size: 0.8rem; cursor: pointer; }

.empty { color: var(--text2); text-align: center; padding: 2rem; background: var(--bg2); border: 1px solid var(--border); border-radius: var(--radius); }
</style>
