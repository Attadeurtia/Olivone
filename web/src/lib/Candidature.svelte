<script lang="ts">
  import { api } from './api'

  type App = {
    id: number
    job_title: string
    company: string
    recipient_email: string
    source_type: string
    status: string
    notes: string
    external: boolean
    created_at: string
    sent_at: string | null
    has_letter: boolean
    has_offer_pdf: boolean
  }

  const STATUSES = [
    'brouillon', 'envoyée', 'en_attente', 'relancée',
    'réponse_reçue', 'entretien', 'refusée', 'acceptée', 'archivée',
  ]

  let apps = $state<App[]>([])
  let selected = $state<App | null>(null)
  let pdfBust = $state(0)

  let mode = $state<'texte' | 'pdf'>('texte')
  let jobTitle = $state('')
  let company = $state('')
  let recipient = $state('')
  let offerText = $state('')
  let offerFile = $state<File | null>(null)
  let creating = $state(false)
  let error = $state('')
  let uploadInput: HTMLInputElement

  function msg(e: unknown) {
    return e instanceof Error ? e.message : 'Erreur'
  }

  async function loadApps() {
    try {
      apps = await api.get('/api/applications')
    } catch (e) {
      error = msg(e)
    }
  }
  loadApps()

  function select(app: App) {
    selected = app
    pdfBust = Date.now()
  }

  async function create(e: Event) {
    e.preventDefault()
    creating = true
    error = ''
    try {
      let app: App
      if (mode === 'pdf' && offerFile) {
        const fd = new FormData()
        fd.set('job_title', jobTitle)
        fd.set('company', company)
        fd.set('recipient_email', recipient)
        fd.set('offer_pdf', offerFile)
        app = await api.postForm('/api/applications', fd)
      } else {
        app = await api.post('/api/applications', {
          job_title: jobTitle,
          company,
          recipient_email: recipient,
          offer_text: offerText,
        })
      }
      await loadApps()
      select(app)
      offerText = ''
      offerFile = null
    } catch (e) {
      error = msg(e)
    } finally {
      creating = false
    }
  }

  async function regenerate() {
    if (!selected) return
    creating = true
    error = ''
    try {
      selected = await api.post(`/api/applications/${selected.id}/regenerate`)
      pdfBust = Date.now()
      await loadApps()
    } catch (e) {
      error = msg(e)
    } finally {
      creating = false
    }
  }

  async function onUploadCorrected(ev: Event) {
    const input = ev.currentTarget as HTMLInputElement
    const f = input.files?.[0]
    if (!f || !selected) return
    error = ''
    try {
      const text = await f.text()
      selected = await api.put(`/api/applications/${selected.id}/letter`, { markdown: text })
      pdfBust = Date.now()
    } catch (e) {
      error = msg(e)
    }
    input.value = ''
  }

  async function setStatus(status: string) {
    if (!selected) return
    try {
      selected = await api.patch(`/api/applications/${selected.id}`, { status })
      await loadApps()
    } catch (e) {
      error = msg(e)
    }
  }

  async function remove(app: App) {
    if (!confirm(`Supprimer la candidature « ${app.job_title || app.company} » ? Action irréversible.`)) return
    try {
      await api.del(`/api/applications/${app.id}`)
      if (selected?.id === app.id) selected = null
      await loadApps()
    } catch (e) {
      error = msg(e)
    }
  }
</script>

