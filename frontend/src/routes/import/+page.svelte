<script lang="ts">
  import { subjects, currentSubject } from '$lib/stores/subject';
  import { importApi, type ImportStudentRow, type ImportStructureRow, type ImportPayload } from '$lib/api/import';
  import * as XLSX from 'xlsx';

  let selectedSubjectId = $currentSubject?.id ?? '';
  let dragOver = false;
  let parsing = false;
  let importing = false;
  let error = '';
  let result: { students_created: number; students_enrolled: number; cuts_created: number; activities_created: number } | null = null;

  let students: ImportStudentRow[] = [];
  let structure: ImportStructureRow[] = [];

  $: hasPreview = students.length > 0 || structure.length > 0;
  $: canImport = hasPreview && selectedSubjectId;

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
      error = 'Solo se aceptan archivos .xlsx o .xls';
      return;
    }
    parsing = true;
    error = '';
    result = null;
    const reader = new FileReader();
    reader.onload = (e) => {
      try {
        const data = new Uint8Array(e.target!.result as ArrayBuffer);
        const wb = XLSX.read(data, { type: 'array' });
        students = parseStudents(wb);
        structure = parseStructure(wb);
        if (students.length === 0 && structure.length === 0) {
          error = 'No se encontraron datos. Verifica que el Excel tenga hojas "Estudiantes" y/o "Estructura".';
        }
      } catch (err: any) {
        error = 'Error al leer el archivo: ' + err.message;
      } finally {
        parsing = false;
      }
    };
    reader.readAsArrayBuffer(file);
  }

  function parseStudents(wb: XLSX.WorkBook): ImportStudentRow[] {
    const sheet = wb.Sheets['Estudiantes'] ?? wb.Sheets[wb.SheetNames[0]];
    if (!sheet) return [];
    const rows: any[] = XLSX.utils.sheet_to_json(sheet, { defval: '' });
    return rows
      .filter(r => r['nombre'] || r['full_name'])
      .map(r => ({
        full_name: String(r['nombre'] ?? r['full_name'] ?? '').trim().toUpperCase(),
        email:     String(r['email'] ?? '').trim().toLowerCase(),
        code:      String(r['codigo'] ?? r['code'] ?? '').trim()
      }))
      .filter(r => r.full_name && r.email);
  }

  function parseStructure(wb: XLSX.WorkBook): ImportStructureRow[] {
    const sheetName = wb.SheetNames.find(n => n.toLowerCase().includes('estructura') || n.toLowerCase().includes('notas'));
    if (!sheetName) return [];
    const sheet = wb.Sheets[sheetName];
    const rows: any[] = XLSX.utils.sheet_to_json(sheet, { defval: '' });
    return rows
      .filter(r => r['corte'] && r['actividad'])
      .map(r => ({
        cut_number:      Number(r['corte']),
        cut_name:        String(r['nombre_corte'] ?? `Corte ${r['corte']}`).trim(),
        cut_weight:      Number(r['peso_corte'] ?? 0),
        activity_name:   String(r['actividad']).trim(),
        activity_weight: Number(r['peso_actividad'] ?? 0)
      }))
      .filter(r => r.cut_number > 0 && r.activity_name);
  }

  async function runImport() {
    if (!canImport) return;
    importing = true;
    error = '';
    try {
      const payload: ImportPayload = { students, structure };
      result = await importApi.submit(selectedSubjectId, payload);
      students = [];
      structure = [];
    } catch (e: any) {
      error = e.message;
    } finally {
      importing = false;
    }
  }

  function reset() {
    students = []; structure = []; error = ''; result = null;
  }

  function downloadTemplate() {
    const wb = XLSX.utils.book_new();

    const wsStudents = XLSX.utils.aoa_to_sheet([
      ['codigo', 'nombre', 'email'],
      ['240220211012', 'ARCE PAREJA SEBASTIAN', '240220211012@noteops.edu'],
      ['240220212004', 'ARENAS RINCON JUAN MANUEL', '240220212004@noteops.edu']
    ]);
    XLSX.utils.book_append_sheet(wb, wsStudents, 'Estudiantes');

    const wsStructure = XLSX.utils.aoa_to_sheet([
      ['corte', 'nombre_corte', 'peso_corte', 'actividad', 'peso_actividad'],
      [1, 'Corte 1', 0.3, 'Parcial 1', 0.5],
      [1, 'Corte 1', 0.3, 'Taller 1', 0.5],
      [2, 'Corte 2', 0.3, 'Parcial 2', 1.0],
      [3, 'Corte 3', 0.4, 'Proyecto Final', 1.0]
    ]);
    XLSX.utils.book_append_sheet(wb, wsStructure, 'Estructura');

    XLSX.writeFile(wb, 'plantilla_noteops.xlsx');
  }
