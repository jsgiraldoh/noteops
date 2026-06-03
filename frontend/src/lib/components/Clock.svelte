<script lang="ts">
  import { clockStore } from '$lib/stores/clock';

  $: tick = $clockStore;
  $: remaining = tick?.remaining_sec ?? 0;
  $: minutes = Math.floor(remaining / 60);
  $: seconds = remaining % 60;
  $: urgent = remaining > 0 && remaining <= 300;
  $: expired = tick !== null && remaining === 0;
  $: pad = (n: number) => String(n).padStart(2, '0');

  function pct() {
    if (!tick) return 0;
    const total = tick.duration_min * 60;
    return Math.min(100, ((total - remaining) / total) * 100);
  }
</script>

<div class="clock-wrap" class:urgent class:expired>
  <div class="digits">{pad(minutes)}:{pad(seconds)}</div>
  <div class="label">
    {#if !tick}Sin sesión activa
    {:else if expired}Sesión finalizada
    {:else if urgent}¡Tiempo casi agotado!
    {:else}Tiempo restante
    {/if}
  </div>
  <div class="bar-track">
    <div class="bar-fill" style="width:{pct()}%"></div>
  </div>
</div>

<style>
.clock-wrap {
  text-align: center;
  padding: 2rem 1rem;
  border-radius: 16px;
  background: var(--bg2);
  border: 1px solid var(--border);
  transition: border-color 0.4s;
}
.digits {
  font-size: clamp(3rem, 8vw, 6rem);
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  letter-spacing: -2px;
  color: var(--text);
  transition: color 0.4s;
}
.label { color: var(--text2); font-size: 0.9rem; margin-top: 0.5rem; }
.bar-track { background: var(--bg3); border-radius: 999px; height: 6px; margin-top: 1.2rem; overflow: hidden; }
.bar-fill { height: 100%; background: var(--accent); border-radius: 999px; transition: width 1s linear, background 0.4s; }
.urgent .digits { color: var(--warn); }
.urgent .bar-fill { background: var(--warn); }
.urgent { border-color: var(--warn); }
.expired .digits { color: var(--danger); }
.expired .bar-fill { background: var(--danger); width: 100% !important; }
.expired { border-color: var(--danger); }
</style>