<div class="cand">
  <section class="col left">
    <div class="card">
      <h3>Nouvelle candidature</h3>
      <form onsubmit={create}>
        <div class="tabs2">
          <button type="button" class:active={mode === 'texte'} onclick={() => (mode = 'texte')}>Coller le texte</button>
          <button type="button" class:active={mode === 'pdf'} onclick={() => (mode = 'pdf')}>Importer un PDF</button>
        </div>

        {#if mode === 'texte'}
          <textarea bind:value={offerText} placeholder="Collez ici le texte de l'offre d'emploi…"></textarea>
        {:else}
          <input
            type="file"
            accept="application/pdf"
            onchange={(e) => (offerFile = (e.currentTarget as HTMLInputElement).files?.[0] ?? null)} />
        {/if}

        <div class="row">
          <input bind:value={jobTitle} placeholder="Intitulé du poste" />
          <input bind:value={company} placeholder="Entreprise" />
        </div>
        <input type="email" bind:value={recipient} placeholder="Destinataire (défaut : réglages)" />

        {#if error}<p class="err">{error}</p>{/if}
        <button class="btn" type="submit" disabled={creating}>
          {creating ? 'Génération…' : 'Générer la lettre'}
        </button>
      </form>
    </div>

    <div class="card">
      <h3>Mes candidatures</h3>
      {#if apps.length === 0}
        <p class="muted">Aucune candidature pour l'instant.</p>
      {:else}
        <ul class="list">
          {#each apps as a}
            <li class:sel={selected?.id === a.id}>
              <button class="pick" onclick={() => select(a)}>
                <span class="t">{a.job_title || '(sans titre)'}</span>
                <span class="muted co">{a.company}</span>
                <span class="badge">{a.status}</span>
              </button>
              <button class="del" title="Supprimer" onclick={() => remove(a)}>×</button>
            </li>
          {/each}
        </ul>
      {/if}
    </div>
  </section>

  <section class="col right">
    {#if !selected}
      <div class="card empty muted">
        Génère une lettre ou sélectionne une candidature pour la visualiser.
      </div>
    {:else}
      <div class="card">
        <div class="head">
          <div>
            <strong>{selected.job_title || '(sans titre)'}</strong>
            {#if selected.company}· {selected.company}{/if}
            <div class="muted small">→ {selected.recipient_email}</div>
          </div>
          <select
            aria-label="Statut"
            value={selected.status}
            onchange={(e) => setStatus((e.currentTarget as HTMLSelectElement).value)}>
            {#each STATUSES as s}<option value={s}>{s}</option>{/each}
          </select>
        </div>

        {#if selected.has_letter}
          <iframe class="pdf" title="Lettre de motivation" src={`/api/applications/${selected.id}/letter.pdf?t=${pdfBust}`}></iframe>
        {:else}
          <div class="empty muted">Aucune lettre pour cette candidature.</div>
        {/if}

        <div class="actions">
          <a class="btn btn-ghost" href={`/api/applications/${selected.id}/letter.md`} download>Télécharger .md</a>
          <button class="btn btn-ghost" type="button" onclick={() => uploadInput.click()}>Uploader une version corrigée</button>
          <input bind:this={uploadInput} type="file" accept=".md,text/markdown" hidden onchange={onUploadCorrected} />
          <button class="btn btn-ghost" type="button" onclick={regenerate} disabled={creating}>Régénérer</button>
          <button class="btn" type="button" disabled title="Envoi e-mail : jalon M3">Envoyer (M3)</button>
        </div>
      </div>
    {/if}
  </section>
</div>

<style>
  .cand {
    display: grid;
    grid-template-columns: 340px 1fr;
    gap: 16px;
    align-items: start;
  }
  @media (max-width: 760px) {
    .cand {
      grid-template-columns: 1fr;
    }
  }
  .col {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  h3 {
    margin: 0 0 12px;
    font-size: 15px;
  }
  .tabs2 {
    display: flex;
    gap: 6px;
    margin-bottom: 10px;
  }
  .tabs2 button {
    flex: 1;
    padding: 7px;
    border: 1px solid var(--border);
    border-radius: 8px;
    background: transparent;
    color: var(--fg);
    cursor: pointer;
    font-size: 13px;
  }
  .tabs2 button.active {
    background: var(--accent);
    color: #fff;
    border-color: var(--accent);
  }
  form {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }
  .list {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .list li {
    display: flex;
    align-items: center;
    border-bottom: 1px solid var(--border);
  }
  .list li.sel {
    background: color-mix(in srgb, var(--accent) 12%, transparent);
  }
  .pick {
    flex: 1;
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 9px 6px;
    background: transparent;
    border: none;
    color: var(--fg);
    cursor: pointer;
    text-align: left;
    font: inherit;
  }
  .pick .t {
    font-size: 14px;
  }
  .pick .co {
    font-size: 13px;
  }
  .badge {
    margin-left: auto;
    font-size: 11px;
    padding: 2px 8px;
    border-radius: 999px;
    border: 1px solid var(--border);
    color: var(--muted);
    white-space: nowrap;
  }
  .del {
    background: transparent;
    border: none;
    color: var(--muted);
    cursor: pointer;
    font-size: 18px;
    padding: 0 8px;
  }
  .head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 12px;
  }
  .head select {
    width: auto;
  }
  .small {
    font-size: 12.5px;
  }
  .pdf {
    width: 100%;
    height: 70vh;
    border: 1px solid var(--border);
    border-radius: 8px;
    background: #fff;
  }
  .empty {
    padding: 40px 16px;
    text-align: center;
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 12px;
  }
  .actions .btn {
    text-decoration: none;
  }
</style>
