<script lang="ts">
  import { currentSubject } from '$lib/stores/subject';
  import { studentsApi, type Student } from '$lib/api/students';
  import StudentForm from '$lib/components/StudentForm.svelte';
  import { notify } from '$lib/stores/notify';

  let students: Student[] = [];
  let loading = true;

  // Edición inline
  let editingId: string | null = null;
  let editName = '';
  let editEmail = '';
  let editCode = '';
  let saving = false;
  let saveError = '';

  $: if ($currentSubject) load($currentSubject.id);

  async function load(sid: string) {
    loading = true;
    try { students = await studentsApi.bySubject(sid) ?? []; }
    catch { students = []; }
    finally { loading = false; }
  }

  function startEdit(s: Student) {
    editingId = s.id;
    editName = s.full_name;
    editEmail = s.email;
    editCode = s.code || '';
    saveError = '';
  }

  function cancelEdit() { editingId = null; saveError = ''; }

  async function saveEdit() {
    if (!editingId || !editName.trim() || !editEmail.trim()) return;
    saving = true; saveError = '';
    try {
      const updated = await studentsApi.update(editingId, {
        full_name: editName.trim(),
        email: editEmail.trim(),
        code: editCode.trim() || undefined
      });
      students = students.map(s => s.id === updated.id ? updated : s);
      notify.success('Estudiante actualizado');
      editingId = null;
    } catch (e: any) { saveError = e.message; }
    finally { saving = false; }
  }
</script>

<svelte:head><title>Estudiantes — NoteOPs</title></svelte:head>

<div class="page-header">
  <h1>Estudiantes</h1>
  {#if $currentSubject}<p class="sub">{$currentSubject.name} · {students.length} inscritos</p>{/if}
</div>

{#if $currentSubject}
  <div class="card" style="margin-bottom:1.5rem">
    <h2>Registrar estudiante</h2>
    <StudentForm subjectId={$currentSubject.id} onCreated={() => load($currentSubject!.id)} />
  </div>
  <div class="card">
    {#if loading}
      <p class="state">Cargando…</p>
    {:else if !students.length}
      <p class="state">Sin estudiantes inscritos aún.</p>
    {:else}
      <table>
        <thead>
          <tr><th>#</th><th>Nombre</th><th>Correo</th><th>Código</th><th></th></tr>
        </thead>
        <tbody>
          {#each students as s, i (s.id)}
            {#if editingId === s.id}
              <tr class="edit-row">
                <td style="color:var(--text2)">{i+1}</td>
                <td><input class="edit-input" bind:value={editName} placeholder="Nombre completo" /></td>
                <td><input class="edit-input" bind:value={editEmail} placeholder="correo@ejemplo.com" type="email" /></td>
                <td><input class="edit-input" bind:value={editCode} placeholder="Código" style="width:90px" /></td>
                <td class="actions-cell">
                  <button class="btn-save" on:click={saveEdit} disabled={saving || !editName.trim() || !editEmail.trim()}>
                    {saving ? '…' : '✓'}
                  </button>
                  <button class="btn-cancel" on:click={cancelEdit}>✕</button>
                </td>
              </tr>
              {#if saveError}
                <tr><td colspan="5"><p class="error" style="margin:0.25rem 0">{saveError}</p></td></tr>
              {/if}
            {:else}
              <tr>
                <td style="color:var(--text2)">{i+1}</td>
                <td style="font-weight:600">{s.full_name}</td>
                <td style="color:var(--text2)">{s.email}</td>
                <td style="font-family:monospace">{s.code || '—'}</td>
                <td class="actions-cell">
                  <button class="btn-edit-row" on:click={() => startEdit(s)} title="Editar">✏</button>
                </td>
              </tr>
            {/if}
          {/each}
        </tbody>
      </table>
    {/if}
  </div>
{:else}
  <p class="state">Selecciona una materia en el menú lateral.</p>
{/if}

<style>
.page-header { margin-bottom: 1.5rem; }
h1 { font-size: 1.5rem; font-weight: 700; }
h2 { font-size: 1rem; font-weight: 600; margin-bottom: 1rem; }
.sub { color: var(--text2); font-size: 0.85rem; }
.state { text-align: center; padding: 2rem; color: var(--text2); }
.edit-input {
  width: 100%; background: var(--bg3); border: 1px solid var(--accent);
  border-radius: 6px; color: var(--text); font-family: inherit;
  font-size: 0.85rem; padding: 0.3rem 0.5rem;
}
.edit-row td { background: var(--bg3); }
.actions-cell { display: flex; gap: 0.3rem; align-items: center; white-space: nowrap; }
.btn-edit-row {
  background: transparent; border: 1px solid var(--border); color: var(--text2);
  border-radius: 6px; padding: 0.2rem 0.5rem; font-size: 0.8rem; cursor: pointer;
  transition: border-color 0.15s, color 0.15s;
}
.btn-edit-row:hover { border-color: var(--accent); color: var(--accent); }
.btn-save {
  background: var(--accent2); color: #000; border: none;
  border-radius: 6px; padding: 0.2rem 0.6rem; font-size: 0.85rem; cursor: pointer; font-weight: 700;
}
.btn-save:disabled { opacity: 0.4; cursor: not-allowed; }
.btn-cancel {
  background: transparent; border: 1px solid var(--border); color: var(--text2);
  border-radius: 6px; padding: 0.2rem 0.5rem; font-size: 0.8rem; cursor: pointer;
}
</style>
