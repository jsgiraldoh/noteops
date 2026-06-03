<script lang="ts">
  import { gradesApi } from '$lib/api/grades';
  export let gradeId: string;
  export let initial = '';
  export let onClose: () => void = () => {};
  export let onSaved: () => void = () => {};

  let comment = initial;
  let loading = false;
  let error = '';

  async function save() {
    loading = true; error = '';
    try {
      await gradesApi.updateComment(gradeId, comment);
      onSaved();
      onClose();
    } catch (e: any) { error = e.message; }
    finally { loading = false; }
  }
</script>

<div class="overlay" on:click|self={onClose} role="dialog" aria-modal="true">
  <div class="modal card">
    <h3>Reflexión / comentario</h3>
    <p class="hint">Explica qué puede mejorar el estudiante en esta nota.</p>
    <textarea rows="5" bind:value={comment} placeholder="Buena entrega, pero puede mejorar la documentación del código…"></textarea>
    {#if error}<p class="error">{error}</p>{/if}
    <div class="actions">
      <button class="btn-secondary" on:click={onClose}>Cancelar</button>
      <button class="btn-primary" on:click={save} disabled={loading}>
        {loading ? 'Guardando…' : 'Guardar comentario'}
      </button>
    </div>
  </div>
</div>

<style>
.overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.6); display: flex; align-items: center; justify-content: center; z-index: 100; }
.modal { width: min(520px, 95vw); }
h3 { font-size: 1.1rem; margin-bottom: 0.3rem; }
.hint { color: var(--text2); font-size: 0.85rem; margin-bottom: 1rem; }
textarea { resize: vertical; min-height: 120px; margin-bottom: 1rem; }
.actions { display: flex; gap: 0.75rem; justify-content: flex-end; }
</style>
