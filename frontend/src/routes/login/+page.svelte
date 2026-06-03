<script lang="ts">
  import { goto } from '$app/navigation';
  import { login } from '$lib/api/auth';
  import { user } from '$lib/stores/auth';

  let email = '', password = '', error = '', loading = false;

  async function submit() {
    loading = true; error = '';
    try {
      const res = await login(email, password);
      user.set(res.user);
      goto('/');
    } catch (e: any) { error = e.message; }
    finally { loading = false; }
  }
</script>

<svelte:head><title>Iniciar sesión — NoteOPs</title></svelte:head>

<div class="wrap">
  <div class="box card">
    <div class="logo">NoteOPs</div>
    <p class="tagline">Sistema de gestión de notas académicas</p>
    <form on:submit|preventDefault={submit}>
      <div class="field">
        <label class="label" for="email">Correo</label>
        <input id="email" type="email" bind:value={email} placeholder="docente@universidad.edu.co" required />
      </div>
      <div class="field">
        <label class="label" for="pwd">Contraseña</label>
        <input id="pwd" type="password" bind:value={password} required />
      </div>
      {#if error}<p class="error">{error}</p>{/if}
      <button type="submit" class="btn-primary full" disabled={loading}>
        {loading ? 'Ingresando…' : 'Ingresar'}
      </button>
    </form>
  </div>
</div>

<style>
.wrap { min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 1rem; }
.box { width: min(420px, 100%); }
.logo { font-size: 2rem; font-weight: 800; color: var(--accent); margin-bottom: 0.3rem; }
.tagline { color: var(--text2); font-size: 0.9rem; margin-bottom: 2rem; }
.full { width: 100%; margin-top: 0.5rem; padding: 0.75rem; font-size: 1rem; }
</style>
