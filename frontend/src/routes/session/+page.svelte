<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { currentSubject } from '$lib/stores/subject';
  import { activeSession } from '$lib/stores/session';
  import { sessionsApi, type Slot } from '$lib/api/sessions';
  import { connectClock, disconnectClock } from '$lib/stores/clock';
  import Clock from '$lib/components/Clock.svelte';
  import SlotGrid from '$lib/components/SlotGrid.svelte';

  let slots: Slot[] = [];
  let loading = false;
  let error = '';
  let durationMin = 120;
  let slotMin = 20;
  let room = '';
  const slotOptions = [5, 10, 20, 30, 60, 120];

  // Reserva inline — reemplaza prompt/alert
  let reservingSlot: Slot | null = null;
  let reserveStudentId = '';
  let reserveError = '';

  // Toast de notificación
  let toast = '';
  let toastTimer: ReturnType<typeof setTimeout> | null = null;

  // Polling en tiempo real
  let pollInterval: ReturnType<typeof setInterval> | null = null;

  onMount(async () => {
    if ($activeSession) {
      try {
        slots = await sessionsApi.slots($activeSession.id);
        if ($activeSession.active) {
          connectClock($activeSession.id);
          startPolling();
        }
      } catch { activeSession.set(null); }
    }
  });

  onDestroy(() => { stopPolling(); disconnectClock(); });

  function startPolling() {
    stopPolling();
    pollInterval = setInterval(async () => {
      if (!$activeSession) return;
      try {
        const fresh = await sessionsApi.slots($activeSession.id);
        const newlyReserved = fresh.filter(s =>
          !!s.student_id && !slots.find(o => o.id === s.id && !!o.student_id)
        );
        if (newlyReserved.length) {
          showToast(`Turno${newlyReserved.length > 1 ? 's' : ''} #${newlyReserved.map(s => s.number).join(', #')} reservado`);
        }
        slots = fresh;
      } catch {}
    }, 5000);
  }

  function stopPolling() {
    if (pollInterval) { clearInterval(pollInterval); pollInterval = null; }
  }

  function showToast(msg: string) {
    toast = msg;
    if (toastTimer) clearTimeout(toastTimer);
    toastTimer = setTimeout(() => { toast = ''; }, 3500);
  }

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
      activeSession.set(res.session);
      slots = res.slots;
    } catch (e: any) { error = e.message; }
    finally { loading = false; }
  }

  async function activate() {
    if (!$activeSession) return;
    await sessionsApi.activate($activeSession.id);
    activeSession.update(s => s ? { ...s, active: true } : s);
    connectClock($activeSession.id);
    startPolling();
  }

  async function stopSession() {
    if (!$activeSession) return;
    await sessionsApi.deactivate($activeSession.id);
    stopPolling();
    disconnectClock();
    activeSession.set(null);
    slots = [];
  }

  function newSession() {
    stopPolling();
    disconnectClock();
    activeSession.set(null);
    slots = [];
    error = '';
    reservingSlot = null;
  }

  async function refreshSlots() {
    if (!$activeSession) return;
    slots = await sessionsApi.slots($activeSession.id);
  }

  function openReserve(slot: Slot) {
    reservingSlot = slot;
    reserveStudentId = '';
    reserveError = '';
  }

  function cancelReserve() {
    reservingSlot = null;
    reserveStudentId = '';
    reserveError = '';
  }

  async function confirmReserve() {
    if (!reserveStudentId.trim() || !$activeSession || !reservingSlot) return;
    try {
      const updated = await sessionsApi.reserve($activeSession.id, reservingSlot.id, reserveStudentId.trim());
      slots = slots.map(s => s.id === updated.id ? updated : s);
      showToast(`Turno #${updated.number} reservado`);
      cancelReserve();
    } catch {
      reserveError = 'Espacio ya reservado o student_id inválido.';
    }
  }
</script>

<svelte:head><title>Sesión — NoteOPs</title></svelte:head>

