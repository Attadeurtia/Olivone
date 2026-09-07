<script lang="ts">
  // Frontend M0 : en-tête, sélecteur de thème jour/nuit, état de l'API.
  // Les écrans (Candidature, Agenda, Prompt, Réglages) arriveront aux
  // jalons suivants.
  type Theme = 'auto' | 'light' | 'dark'

  let theme = $state<Theme>(readTheme())
  let health = $state('vérification…')

  function readTheme(): Theme {
    try {
      return (localStorage.getItem('olivone-theme') as Theme) || 'auto'
    } catch {
      return 'auto'
    }
  }

  function applyTheme(t: Theme) {
    const root = document.documentElement
    if (t === 'auto') root.removeAttribute('data-theme')
    else root.setAttribute('data-theme', t)
  }

  function setTheme(t: Theme) {
    theme = t
    try {
      localStorage.setItem('olivone-theme', t)
    } catch {}
  }

  $effect(() => {
    applyTheme(theme)
  })

  async function ping() {
    try {
      const r = await fetch('/api/health')
      const j = await r.json()
      health = `${j.status} · v${j.version} · ${j.env}`
    } catch {
      health = 'hors ligne'
    }
  }
  ping()

  const views = ['Candidature', 'Agenda', 'Prompt', 'Réglages']
</script>

<div class="wrap">
  <header>
    <h1>Olivone</h1>
    <div class="themes" role="group" aria-label="Thème">
      <button class:active={theme === 'auto'} onclick={() => setTheme('auto')}>Auto</button>
      <button class:active={theme === 'light'} onclick={() => setTheme('light')}>Jour</button>
      <button class:active={theme === 'dark'} onclick={() => setTheme('dark')}>Nuit</button>
    </div>
  </header>

  <p class="status">API : <b>{health}</b></p>

  <nav>
    {#each views as v}
      <span class="pill">{v}</span>
    {/each}
  </nav>

  <p class="hint">Squelette M0 — les écrans arriveront aux jalons suivants.</p>
</div>

<style>
  .wrap {
    max-width: 720px;
    margin: 0 auto;
    padding: 40px 20px;
  }
  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
  }
  h1 {
    margin: 0;
    font-size: 28px;
    letter-spacing: -0.02em;
  }
  .themes {
    display: flex;
    gap: 6px;
  }
  button {
    padding: 7px 12px;
    border: 1px solid var(--border);
    border-radius: 9px;
    background: transparent;
    color: var(--fg);
    cursor: pointer;
    font-size: 13px;
  }
  button.active {
    background: var(--accent);
    color: #fff;
    border-color: var(--accent);
  }
  .status {
    margin-top: 22px;
    padding: 10px 12px;
    border: 1px solid var(--border);
    border-radius: 10px;
    background: color-mix(in srgb, var(--accent) 10%, transparent);
    font-size: 14px;
  }
  .status b {
    color: var(--accent);
  }
  nav {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 24px;
  }
  .pill {
    padding: 8px 14px;
    border: 1px solid var(--border);
    border-radius: 999px;
    background: var(--card);
    color: var(--muted);
    font-size: 14px;
  }
  .hint {
    margin-top: 24px;
    color: var(--muted);
    font-size: 13px;
  }
</style>
