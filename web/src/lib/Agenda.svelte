<script lang="ts">
  import { api } from './api'

  type App = {
    id: number
    job_title: string
    company: string
    recipient_email: string
    status: string
    external: boolean
    created_at: string
    sent_at: string | null
    next_followup_at: string | null
    has_letter: boolean
  }

  const STATUSES = [
    'brouillon', 'envoyée', 'en_attente', 'relancée',
    'réponse_reçue', 'entretien', 'refusée', 'acceptée', 'archivée',
  ]

  let apps = $state<App[]>([])
  let error = $state('')
  let showAdd = $state(false)

  let jobTitle = $state('')
  let company = $state('')
  let recipient = $state('')
  let status = $state('envoyée')
  let sentAt = $state('')
  let notes = $state('')
  let adding = $state(false)

  function msg(e: unknown) {
    return e instanceof Error ? e.message : 'Erreur'
  }

  async function load() {
    try {
      apps = await api.get('/api/applications')
    } catch (e) {
      error = msg(e)
    }
  }
  load()

  async function setStatus(a: App, s: string) {
    try {
      await api.patch(`/api/applications/${a.id}`, { status: s })
      await load()
    } catch (e) {
      error = msg(e)
    }
  }

  async function addExternal(e: Event) {
    e.preventDefault()
    adding = true
    error = ''
    try {
      await api.post('/api/applications/external', {
        job_title: jobTitle,
        company,
        recipient_email: recipient,
        status,
        sent_at: sentAt,
        notes,
      })
      jobTitle = company = recipient = notes = sentAt = ''
      status = 'envoyée'
      showAdd = false
      await load()
    } catch (e) {
      error = msg(e)
    } finally {
      adding = false
    }
  }

  const fmtDate = (s: string | null) => (s ? s.slice(0, 10) : '—')
</script>

<div class="agenda">
  <div class="bar">
    <h3>Agenda des candidatures</h3>
    <div class="tools">
      <a class="btn btn-ghost" href="/api/agenda.ics" download>Exporter .ics</a>
      <button class="btn" type="button" onclick={() => (showAdd = !showAdd)}>
        {showAdd ? 'Annuler' : '+ Candidature externe'}
      </button>
    </div>
  </div>

  {#if showAdd}
    <form class="card addform" onsubmit={addExternal}>
      <p class="muted intro">Ajoute une candidature faite en dehors d'Olivone (pour la suivre ici).</p>
      <div class="row">
        <input bind:value={jobTitle} placeholder="Intitulé du poste" />
        <input bind:value={company} placeholder="Entreprise" />
      </div>
      <div class="row">
        <input type="email" bind:value={recipient} placeholder="Destinataire (optionnel)" />
        <select bind:value={status}>
          {#each STATUSES as s}<option value={s}>{s}</option>{/each}
        </select>
        <input type="date" bind:value={sentAt} title="Date d'envoi" />
      </div>
      <input bind:value={notes} placeholder="Notes (optionnel)" />
      {#if error}<p class="err">{error}</p>{/if}
      <button class="btn" type="submit" disabled={adding}>{adding ? 'Ajout…' : 'Ajouter'}</button>
    </form>
  {/if}

  {#if apps.length === 0}
    <p class="muted">Aucune candidature pour l'instant.</p>
  {:else}
    <div class="tablewrap card">
      <table>
        <thead>
          <tr>
            <th>Poste</th>
            <th>Entreprise</th>
            <th>Statut</th>
            <th>Envoyée</th>
            <th>Relance</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {#each apps as a}
            <tr>
              <td>
                {a.job_title || '(sans titre)'}
                {#if a.external}<span class="ext">externe</span>{/if}
              </td>
              <td class="muted">{a.company || '—'}</td>
              <td>
                <select value={a.status} onchange={(e) => setStatus(a, (e.currentTarget as HTMLSelectElement).value)}>
                  {#each STATUSES as s}<option value={s}>{s}</option>{/each}
                </select>
              </td>
              <td class="muted">{fmtDate(a.sent_at)}</td>
              <td class="muted">{fmtDate(a.next_followup_at)}</td>
              <td>{#if a.has_letter}<span title="Lettre générée">📄</span>{/if}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<style>
  .agenda {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  .bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    flex-wrap: wrap;
  }
  h3 {
    margin: 0;
    font-size: 16px;
  }
  .tools {
    display: flex;
    gap: 8px;
  }
  .tools .btn {
    text-decoration: none;
  }
  .addform {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }
  .intro {
    margin: 0;
    font-size: 13px;
  }
  .tablewrap {
    overflow-x: auto;
    padding: 4px;
  }
  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 14px;
  }
  th,
  td {
    text-align: left;
    padding: 9px 10px;
    border-bottom: 1px solid var(--border);
    vertical-align: top;
  }
  /* Colonnes courtes : pas de retour à la ligne. */
  td:nth-child(3),
  td:nth-child(4),
  td:nth-child(5),
  td:nth-child(6) {
    white-space: nowrap;
  }
  th {
    font-size: 12.5px;
    color: var(--muted);
    font-weight: 600;
  }
  td select {
    width: auto;
    padding: 5px 8px;
    font-size: 13px;
  }
  .ext {
    margin-left: 6px;
    font-size: 11px;
    padding: 1px 6px;
    border-radius: 999px;
    border: 1px solid var(--border);
    color: var(--muted);
  }
</style>