{#if toast}
  <div class="toast">{toast}</div>
{/if}

<div class="page-header">
  <h1>Sesión de clase</h1>
  {#if $currentSubject}<p class="sub">{$currentSubject.name}</p>{/if}
</div>

<div class="two-col">
  <div class="left">
    <Clock />
    {#if !$activeSession}
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
        <div class="info-row"><span class="label-small">Aula</span><span>{$activeSession.room || '—'}</span></div>
        <div class="info-row">
          <span class="label-small">Espacios</span>
          <span>{Math.floor($activeSession.duration_min / $activeSession.slot_min)} × {$activeSession.slot_min} min</span>
        </div>
        <div class="info-row">
          <span class="label-small">Estado</span>
          <span class="badge {$activeSession.active ? 'badge-green' : 'badge-yellow'}">{$activeSession.active ? 'Activa' : 'En espera'}</span>
        </div>
        <div class="session-actions">
          {#if !$activeSession.active}
            <button class="btn-success" on:click={activate}>▶ Iniciar sesión</button>
          {:else}
            <button class="btn-danger" on:click={stopSession}>⏹ Finalizar sesión</button>
          {/if}
          <button class="btn-secondary" on:click={newSession}>+ Nueva sesión</button>
        </div>
      </div>

      {#if reservingSlot}
        <div class="card reserve-panel">
          <p class="reserve-title">Reservar turno <strong>#{reservingSlot.number}</strong></p>
          <input
            bind:value={reserveStudentId}
            placeholder="UUID del estudiante"
            on:keydown={e => e.key === 'Enter' && confirmReserve()}
          />
          {#if reserveError}<p class="error">{reserveError}</p>{/if}
          <div class="reserve-actions">
            <button class="btn-primary" on:click={confirmReserve} disabled={!reserveStudentId.trim()}>Reservar</button>
            <button class="btn-secondary" on:click={cancelReserve}>Cancelar</button>
          </div>
        </div>
      {/if}
    {/if}
  </div>

  <div class="right">
    {#if $activeSession}
      <div class="slots-header">
        <h2>Espacios ({slots.filter(s=>!!s.student_id).length}/{slots.length} ocupados)</h2>
        <button class="btn-secondary btn-sm" on:click={refreshSlots}>↻</button>
      </div>
      <SlotGrid {slots} onReserve={openReserve} />
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
@media (max-width: 900px) { .two-col { grid-template-columns: 1fr; } }
.left { display: flex; flex-direction: column; gap: 1rem; }
.setup, .session-info { display: flex; flex-direction: column; gap: 0.75rem; }
.slot-opts { display: flex; gap: 0.5rem; flex-wrap: wrap; }
.opt-btn { background: var(--bg3); border: 1px solid var(--border); color: var(--text2); border-radius: 8px; padding: 0.4rem 0.9rem; }
.opt-btn.selected { background: var(--accent); color: #fff; border-color: var(--accent); }
.info-row { display: flex; justify-content: space-between; align-items: center; }
.label-small { color: var(--text2); font-size: 0.8rem; }
.slots-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 1rem; }
.empty { color: var(--text2); text-align: center; padding: 3rem; }
.session-actions { display: flex; flex-direction: column; gap: 0.5rem; margin-top: 0.75rem; }
.session-actions button { width: 100%; }
.btn-danger { background: #dc2626; color: #fff; border: none; border-radius: 8px; padding: 0.6rem 1rem; font-size: 0.9rem; cursor: pointer; }
.btn-danger:hover { background: #b91c1c; }
.btn-sm { padding: 0.3rem 0.7rem; font-size: 0.8rem; }

/* Reserve panel */
.reserve-panel { display: flex; flex-direction: column; gap: 0.6rem; }
.reserve-title { font-size: 0.9rem; color: var(--text2); }
.reserve-title strong { color: var(--text); }
.reserve-actions { display: flex; gap: 0.5rem; }
.reserve-actions button { flex: 1; }

/* Toast */
.toast {
  position: fixed; bottom: 1.5rem; right: 1.5rem;
  background: #14401e; color: #4ade80;
  border: 1px solid #166534; border-radius: 8px;
  padding: 0.6rem 1.2rem; font-size: 0.85rem; font-weight: 500;
  z-index: 1000; animation: fadeIn 0.2s ease;
  box-shadow: 0 4px 12px rgba(0,0,0,0.4);
}
@keyframes fadeIn { from { opacity: 0; transform: translateY(8px); } to { opacity: 1; transform: translateY(0); } }
</style>
