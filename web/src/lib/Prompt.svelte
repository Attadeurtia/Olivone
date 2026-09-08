<script lang="ts">
  import { api } from './api'

  let content = $state('')
  let appPrompt = $state('')
  let loaded = $state(false)
  let saving = $state(false)
  let error = $state('')
  let okMsg = $state('')
  let showApp = $state(false)

  function msg(e: unknown) {
    return e instanceof Error ? e.message : 'Erreur'
  }

  async function load() {
    try {
      const t = await api.get('/api/template')
      content = t.content_md ?? ''
      const a = await api.get('/api/app-prompt')
      appPrompt = a.content ?? ''
      loaded = true
    } catch (e) {
      error = msg(e)
    }
  }
  load()

  async function save() {
    saving = true
    error = ''
    okMsg = ''
    try {
      await api.put('/api/template', { content_md: content })
      okMsg = 'Template enregistré.'
    } catch (e) {
      error = msg(e)
    } finally {
      saving = false
    }
  }
</script>

<div class="prompt">
  <section class="card">
    <h3>Mon template de lettre (Markdown)</h3>
    <p class="muted intro">
      Tes consignes personnelles de rédaction. Elles s'ajoutent au contexte
      global et à ton profil au moment de générer la lettre. Laisse vide pour
      t'appuyer uniquement sur le contexte global et ton profil.
    </p>
    <textarea
      bind:value={content}
      placeholder="Ex. : ton direct et concret ; mets en avant mes projets open-source ; accroche courte ; évite les formules toutes faites…"
    ></textarea>
    <div class="actions">
      <button class="btn" onclick={save} disabled={saving || !loaded}>
        {saving ? 'Enregistrement…' : 'Enregistrer'}
      </button>
      {#if okMsg}<span class="ok">{okMsg}</span>{/if}
      {#if error}<span class="err">{error}</span>{/if}
    </div>
  </section>

  <section class="card">
    <button class="disclosure" type="button" onclick={() => (showApp = !showApp)}>
      {showApp ? '▾' : '▸'} Contexte applicatif global (lecture seule)
    </button>
    {#if showApp}
      <p class="muted intro">
        Commun à toute l'application. Il n'est modifiable qu'en éditant le fichier
        <code>data/prompt_app.md</code> sur le serveur — jamais depuis l'interface.
      </p>
      <pre class="appprompt">{appPrompt}</pre>
    {/if}
  </section>
</div>

<style>
  .prompt {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  h3 {
    margin: 0 0 8px;
    font-size: 15px;
  }
  .intro {
    margin: 0 0 12px;
    font-size: 13px;
  }
  textarea {
    min-height: 260px;
    font-family: ui-monospace, 'SFMono-Regular', Menlo, monospace;
    font-size: 13px;
  }
  .actions {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-top: 12px;
  }
  .disclosure {
    background: transparent;
    border: none;
    color: var(--fg);
    cursor: pointer;
    font: inherit;
    font-size: 15px;
    font-weight: 600;
    padding: 0;
  }
  code {
    background: color-mix(in srgb, var(--accent) 12%, transparent);
    padding: 1px 5px;
    border-radius: 5px;
    font-size: 12.5px;
  }
  .appprompt {
    margin: 10px 0 0;
    padding: 12px;
    background: var(--bg);
    border: 1px solid var(--border);
    border-radius: 8px;
    white-space: pre-wrap;
    word-break: break-word;
    font-size: 12.5px;
    line-height: 1.5;
    max-height: 340px;
    overflow: auto;
  }
</style>
