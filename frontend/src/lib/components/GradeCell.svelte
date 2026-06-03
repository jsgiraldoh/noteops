<script lang="ts">
  export let value: number | null = null;
  export let editable = false;
  export let onSave: (v: number) => void = () => {};

  let editing = false;
  let draft = '';

  function startEdit() { if (!editable) return; draft = value?.toString() ?? ''; editing = true; }
  function commit() {
    const n = parseFloat(draft);
    if (!isNaN(n) && n >= 0 && n <= 5) { onSave(n); }
    editing = false;
  }

  function color(v: number | null) {
    if (v === null) return '';
    if (v >= 4) return 'green';
    if (v >= 3) return 'yellow';
    return 'red';
  }
</script>

{#if editing}
  <input
    type="number" min="0" max="5" step="0.1"
    bind:value={draft}
    on:blur={commit}
    on:keydown={(e) => e.key === 'Enter' && commit()}
    style="width:70px;padding:0.2rem 0.4rem;font-size:0.85rem"
    autofocus
  />
{:else}
  <span
    class="cell badge badge-{color(value)}"
    class:empty={value === null}
    role={editable ? 'button' : undefined}
    tabindex={editable ? 0 : undefined}
    on:click={startEdit}
    on:keydown={(e) => e.key === 'Enter' && startEdit()}
  >
    {value !== null ? value.toFixed(1) : '—'}
  </span>
{/if}

<style>
.cell { cursor: pointer; min-width: 40px; text-align: center; display: inline-block; }
.empty { background: transparent; color: var(--text2); }
</style>
