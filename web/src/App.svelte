<script lang="ts">
  import { api, type User } from './lib/api'
  import Login from './lib/Login.svelte'
  import Settings from './lib/Settings.svelte'
  import Users from './lib/Users.svelte'
  import Candidature from './lib/Candidature.svelte'

  type Theme = 'auto' | 'light' | 'dark'
  type View = 'candidature' | 'agenda' | 'prompt' | 'reglages' | 'utilisateurs'

  let user = $state<User | null>(null)
  let loading = $state(true)
  let view = $state<View>('candidature')
  let theme = $state<Theme>(readTheme())

  function readTheme(): Theme {
    try {
      return (localStorage.getItem('olivone-theme') as Theme) || 'auto'
    } catch {
      return 'auto'
    }
  }
  function setTheme(t: Theme) {
    theme = t
    try {
      localStorage.setItem('olivone-theme', t)
    } catch {}
  }
  $effect(() => {
    const r = document.documentElement
    if (theme === 'auto') r.removeAttribute('data-theme')
    else r.setAttribute('data-theme', theme)
  })

  async function loadMe() {
    loading = true
    try {
      user = await api.get('/api/auth/me')
    } catch {
      user = null
    }
    loading = false
  }
  loadMe()

  async function logout() {
    try {
      await api.post('/api/auth/logout')
    } catch {}
    user = null
  }

  const futureViews: { id: View; label: string; milestone: string }[] = [
    { id: 'agenda', label: 'Agenda', milestone: 'M4' },
    { id: 'prompt', label: 'Prompt', milestone: 'M4' },
  ]
</script>

{#if loading}
  <div class="splash muted">Chargement…</div>
{:else if !user}
  <Login onSuccess={(u) => (user = u)} {theme} {setTheme} />
{:else}
  <div class="shell">
    <header>
      <div class="brand">Olivone</div>
      <nav>
        <button class="tab" class:active={view === 'candidature'} onclick={() => (view = 'candidature')}>Candidature</button>
        {#each futureViews as v}
          <button
            class="tab"
            class:active={view === v.id}
            onclick={() => (view = v.id)}>{v.label}</button>
        {/each}
        <button class="tab" class:active={view === 'reglages'} onclick={() => (view = 'reglages')}>
          Réglages
        </button>
        {#if user.is_admin}
          <button
            class="tab"
            class:active={view === 'utilisateurs'}
            onclick={() => (view = 'utilisateurs')}>Utilisateurs</button>
        {/if}
      </nav>
      <div class="right">
        <select
          aria-label="Thème"
          value={theme}
          onchange={(e) => setTheme((e.currentTarget as HTMLSelectElement).value as Theme)}>
          <option value="auto">Auto</option>
          <option value="light">Jour</option>
          <option value="dark">Nuit</option>
        </select>
        <span class="who muted">{user.email}</span>
        <button class="btn btn-ghost" onclick={logout}>Déconnexion</button>
      </div>
    </header>

    <main>
      {#if view === 'candidature'}
        <Candidature />
      {:else if view === 'reglages'}
        <Settings />
      {:else if view === 'utilisateurs' && user.is_admin}
        <Users />
      {:else}
        {@const v = futureViews.find((f) => f.id === view)}
        <div class="card placeholder">
          <h2>{v?.label}</h2>
          <p class="muted">Écran prévu au jalon {v?.milestone}. Rien à afficher pour l'instant.</p>
        </div>
      {/if}
    </main>
  </div>
{/if}

<style>
  .splash {
    display: grid;
    place-items: center;
    min-height: 100vh;
  }
  .shell {
    max-width: 860px;
    margin: 0 auto;
    padding: 20px;
  }
  header {
    display: flex;
    align-items: center;
    gap: 14px;
    flex-wrap: wrap;
    padding-bottom: 16px;
    border-bottom: 1px solid var(--border);
  }
  .brand {
    font-size: 20px;
    font-weight: 700;
    letter-spacing: -0.02em;
  }
  nav {
    display: flex;
    gap: 4px;
    flex-wrap: wrap;
  }
  .tab {
    padding: 6px 12px;
    border: 1px solid transparent;
    border-radius: 8px;
    background: transparent;
    color: var(--muted);
    cursor: pointer;
    font: inherit;
    font-size: 14px;
  }
  .tab.active {
    color: var(--fg);
    border-color: var(--border);
    background: var(--card);
  }
  .right {
    margin-left: auto;
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .right select {
    width: auto;
  }
  .who {
    font-size: 13px;
  }
  main {
    padding-top: 20px;
  }
  .placeholder h2 {
    margin: 0 0 6px;
  }
</style>
