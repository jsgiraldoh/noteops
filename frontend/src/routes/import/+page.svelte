<script lang="ts">
  import { subjects } from '$lib/stores/subject';
  import { subjectsApi } from '$lib/api/subjects';
  import { importApi, type ImportStudentRow, type ImportStructureRow, type ImportGradeRow } from '$lib/api/import';
  import * as XLSX from 'xlsx';

  // ─── Tipos ────────────────────────────────────────────────────────────────
  interface ParsedSubject {
    sheetName: string;
    subjectName: string;
    period: string;
    group: string;
    faculty: string;
    students: ImportStudentRow[];
    structure: ImportStructureRow[];
    grades: ImportGradeRow[];
    error?: string;
  }

  interface ImportStatus {
    subjectName: string;
    state: 'pending' | 'done' | 'error';
    message: string;
  }

  // ─── Estado ───────────────────────────────────────────────────────────────
  let dragOver = false;
  let parsing = false;
  let importing = false;
  let fileName = '';
  let parsed: ParsedSubject[] = [];
  let statuses: ImportStatus[] = [];
  let parseError = '';
  let importDone = false;
  let importGrades = true; // toggle: importar notas además de la estructura

  $: hasPreview = parsed.length > 0;
  $: totalStudents  = parsed.reduce((s, p) => s + p.students.length, 0);
  $: totalActivities = parsed.reduce((s, p) => s + p.structure.length, 0);
  $: totalGrades    = parsed.reduce((s, p) => s + p.grades.length, 0);

  // ─── Carga de archivo ─────────────────────────────────────────────────────
  function handleDrop(e: DragEvent) {
    e.preventDefault();
    dragOver = false;
    const file = e.dataTransfer?.files[0];
    if (file) parseFile(file);
  }

  function handleFileInput(e: Event) {
    const file = (e.target as HTMLInputElement).files?.[0];
    if (file) parseFile(file);
  }

  function parseFile(file: File) {
    if (!file.name.match(/\.(xlsx|xls)$/i)) {
      parseError = 'Solo se aceptan archivos .xlsx o .xls';
      return;
    }
    parsing = true;
    parseError = '';
    parsed = [];
    importDone = false;
    fileName = file.name;

    const reader = new FileReader();
    reader.onload = (e) => {
      try {
        const data = new Uint8Array(e.target!.result as ArrayBuffer);
        const wb = XLSX.read(data, { type: 'array' });
        parsed = parsePlanilla(wb);
        if (parsed.length === 0) {
          parseError = 'No se detectaron materias. Verifica que el Excel tenga el formato de planilla institucional.';
        }
      } catch (err: any) {
        parseError = 'Error al leer el archivo: ' + err.message;
      } finally {
        parsing = false;
      }
    };
    reader.readAsArrayBuffer(file);
  }

  // ─── Parser de planilla institucional ────────────────────────────────────
  function parsePlanilla(wb: XLSX.WorkBook): ParsedSubject[] {
    const result: ParsedSubject[] = [];

    for (const sheetName of wb.SheetNames) {
      const ws = wb.Sheets[sheetName];
      if (!ws['!ref']) continue;

      const rows: any[][] = XLSX.utils.sheet_to_json(ws, { header: 1, defval: '' });

      // Verificar que es formato de planilla buscando "ESPACIO ACADÉMICO" en R6
      const r6 = (rows[5] ?? []).map(v => String(v).trim());
      if (!r6.some(v => v.toUpperCase().includes('ESPACIO'))) continue;

      // Metadatos de la materia
      const faculty   = String(rows[1]?.[0] ?? '').trim();
      const subjectName = String(rows[5]?.[4] ?? sheetName).trim();
      const period    = String(rows[7]?.[4] ?? '').trim();
      const group     = String(rows[7]?.[16] ?? '').trim();

      // Pesos (R12 = índice 11) y nombres (R13 = índice 12)
      const weightsRow: any[]  = rows[11] ?? [];
      const headersRow: any[]  = rows[12] ?? [];
      const corteRow: any[]    = rows[10] ?? [];   // R11

      // Peso de cada corte (columnas DFC: 12, 20, 28)
      const cutWeights = [
        Number(weightsRow[12] ?? 0),
        Number(weightsRow[20] ?? 0),
        Number(weightsRow[28] ?? 0)
      ];

      // Nombre de cada corte desde R11
      const cutNames = [
        String(corteRow[6]  ?? 'Primer Corte').replace(/-?\s*CORTE/i, '').trim() || 'Primer Corte',
        String(corteRow[14] ?? 'Segundo Corte').replace(/-?\s*CORTE/i, '').trim() || 'Segundo Corte',
        String(corteRow[22] ?? 'Tercer Corte').replace(/-?\s*CORTE/i, '').trim() || 'Tercer Corte'
      ];

      // Rangos de columnas por corte (actividades, excluye la columna DFC)
      const corteRanges = [
        { cutNum: 1, start: 6,  end: 11 },
        { cutNum: 2, start: 14, end: 19 },
        { cutNum: 3, start: 22, end: 27 }
      ];

      const structure: ImportStructureRow[] = [];
      for (const { cutNum, start, end } of corteRanges) {
        const cutWeight = cutWeights[cutNum - 1];
        const cutName   = cutNames[cutNum - 1];
        for (let c = start; c <= end; c++) {
          const actName   = String(headersRow[c] ?? '').trim();
          const actWeight = Number(weightsRow[c] ?? 0);
          if (actName && actWeight > 0 && actName !== 'DFC' + cutNum) {
            structure.push({ cut_number: cutNum, cut_name: cutName, cut_weight: cutWeight, activity_name: actName, activity_weight: actWeight });
          }
        }
      }

      // Estudiantes y notas desde R14 en adelante (índice 13+)
      const students: ImportStudentRow[] = [];
      const grades: ImportGradeRow[] = [];

      // Rangos de columnas de actividades por corte (mismo orden que structure)
      const gradeRanges = [
        { cutNum: 1, start: 6,  end: 11 },
        { cutNum: 2, start: 14, end: 19 },
        { cutNum: 3, start: 22, end: 27 }
      ];

      for (let i = 13; i < rows.length; i++) {
        const row = rows[i] ?? [];
        const code = String(row[1] ?? '').trim();
        const name = String(row[2] ?? '').trim().toUpperCase();
        if (!code || !name || !/^\d{5,}/.test(code)) continue;

        students.push({ full_name: name, email: `${code}@noteops.edu`, code });

        // Extraer notas de cada columna de actividad
        for (const { cutNum, start, end } of gradeRanges) {
          for (let c = start; c <= end; c++) {
            const actName   = String(headersRow[c] ?? '').trim();
            const actWeight = Number(weightsRow[c] ?? 0);
            const rawVal    = row[c];
            // Solo importar si la actividad existe en la estructura y tiene valor numérico
            if (actName && actWeight > 0 && rawVal !== '' && rawVal !== null && rawVal !== undefined) {
              const val = Number(rawVal);
              if (!isNaN(val)) {
                grades.push({ student_code: code, cut_number: cutNum, activity_name: actName, value: val });
              }
            }
          }
        }
      }

      if (subjectName && (students.length > 0 || structure.length > 0)) {
        result.push({ sheetName, subjectName, period, group, faculty, students, structure, grades });
      }
    }

    return result;
  }

  // ─── Importación ──────────────────────────────────────────────────────────
  async function runImport() {
    if (!hasPreview) return;
    importing = true;
    importDone = false;
    statuses = parsed.map(p => ({ subjectName: p.subjectName, state: 'pending', message: 'En espera…' }));

    for (let i = 0; i < parsed.length; i++) {
      const p = parsed[i];
      statuses[i] = { ...statuses[i], state: 'pending', message: 'Creando materia…' };
      statuses = [...statuses];

      try {
        // 1. Crear la materia
        const created = await subjectsApi.create({
          name: p.subjectName,
          period: p.period,
          group_name: p.group,
          faculty: p.faculty
        });

        // Actualizar store de materias
        subjects.update(list => [created, ...list]);

        statuses[i] = { ...statuses[i], message: 'Importando estudiantes y estructura…' };
        statuses = [...statuses];

        // 2. Importar estudiantes, estructura y (opcionalmente) notas
        const result = await importApi.submit(created.id, {
          students: p.students,
          structure: p.structure,
          grades: importGrades ? p.grades : []
        });

        const gradeMsg = result.grades_imported > 0 ? ` · ${result.grades_imported} notas` : '';
        statuses[i] = {
          subjectName: p.subjectName,
          state: 'done',
          message: `${result.students_created} estudiantes · ${result.cuts_created} cortes · ${result.activities_created} actividades${gradeMsg}`
        };
      } catch (err: any) {
        statuses[i] = { subjectName: p.subjectName, state: 'error', message: err.message };
      }
      statuses = [...statuses];
    }

    importing = false;
    importDone = true;
  }

  function reset() {
    parsed = [];
    statuses = [];
    parseError = '';
    importDone = false;
    fileName = '';
  }
