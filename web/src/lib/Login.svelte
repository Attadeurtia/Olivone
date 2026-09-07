<script lang="ts">
  import { api, type User } from './api'

  type Theme = 'auto' | 'light' | 'dark'
  let {
    onSuccess,
    theme,
    setTheme,
  }: {
    onSuccess: (u: User) => void
    theme: Theme
    setTheme: (t: Theme) => void
  } = $props()

  let email = $state('')
  let password = $state('')
  let error = $state('')
  let busy = $state(false)

  async function submit(e: Event) {
    e.preventDefault()
    busy = true
    error = ''
    try {
      const u = await api.post('/api/auth/login', { email, password })
      onSuccess(u)
    } catch (err) {
      error = err instanceof Error ? err.message : 'Échec de connexion'
    } finally {
      busy = false
    }
  }
</script>

<div class="wrap">
  <form class="card" onsubmit={submit}>
    <h1>Olivone</h1>
    <p class="muted sub">Connexion</p>

    <div class="field">
      <label for="email">Adresse e-mail</label>
      <input id="email" type="email" bind:value={email} required autocomplete="username" />
    </div>
    <div class="field">
      <label for="pw">Mot de passe</label>
      <input id="pw" type="password" bind:value={password} required autocomplete="current-password" />
    </div>

    {#if error}<p class="err">{error}</p>{/if}

    <button class="btn" type="submit" disabled={busy}>
      {busy ? '…' : 'Se connecter'}
    </button>

    <div class="themes">
      <button type="button" class:active={theme === 'auto'} onclick={() => setTheme('auto')}>Auto</button>
      <button type="button" class:active={theme === 'light'} onclick={() => setTheme('light')}>Jour</button>
      <button type="button" class:active={theme === 'dark'} onclick={() => setTheme('dark')}>Nuit</button>
    </div>
  </form>
</div>

<style>
  .wrap {
    min-height: 100vh;
    display: grid;
    place-items: center;
    padding: 20px;
  }
  .card {
    width: 100%;
    max-width: 360px;
  }
  h1 {
    margin: 0;
    font-size: 26px;
    letter-spacing: -0.02em;
  }
  .sub {
    margin: 2px 0 18px;
  }
  .btn {
    width: 100%;
  }
  .themes {
    display: flex;
    gap: 6px;
    margin-top: 16px;
  }
  .themes button {
    flex: 1;
    padding: 6px 8px;
    border: 1px solid var(--border);
    border-radius: 8px;
    background: transparent;
    color: var(--fg);
    cursor: pointer;
    font-size: 12.5px;
  }
  .themes button.active {
    background: var(--accent);
    color: #fff;
    border-color: var(--accent);
  }
</style>
