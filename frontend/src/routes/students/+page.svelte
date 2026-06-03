<script lang="ts">
  import { currentSubject } from '$lib/stores/subject';
  import { studentsApi, type Student } from '$lib/api/students';
  import StudentForm from '$lib/components/StudentForm.svelte';

  let students: Student[] = [];
  let loading = true;

  $: if ($currentSubject) load($currentSubject.id);

  async function load(sid: string) {
    loading = true;
    try { students = await studentsApi.bySubject(sid) ?? []; }
    catch { students = []; }
    finally { loading = false; }
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
        <thead><tr><th>#</th><th>Nombre</th><th>Correo</th><th>Código</th></tr></thead>
        <tbody>
          {#each students as s, i}
            <tr>
              <td style="color:var(--text2)">{i+1}</td>
              <td style="font-weight:600">{s.full_name}</td>
              <td style="color:var(--text2)">{s.email}</td>
              <td style="font-family:monospace">{s.code || '—'}</td>
            </tr>
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
</style>