</script>

<svelte:head><title>Importar planilla — NoteOPs</title></svelte:head>

<div class="page-header">
  <h1>Importar planilla institucional</h1>
  <p class="sub">Carga la planilla Excel del semestre — el sistema detecta las materias, estudiantes y estructura de notas automáticamente</p>
</div>

<div class="layout">
  <!-- Columna izquierda — carga -->
  <div class="left">
    <div class="card">
      <h2>Subir planilla</h2>
      <p class="hint">
        El Excel debe tener <strong>una hoja por materia</strong> con el formato institucional
        (ESPACIO ACADÉMICO en R6, PORCENTAJES en R12, estudiantes desde R14).
      </p>

      <div
        class="dropzone"
        class:dragover={dragOver}
        on:dragover|preventDefault={() => dragOver = true}
        on:dragleave={() => dragOver = false}
        on:drop={handleDrop}
        role="region"
        aria-label="Zona de carga"
      >
        {#if parsing}
          <span class="dz-icon">⏳</span>
          <span class="dz-text">Leyendo planilla…</span>
        {:else if hasPreview}
          <span class="dz-icon">✓</span>
          <span class="dz-text dz-ok">{fileName}</span>
          <span class="dz-sub">{parsed.length} materias · {totalStudents} estudiantes · {totalActivities} actividades</span>
        {:else}
          <span class="dz-icon">📂</span>
          <span class="dz-text">Arrastra la planilla .xlsx aquí</span>
        {/if}
        <label class="btn-secondary file-btn">
          {hasPreview ? 'Reemplazar archivo' : 'Seleccionar archivo'}
          <input type="file" accept=".xlsx,.xls" on:change={handleFileInput} style="display:none" />
        </label>
      </div>

      {#if parseError}<p class="error">{parseError}</p>{/if}
    </div>

    {#if hasPreview && !importDone}
      <div class="card action-card">
        <p class="action-summary">
          Se crearán <strong>{parsed.length}</strong> materia(s) con
          <strong>{totalStudents}</strong> estudiantes y
          <strong>{totalActivities}</strong> actividades en total.
          {#if totalGrades > 0}
            El archivo contiene <strong>{totalGrades}</strong> registros de notas.
          {/if}
        </p>
        {#if totalGrades > 0}
          <label class="toggle-label">
            <input type="checkbox" bind:checked={importGrades} />
            Importar notas existentes en el Excel ({totalGrades} registros)
          </label>
        {/if}
        <button class="btn-primary" on:click={runImport} disabled={importing}>
          {importing ? 'Importando…' : '⬆ Importar todo'}
        </button>
        <button class="btn-secondary" on:click={reset} disabled={importing}>Cancelar</button>
      </div>
    {/if}

    {#if importDone}
      <div class="card done-card">
        <h2>✓ Importación completa</h2>
        {#each statuses as s}
          <div class="status-row" class:status-ok={s.state==='done'} class:status-err={s.state==='error'}>
            <span class="status-icon">{s.state==='done' ? '✓' : s.state==='error' ? '✕' : '…'}</span>
            <div>
              <span class="status-name">{s.subjectName}</span>
              <span class="status-msg">{s.message}</span>
            </div>
          </div>
        {/each}
        <button class="btn-secondary" on:click={reset} style="margin-top:0.5rem">Importar otra planilla</button>
      </div>
    {/if}
  </div>

  <!-- Columna derecha — preview por materia -->
  <div class="right">
    {#if importing}
      <div class="card progress-card">
        <h2>Progreso</h2>
        {#each statuses as s}
          <div class="progress-row" class:prog-done={s.state==='done'} class:prog-err={s.state==='error'}>
            <span class="prog-dot"></span>
            <div>
              <span class="prog-name">{s.subjectName}</span>
              <span class="prog-msg">{s.message}</span>
            </div>
          </div>
        {/each}
      </div>
    {:else if hasPreview}
      {#each parsed as p}
        <div class="card subject-preview">
          <div class="sp-header">
            <div>
              <span class="sp-name">{p.subjectName}</span>
              <div class="sp-meta">
                <span class="badge badge-blue">{p.period}</span>
                {#if p.group}<span class="meta-item">Grupo {p.group}</span>{/if}
                {#if p.faculty}<span class="meta-item">{p.faculty}</span>{/if}
              </div>
            </div>
            <div class="sp-counts">
              <span class="count-badge">{p.students.length} estudiantes</span>
              <span class="count-badge">{p.structure.length} actividades</span>
              {#if p.grades.length > 0}
                <span class="count-badge count-grades">{p.grades.length} notas</span>
              {/if}
            </div>
          </div>

          <!-- Estructura de notas -->
          {#if p.structure.length > 0}
            <details class="sp-details">
              <summary>Ver estructura de notas ({p.structure.length} actividades)</summary>
              <table class="sp-table">
                <thead><tr><th>Corte</th><th>Peso corte</th><th>Actividad</th><th>Peso actividad</th></tr></thead>
                <tbody>
                  {#each p.structure as row}
                    <tr>
                      <td>{row.cut_name}</td>
                      <td>{(row.cut_weight * 100).toFixed(0)}%</td>
                      <td style="font-weight:600">{row.activity_name}</td>
                      <td>{(row.activity_weight * 100).toFixed(0)}%</td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </details>
          {/if}

          <!-- Primeros estudiantes -->
          {#if p.students.length > 0}
            <details class="sp-details">
              <summary>Ver estudiantes ({p.students.length})</summary>
              <table class="sp-table">
                <thead><tr><th>Código</th><th>Nombre</th></tr></thead>
                <tbody>
                  {#each p.students.slice(0, 8) as s}
                    <tr><td style="font-family:monospace">{s.code}</td><td>{s.full_name}</td></tr>
                  {/each}
                  {#if p.students.length > 8}
                    <tr><td colspan="2" style="color:var(--text2);text-align:center">… y {p.students.length - 8} más</td></tr>
                  {/if}
                </tbody>
              </table>
            </details>
          {/if}
        </div>
      {/each}
    {:else}
      <div class="empty-state">
        <p>El preview de cada materia aparecerá aquí</p>
        <div class="format-hint">
          <strong>Formato detectado automáticamente:</strong>
          <ul>
            <li>Una hoja por materia</li>
            <li>R6 col E → nombre de la materia</li>
            <li>R8 → período y grupo</li>
            <li>R12 → pesos de actividades y cortes</li>
            <li>R14+ → listado de estudiantes (código, nombre)</li>
          </ul>
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
.page-header { margin-bottom: 1.5rem; }
h1 { font-size: 1.5rem; font-weight: 700; }
h2 { font-size: 1rem; font-weight: 600; margin-bottom: 0.75rem; }
.sub { color: var(--text2); font-size: 0.85rem; margin-top: 0.25rem; }
.hint { font-size: 0.82rem; color: var(--text2); margin-bottom: 0.75rem; }

.layout { display: grid; grid-template-columns: 340px 1fr; gap: 1.5rem; align-items: start; }
@media (max-width: 900px) { .layout { grid-template-columns: 1fr; } }
.left { display: flex; flex-direction: column; gap: 1rem; }
.right { display: flex; flex-direction: column; gap: 1rem; }

/* Drop zone */
.dropzone {
  border: 2px dashed var(--border); border-radius: var(--radius);
  padding: 1.5rem 1rem; text-align: center;
  display: flex; flex-direction: column; align-items: center; gap: 0.5rem;
  transition: border-color 0.15s, background 0.15s;
}
.dropzone.dragover { border-color: var(--accent); background: var(--bg3); }
.dz-icon { font-size: 1.8rem; }
.dz-text { font-size: 0.9rem; color: var(--text2); }
.dz-ok { color: var(--accent2); font-weight: 600; }
.dz-sub { font-size: 0.78rem; color: var(--text2); }
.file-btn { cursor: pointer; font-size: 0.82rem; padding: 0.35rem 1rem; margin-top: 0.25rem; }

/* Action card */
.action-card { display: flex; flex-direction: column; gap: 0.5rem; }
.action-summary { font-size: 0.85rem; color: var(--text2); margin-bottom: 0.25rem; }
.action-summary strong { color: var(--text); }
.toggle-label { display: flex; align-items: center; gap: 0.5rem; font-size: 0.85rem; color: var(--text); cursor: pointer; padding: 0.4rem 0; }
.toggle-label input[type="checkbox"] { width: 16px; height: 16px; accent-color: var(--accent); cursor: pointer; }

/* Done card */
.done-card { border-color: var(--accent2); display: flex; flex-direction: column; gap: 0.5rem; }
.status-row { display: flex; gap: 0.75rem; align-items: flex-start; padding: 0.4rem 0; border-bottom: 1px solid var(--border); }
.status-row:last-of-type { border-bottom: none; }
.status-ok .status-icon { color: var(--accent2); }
.status-err .status-icon { color: var(--danger); }
.status-icon { font-size: 1rem; flex-shrink: 0; margin-top: 0.1rem; }
.status-name { display: block; font-weight: 600; font-size: 0.85rem; }
.status-msg { display: block; font-size: 0.78rem; color: var(--text2); }

/* Progress */
.progress-card { display: flex; flex-direction: column; gap: 0.5rem; }
.progress-row { display: flex; gap: 0.75rem; align-items: flex-start; padding: 0.5rem; border-radius: 6px; background: var(--bg3); }
.prog-done { background: #14401e22; }
.prog-err  { background: #450a0a44; }
.prog-dot { width: 8px; height: 8px; border-radius: 50%; background: var(--text2); flex-shrink: 0; margin-top: 0.4rem; }
.prog-done .prog-dot { background: var(--accent2); }
.prog-err  .prog-dot { background: var(--danger); }
.prog-name { display: block; font-weight: 600; font-size: 0.85rem; }
.prog-msg  { display: block; font-size: 0.78rem; color: var(--text2); }

/* Subject preview */
.subject-preview { padding: 1rem 1.25rem; }
.sp-header { display: flex; justify-content: space-between; align-items: flex-start; gap: 1rem; margin-bottom: 0.75rem; }
.sp-name { font-weight: 700; font-size: 0.95rem; display: block; margin-bottom: 0.35rem; }
.sp-meta { display: flex; align-items: center; gap: 0.5rem; flex-wrap: wrap; }
.meta-item { font-size: 0.78rem; color: var(--text2); }
.sp-counts { display: flex; flex-direction: column; gap: 0.25rem; align-items: flex-end; flex-shrink: 0; }
.count-badge { background: var(--bg3); border: 1px solid var(--border); border-radius: 999px; font-size: 0.72rem; padding: 0.15rem 0.5rem; white-space: nowrap; }
.count-grades { background: #14401e33; border-color: #166534; color: #4ade80; }

.sp-details { margin-top: 0.5rem; }
.sp-details summary { font-size: 0.82rem; color: var(--text2); cursor: pointer; padding: 0.25rem 0; }
.sp-details summary:hover { color: var(--text); }
.sp-table { width: 100%; margin-top: 0.5rem; font-size: 0.82rem; }

/* Empty state */
.empty-state { background: var(--bg2); border: 1px solid var(--border); border-radius: var(--radius); padding: 2rem; color: var(--text2); }
.format-hint { margin-top: 1rem; font-size: 0.82rem; }
.format-hint ul { margin-top: 0.5rem; padding-left: 1.25rem; display: flex; flex-direction: column; gap: 0.25rem; }
</style>
