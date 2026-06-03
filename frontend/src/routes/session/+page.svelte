<script lang="ts">
  import { onDestroy } from 'svelte';
  import { currentSubject } from '$lib/stores/subject';
  import { sessionsApi, type Session, type Slot } from '$lib/api/sessions';
  import { connectClock, disconnectClock } from '$lib/stores/clock';
  import Clock from '$lib/components/Clock.svelte';
  import SlotGrid from '$lib/components/SlotGrid.svelte';

  let session: Session | null = null;
  let slots: Slot[] = [];
  let loading = false;
  let error = '';
  let durationMin = 120;
  let slotMin = 20;
  let room = '';
  const slotOptions = [5, 10, 20, 30, 60, 120];

  onDestroy(disconnectClock);

  async function createSession() {
    if (!$currentSubject) return;
    loading = true; error = '';
    try {
      const res = await sessionsApi.create({
        subject_id: $currentSubject.id,
        starts_at: new Date().toISOString(),
        duration_min: durationMin,
        slot_min: slotMin,
        room
      });
      session = res.session;
      slots = res.slots;
    } catch (e: any) { error = e.message; }
    finally { loading = false; }
  }

  async function activate() {
    if (!session) return;
    await sessionsApi.activate(session.id);
    session = { ...session, active: true };
    connectClock(session.id);
  }

  async function reserveSlot(slot: Slot) {
    const studentId = prompt('UUID del estudiante:');
    if (!studentId || !session) return;
    try {
      const updated = await sessionsApi.reserve(session.id, slot.id, studentId);
      slots = slots.map(s => s.id === updated.id ? updated : s);
    } catch { alert('Espacio ya reservado o error.'); }
  }
</script>

<svelte:head><title>Sesión — NoteOPs</title></svelte:head>

<div class="page-header">
  <h1>Sesión de clase</h1>
  {#if $currentSubject}<p class="sub">{$currentSubject.name}</p>{/if}
</div>

<div class="two-col">
  <div class="left">
    <Clock />
    {#if !session}
      <div class="card setup">
        <h2>Configurar sesión</h2>
        <div class="field">
          <label class="label">Duración total (min)</label>
          <select bind:value={durationMin}>
            {#each [60,90,120,150,180] as d}<option>{d}</option>{/each}
          </select>
        </div>
        <div class="field">
          <label class="label">Duración de cada espacio</label>
          <div class="slot-opts">
            {#each slotOptions as opt}
              <button class="opt-btn" class:selected={slotMin===opt} on:click={() => slotMin=opt}>{opt} min</button>
            {/each}
          </div>
        </div>
        <div class="field">
          <label class="label">Aula (opcional)</label>
          <input bind:value={room} placeholder="Sala 201" />
        </div>
        {#if error}<p class="error">{error}</p>{/if}
        <button class="btn-primary" on:click={createSession} disabled={loading || !$currentSubject}>
          {loading ? 'Creando…' : 'Crear sesión'}
        </button>
      </div>
    {:else}
      <div class="card session-info">
        <div class="info-row"><span class="label-small">Aula</span><span>{session.room || '—'}</span></div>
        <div class="info-row">
          <span class="label-small">Espacios</span>
          <span>{Math.floor(session.duration_min / session.slot_min)} × {session.slot_min} min</span>
        </div>
        <div class="info-row">
          <span class="label-small">Estado</span>
          <span class="badge {session.active ? 'badge-green' : 'badge-yellow'}">{session.active ? 'Activa' : 'En espera'}</span>
        </div>
        {#if !session.active}
          <button class="btn-success" style="margin-top:1rem;width:100%" on:click={activate}>▶ Iniciar sesión</button>
        {/if}
      </div>
    {/if}
  </div>
  <div class="right">
    {#if session}
      <div class="slots-header">
        <h2>Espacios ({slots.filter(s=>s.student_id).length}/{slots.length} ocupados)</h2>
        <button class="btn-secondary" on:click={async () => { if(session) slots = await sessionsApi.slots(session.id); }}>↻</button>
      </div>
      <SlotGrid {slots} onReserve={reserveSlot} />
    {:else}
      <div class="empty">Crea una sesión para ver los espacios.</div>
    {/if}
  </div>
</div>

<style>
.page-header { margin-bottom: 1.5rem; }
h1 { font-size: 1.5rem; font-weight: 700; }
h2 { font-size: 1.1rem; font-weight: 600; margin-bottom: 1rem; }
.sub { color: var(--text2); font-size: 0.85rem; }
.two-col { display: grid; grid-template-columns: 340px 1fr; gap: 1.5rem; align-items: start; }
@media (max-width: 768px) { .two-col { grid-template-columns: 1fr; } }
.left { display: flex; flex-direction: column; gap: 1rem; }
.setup, .session-info { display: flex; flex-direction: column; gap: 0.75rem; }
.slot-opts { display: flex; gap: 0.5rem; flex-wrap: wrap; }
.opt-btn { background: var(--bg3); border: 1px solid var(--border); color: var(--text2); border-radius: 8px; padding: 0.4rem 0.9rem; }
.opt-btn.selected { background: var(--accent); color: #fff; border-color: var(--accent); }
.info-row { display: flex; justify-content: space-between; align-items: center; }
.label-small { color: var(--text2); font-size: 0.8rem; }
.slots-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 1rem; }
.empty { color: var(--text2); text-align: center; padding: 3rem; }
</style>
