<script lang="ts">
  import { studentsApi } from '$lib/api/students';
  export let subjectId: string;
  export let onCreated: () => void = () => {};

  let fullName = '', email = '', code = '', error = '', loading = false;

  async function submit() {
    if (!fullName || !email) { error = 'Nombre y correo son requeridos'; return; }
    loading = true; error = '';
    try {
      const student = await studentsApi.create({ full_name: fullName, email, code });
      await studentsApi.enroll(subjectId, student.id);
      fullName = ''; email = ''; code = '';
      onCreated();
    } catch (e: any) { error = e.message; }
    finally { loading = false; }
  }
</script>

<form on:submit|preventDefault={submit}>
  <div class="row">
    <div class="field">
      <label class="label" for="fn">Nombre completo</label>
      <input id="fn" bind:value={fullName} placeholder="ARCE PAREJA SEBASTIAN" required />
    </div>
    <div class="field">
      <label class="label" for="em">Correo universitario</label>
      <input id="em" type="email" bind:value={email} placeholder="s.arce@universidad.edu.co" required />
    </div>
    <div class="field">
      <label class="label" for="co">Código</label>
      <input id="co" bind:value={code} placeholder="240220211012" />
    </div>
    <button type="submit" class="btn-primary" disabled={loading}>
      {loading ? 'Registrando…' : 'Registrar estudiante'}
    </button>
  </div>
  {#if error}<p class="error">{error}</p>{/if}
</form>

<style>
.row { display: flex; gap: 0.75rem; align-items: flex-end; flex-wrap: wrap; }
.row .field { flex: 1; min-width: 180px; }
</style>
