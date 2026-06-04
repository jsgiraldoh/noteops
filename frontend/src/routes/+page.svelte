<script lang="ts">
  import { currentSubject } from '$lib/stores/subject';
  import { gradesApi, type SubjectGrades, type FinalGrade, type Grade } from '$lib/api/grades';
  import GradeCell from '$lib/components/GradeCell.svelte';
  import CommentModal from '$lib/components/CommentModal.svelte';

  let data: SubjectGrades | null = null;
  let finals: Record<string, FinalGrade> = {};
  let loading = true;
  let error = '';
  let commentModal: { gradeId: string; current: string } | null = null;

  // student_id → enrollment_id
  let enrollmentMap: Record<string, string> = {};
  // "enrollment_id:activity_id" → Grade
  let gradeMap: Record<string, Grade> = {};

  $: if ($currentSubject) loadGrades($currentSubject.id);

  async function loadGrades(sid: string) {
    loading = true; error = '';
    try {
      data = await gradesApi.bySubject(sid);
      enrollmentMap = Object.fromEntries((data.enrollments ?? []).map(e => [e.student_id, e.id]));
      gradeMap = Object.fromEntries((data.grades ?? []).map(g => [`${g.enrollment_id}:${g.activity_id}`, g]));
      finals = Object.fromEntries((data.final_grades ?? []).map(f => [f.student_id, f]));
    } catch (e: any) { error = e.message; }
    finally { loading = false; }
  }

  async function saveGrade(enrollmentId: string, activityId: string, value: number) {
    if (!enrollmentId) return;
    try {
      await gradesApi.record({ enrollment_id: enrollmentId, activity_id: activityId, value });
      if ($currentSubject) loadGrades($currentSubject.id);
    } catch {}
  }

  function cutPartial(enrollId: string, activities: { id: string; weight: number }[]): number | null {
    let sum = 0, totalWeight = 0;
    for (const act of activities) {
      const g = gradeMap[`${enrollId}:${act.id}`];
      if (g?.value != null) { sum += g.value * act.weight; totalWeight += act.weight; }
    }
    return totalWeight > 0 ? sum / totalWeight : null;
  }

  function colorClass(v: number | null) {
    if (v === null) return '';
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
          {@const enrollId = enrollmentMap[student.id] ?? ''}
          {@const final = finals[student.id]}
          <tr>
            <td>
              <div class="student-name">{student.full_name}</div>
              <div class="student-email">{student.email}</div>
            </td>
            {#each data.cuts as cut}
              {#each cut.activities as act}
                {@const grade = gradeMap[`${enrollId}:${act.id}`]}
                <td>
                  <GradeCell
                    value={grade?.value ?? null}
                    editable={true}
                    onSave={(v) => saveGrade(enrollId, act.id, v)}
                  />
                </td>
              {/each}
              {@const pc = cutPartial(enrollId, cut.activities)}
              <td>
                {#if pc !== null}
                  <span class="badge badge-{colorClass(pc)}">{pc.toFixed(2)}</span>
                {:else}—{/if}
              </td>
            {/each}
            <td>
              {#if final}
                <span class="badge badge-{colorClass(final.final_grade)}">
                  {final.final_grade.toFixed(2)}
                </span>
              {:else}—{/if}
            </td>
            <td>
              <button class="btn-secondary" style="padding:0.3rem 0.6rem;font-size:0.78rem"
                on:click={() => {
                  const firstGrade = data?.cuts.flatMap(c => c.activities)
                    .map(a => gradeMap[`${enrollId}:${a.id}`])
                    .find(g => g?.id);
                  commentModal = { gradeId: firstGrade?.id ?? '', current: firstGrade?.comment ?? '' };
                }}>
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
