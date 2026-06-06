<script lang="ts">
  import type { Slot } from '$lib/api/sessions';
  export let slots: Slot[] = [];
  export let onReserve: (slot: Slot) => void = () => {};

  function fmt(iso: string) {
    return new Date(iso).toLocaleTimeString('es-CO', { hour: '2-digit', minute: '2-digit' });
  }
</script>

<div class="grid">
  {#each slots as slot (slot.id)}
    <button
      class="slot"
      class:taken={slot.student_id !== null}
      disabled={slot.student_id !== null}
      on:click={() => onReserve(slot)}
    >
      <span class="num">#{slot.number}</span>
      <span class="time">{fmt(slot.starts_at)}</span>
      <span class="dur">{slot.duration_min} min</span>
      {#if slot.student_id}
        <span class="badge badge-red">Ocupado</span>
      {:else}
        <span class="badge badge-green">Libre</span>
      {/if}
    </button>
  {/each}
</div>

<style>
.grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(140px, 1fr)); gap: 0.75rem; }
.slot {
  display: flex; flex-direction: column; align-items: center; gap: 0.3rem;
  background: var(--bg2); border: 1px solid var(--border); border-radius: var(--radius);
  padding: 1rem 0.5rem; cursor: pointer; transition: border-color 0.15s, background 0.15s;
}
.slot:not(:disabled):hover { border-color: var(--accent); background: var(--bg3); }
.slot.taken { background: #fee2e2; border-color: #fca5a5; cursor: not-allowed; }
.slot.taken .num  { color: #dc2626; }
.slot.taken .time { color: #ef4444; }
.slot.taken .dur  { color: #dc2626; }
.num { font-size: 1.3rem; font-weight: 700; color: var(--text); }
.time { font-size: 0.85rem; color: var(--accent); font-weight: 600; }
.dur { font-size: 0.75rem; color: var(--text2); }
</style>