</script>

<svelte:head><title>Importar Excel — NoteOPs</title></svelte:head>

<div class="page-header">
  <h1>Importar desde Excel</h1>
  <p class="sub">Carga el listado de estudiantes y la estructura de notas del semestre</p>
</div>

<div class="layout">
  <!-- Panel izquierdo — configuración + carga -->
  <div class="left">
    <div class="card">
      <h2>1. Selecciona la materia</h2>
      <select bind:value={selectedSubjectId}>
        <option value="">— Elige una materia —</option>
        {#each $subjects as s}
          <option value={s.id}>{s.name} · {s.period}</option>
        {/each}
      </select>
    </div>

    <div class="card">
      <h2>2. Carga el archivo Excel</h2>
      <p class="hint">El Excel debe tener dos hojas: <strong>Estudiantes</strong> y <strong>Estructura</strong>.</p>
      <button class="btn-secondary template-btn" on:click={downloadTemplate}>
        ⬇ Descargar plantilla
      </button>

      <!-- Drop zone -->
      <div
        class="dropzone"
        class:dragover={dragOver}
        on:dragover|preventDefault={() => dragOver = true}
        on:dragleave={() => dragOver = false}
        on:drop={handleDrop}
        role="region"
        aria-label="Zona de carga de archivo"
      >
        {#if parsing}
          <span class="dz-text">Leyendo archivo…</span>
        {:else if hasPreview}
          <span class="dz-text dz-ok">✓ Archivo cargado — puedes reemplazarlo</span>
        {:else}
          <span class="dz-text">Arrastra el .xlsx aquí</span>
        {/if}
        <label class="btn-secondary file-btn">
          Seleccionar archivo
          <input type="file" accept=".xlsx,.xls" on:change={handleFileInput} style="display:none" />
        </label>
      </div>

      {#if error}<p class="error">{error}</p>{/if}
    </div>

    {#if result}
      <div class="card result-card">
        <h2>✓ Importación completada</h2>
        <div class="result-grid">
          <div class="result-item"><span class="result-num">{result.students_created}</span><span class="result-label">estudiantes procesados</span></div>
          <div class="result-item"><span class="result-num">{result.students_enrolled}</span><span class="result-label">inscripciones nuevas</span></div>
          <div class="result-item"><span class="result-num">{result.cuts_created}</span><span class="result-label">cortes creados</span></div>
          <div class="result-item"><span class="result-num">{result.activities_created}</span><span class="result-label">actividades creadas</span></div>
        </div>
        <button class="btn-secondary" on:click={reset}>Importar otro</button>
      </div>
    {/if}
  </div>

  <!-- Panel derecho — preview -->
  <div class="right">
    {#if hasPreview}
      <div class="preview-actions">
        <button class="btn-primary" on:click={runImport} disabled={!canImport || importing}>
          {importing ? 'Importando…' : `Importar (${students.length} estudiantes, ${structure.length} actividades)`}
        </button>
        <button class="btn-secondary" on:click={reset}>Limpiar</button>
      </div>

      {#if students.length > 0}
        <div class="card preview-card">
          <h2>Estudiantes ({students.length})</h2>
          <div class="table-wrap">
            <table>
              <thead><tr><th>#</th><th>Nombre</th><th>Email</th><th>Código</th></tr></thead>
              <tbody>
                {#each students.slice(0, 10) as s, i}
                  <tr>
                    <td style="color:var(--text2)">{i+1}</td>
                    <td style="font-weight:600">{s.full_name}</td>
                    <td style="color:var(--text2)">{s.email}</td>
                    <td style="font-family:monospace">{s.code || '—'}</td>
                  </tr>
                {/each}
                {#if students.length > 10}
                  <tr><td colspan="4" style="color:var(--text2);text-align:center">… y {students.length - 10} más</td></tr>
                {/if}
              </tbody>
            </table>
          </div>
        </div>
      {/if}

      {#if structure.length > 0}
        <div class="card preview-card">
          <h2>Estructura de notas ({structure.length} actividades)</h2>
          <div class="table-wrap">
            <table>
              <thead><tr><th>Corte</th><th>Nombre corte</th><th>Peso corte</th><th>Actividad</th><th>Peso act.</th></tr></thead>
              <tbody>
                {#each structure as row}
                  <tr>
                    <td>{row.cut_number}</td>
                    <td>{row.cut_name}</td>
                    <td>{(row.cut_weight * 100).toFixed(0)}%</td>
                    <td style="font-weight:600">{row.activity_name}</td>
                    <td>{(row.activity_weight * 100).toFixed(0)}%</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </div>
      {/if}
    {:else}
      <div class="empty-preview">
        <p>El preview aparecerá aquí después de cargar el archivo</p>
        <div class="format-hint">
          <p><strong>Hoja "Estudiantes"</strong></p>
          <code>codigo | nombre | email</code>
          <p style="margin-top:1rem"><strong>Hoja "Estructura"</strong></p>
          <code>corte | nombre_corte | peso_corte | actividad | peso_actividad</code>
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
.page-header { margin-bottom: 1.5rem; }
h1 { font-size: 1.5rem; font-weight: 700; }
h2 { font-size: 1rem; font-weight: 600; margin-bottom: 0.75rem; }
.sub { color: var(--text2); font-size: 0.85rem; }

.layout { display: grid; grid-template-columns: 360px 1fr; gap: 1.5rem; align-items: start; }
@media (max-width: 900px) { .layout { grid-template-columns: 1fr; } }

.left { display: flex; flex-direction: column; gap: 1rem; }
.right { display: flex; flex-direction: column; gap: 1rem; }

.hint { font-size: 0.82rem; color: var(--text2); margin-bottom: 0.75rem; }
.template-btn { width: 100%; margin-bottom: 0.75rem; font-size: 0.85rem; }

.dropzone {
  border: 2px dashed var(--border);
  border-radius: var(--radius);
  padding: 1.5rem 1rem;
  text-align: center;
  transition: border-color 0.15s, background 0.15s;
  display: flex; flex-direction: column; align-items: center; gap: 0.75rem;
}
.dropzone.dragover { border-color: var(--accent); background: var(--bg3); }
.dz-text { font-size: 0.85rem; color: var(--text2); }
.dz-ok { color: var(--accent2); }
.file-btn { cursor: pointer; font-size: 0.82rem; padding: 0.35rem 1rem; }

.preview-actions { display: flex; gap: 0.75rem; }
.preview-actions button { flex: 1; }
.preview-card { padding: 1.25rem; }
.table-wrap { overflow-x: auto; }

.result-card { border-color: var(--accent2); }
.result-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 0.75rem; margin-bottom: 1rem; }
.result-item { display: flex; flex-direction: column; align-items: center; background: var(--bg3); border-radius: 8px; padding: 0.75rem; }
.result-num { font-size: 1.8rem; font-weight: 700; color: var(--accent2); }
.result-label { font-size: 0.72rem; color: var(--text2); text-align: center; }

.empty-preview {
  background: var(--bg2); border: 1px solid var(--border); border-radius: var(--radius);
  padding: 2rem; text-align: center; color: var(--text2);
}
.format-hint { margin-top: 1.5rem; text-align: left; display: inline-block; }
.format-hint code { display: block; background: var(--bg3); border-radius: 6px; padding: 0.4rem 0.75rem; font-size: 0.8rem; color: var(--accent); margin-top: 0.3rem; }
</style>
