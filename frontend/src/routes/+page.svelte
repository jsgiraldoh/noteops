<script lang="ts">
  import { onMount } from 'svelte';
  import { currentSubject } from '$lib/stores/subject';
  import { gradesApi, type SubjectGrades, type FinalGrade } from '$lib/api/grades';
  import GradeCell from '$lib/components/GradeCell.svelte';
  import CommentModal from '$lib/components/CommentModal.svelte';

  let data: SubjectGrades | null = null;
  let finals: Record<string, FinalGrade> = {};
  let loading = true;
  let error = '';
  let commentModal: { gradeId: string; current: string } | null = null;
  // Map enrollmentId+activityId -> gradeId (for comment lookup)
  let gradeIndex: Record<string, string> = {};

  $: if ($currentSubject) loadGrades($currentSubject.id);

  async function loadGrades(sid: string) {
    loading = true; error = '';
    try {
      data = await gradesApi.bySubject(sid);
      const fg = await gradesApi.finalBySubject(sid);
      finals = Object.fromEntries(fg.map(f => [f.student_id, f]));
    } catch (e: any) { error = e.message; }
    finally { loading = false; }
  }

  async function saveGrade(enrollmentId: string, activityId: string, value: number) {
    try {
      const g = await gradesApi.record({ enrollment_id: enrollmentId, activity_id: activityId, value });
      gradeIndex[`${enrollmentId}:${activityId}`] = g.id;
      if ($currentSubject) loadGrades($currentSubject.id);
    } catch {}
  }

  function finalColor(v: number) {
    if (v >= 4) return 'green'; if (v >= 3) return 'yellow'; return 'red';
  }
</script>

<svelte:head><title>Notas — {$currentSubject?.name ?? 'NoteOPs'}</title></svelte:head>

<div class="page-header">
  <div>
    <h1>{$currentSubject?.name ?? 'Selecciona una materia'}</h1>
    {#if $currentSubject}<p class="sub">{$currentSubject.period} · Grupo {$currentSubject.group_name}</p>{/if}
  </div>
</div>

{#if loading}
  <div class="state">Cargando notas…</div>
{:else if error}
  <div class="state error">{error}</div>
{:else if !data || !data.students?.length}
  <div class="state">No hay estudiantes inscritos en esta materia.</div>
{:else}
  <div class="table-wrap card">
    <table>
      <thead>
        <tr>
          <th>Estudiante</th>
          {#each data.cuts as cut}
            {#each cut.activities as act}
              <th title="Corte {cut.number} · Peso corte {cut.weight} · Peso act. {act.weight}">
                C{cut.number} — {act.name}
              </th>
            {/each}
            <th>PC{cut.number}</th>
          {/each}
          <th>Definitiva</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        {#each data.students as student}
          {@const final = finals[student.id]}
          <tr>
            <td>
              <div class="student-name">{student.full_name}</div>
              <div class="student-email">{student.email}</div>
            </td>
            {#each data.cuts as cut}
              {#each cut.activities as act}
                <td>
                  <GradeCell
                    value={null}
                    editable={true}
                    onSave={(v) => saveGrade('', act.id, v)}
                  />
                </td>
              {/each}
              <td>—</td>
            {/each}
            <td>
              {#if final}
                <span class="badge badge-{finalColor(final.final_grade)}">
                  {final.final_grade.toFixed(2)}
                </span>
              {:else}—{/if}
            </td>
            <td>
              <button class="btn-secondary" style="padding:0.3rem 0.6rem;font-size:0.78rem"
                on:click={() => commentModal = { gradeId: '', current: '' }}>
                💬
              </button>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{/if}

{#if commentModal}
  <CommentModal
    gradeId={commentModal.gradeId}
    initial={commentModal.current}
    onClose={() => commentModal = null}
    onSaved={() => $currentSubject && loadGrades($currentSubject.id)}
  />
{/if}

<style>
.page-header { display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: 1.5rem; }
h1 { font-size: 1.5rem; font-weight: 700; }
.sub { color: var(--text2); font-size: 0.85rem; margin-top: 0.2rem; }
.table-wrap { padding: 0; overflow-x: auto; }
.state { text-align: center; padding: 3rem; color: var(--text2); }
.student-name { font-weight: 600; font-size: 0.9rem; }
.student-email { color: var(--text2); font-size: 0.78rem; }
</style>
