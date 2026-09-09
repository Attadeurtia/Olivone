<script lang="ts">
  import { api } from './api'

  type Theme = 'auto' | 'light' | 'dark'
  let { theme, setTheme }: { theme: Theme; setTheme: (t: Theme) => void } = $props()

  let loaded = $state(false)
  let saving = $state(false)
  let error = $state('')
  let okMsg = $state('')

  // Champs non secrets.
  let defaultRecipient = $state('')
  let profileMd = $state('')
  // Identité affichée dans l'en-tête de la lettre (encart expéditeur).
  let senderName = $state('')
  let senderAddress = $state('')
  let senderPhone = $state('')
  let senderCity = $state('')
  let mistralModel = $state('mistral-small-latest')
  let smtpHost = $state('')
  let smtpPort = $state(587)
  let smtpUsername = $state('')
  let smtpFrom = $state('')
  let imapHost = $state('')
  let imapPort = $state(993)
  let imapUsername = $state('')
  let autoSend = $state(false)
  let followup = $state(false)
  let followupDays = $state(14)

  // Drapeaux « secret défini ».
  let mistralKeySet = $state(false)
  let smtpPwSet = $state(false)
  let imapPwSet = $state(false)

  // Saisies de secrets (écriture seule : envoyées seulement si remplies).
  let mistralKey = $state('')
  let smtpPassword = $state('')
  let imapPassword = $state('')

  // CV (PDF joint aux e-mails)
  let cvSet = $state(false)
  let cvBusy = $state(false)
  let cvInput: HTMLInputElement

  async function load() {
    try {
      const s = await api.get('/api/settings')
      defaultRecipient = s.default_recipient ?? ''
      profileMd = s.profile_md ?? ''
      senderName = s.sender_name ?? ''
      senderAddress = s.sender_address ?? ''
      senderPhone = s.sender_phone ?? ''
      senderCity = s.sender_city ?? ''
      mistralModel = s.mistral_model ?? 'mistral-small-latest'
      smtpHost = s.smtp_host ?? ''
      smtpPort = s.smtp_port ?? 587
      smtpUsername = s.smtp_username ?? ''
      smtpFrom = s.smtp_from ?? ''
      imapHost = s.imap_host ?? ''
      imapPort = s.imap_port ?? 993
      imapUsername = s.imap_username ?? ''
      autoSend = !!s.auto_send_enabled
      followup = !!s.followup_enabled
      followupDays = s.followup_interval_days ?? 14
      mistralKeySet = !!s.mistral_key_set
      smtpPwSet = !!s.smtp_password_set
      imapPwSet = !!s.imap_password_set
      cvSet = !!s.cv_set
      loaded = true
    } catch (err) {
      error = err instanceof Error ? err.message : 'Chargement impossible'
    }
  }
  load()

  async function save(e: Event) {
    e.preventDefault()
    saving = true
    error = ''
    okMsg = ''
    const payload: Record<string, unknown> = {
      default_recipient: defaultRecipient,
      profile_md: profileMd,
      sender_name: senderName,
      sender_address: senderAddress,
      sender_phone: senderPhone,
      sender_city: senderCity,
      mistral_model: mistralModel,
      smtp_host: smtpHost,
      smtp_port: Number(smtpPort),
      smtp_username: smtpUsername,
      smtp_from: smtpFrom,
      imap_host: imapHost,
      imap_port: Number(imapPort),
      imap_username: imapUsername,
      auto_send_enabled: autoSend,
      followup_enabled: followup,
      followup_interval_days: Number(followupDays),
    }
    // Secrets : uniquement s'ils ont été saisis.
    if (mistralKey) payload.mistral_api_key = mistralKey
    if (smtpPassword) payload.smtp_password = smtpPassword
    if (imapPassword) payload.imap_password = imapPassword

    try {
      const s = await api.put('/api/settings', payload)
      mistralKeySet = !!s.mistral_key_set
      smtpPwSet = !!s.smtp_password_set
      imapPwSet = !!s.imap_password_set
      mistralKey = smtpPassword = imapPassword = ''
      okMsg = 'Réglages enregistrés.'
    } catch (err) {
      error = err instanceof Error ? err.message : "Échec de l'enregistrement"
    } finally {
      saving = false
    }
  }

  async function uploadCV(ev: Event) {
    const input = ev.currentTarget as HTMLInputElement
    const f = input.files?.[0]
    if (!f) return
    cvBusy = true
    error = ''
    okMsg = ''
    try {
      const fd = new FormData()
      fd.set('cv', f)
      const r = await api.postForm('/api/cv', fd)
      cvSet = !!r.cv_set
      okMsg = 'CV enregistré.'
    } catch (e) {
      error = e instanceof Error ? e.message : "Échec de l'envoi du CV"
    } finally {
      cvBusy = false
      input.value = ''
    }
  }

  async function removeCV() {
    cvBusy = true
    error = ''
    try {
      await api.del('/api/cv')
      cvSet = false
      okMsg = 'CV retiré.'
    } catch (e) {
      error = e instanceof Error ? e.message : 'Échec'
    } finally {
      cvBusy = false
    }
  }

  const hint = (set: boolean) => (set ? 'défini — laisser vide pour conserver' : 'non défini')
</script>

{#if !loaded && !error}
  <p class="muted">Chargement des réglages…</p>
{:else}
  <form onsubmit={save}>
    {#if error}<p class="err">{error}</p>{/if}

    <section class="card">
      <h3>Apparence</h3>
      <div class="field">
        <label>Thème</label>
        <div class="themes">
          <button type="button" class:active={theme === 'auto'} onclick={() => setTheme('auto')}>Auto</button>
          <button type="button" class:active={theme === 'light'} onclick={() => setTheme('light')}>Jour</button>
          <button type="button" class:active={theme === 'dark'} onclick={() => setTheme('dark')}>Nuit</button>
        </div>
      </div>
    </section>

    <section class="card">
      <h3>Profil & candidature</h3>
      <div class="field">
        <label for="rec">Destinataire par défaut</label>
        <input id="rec" type="email" bind:value={defaultRecipient} placeholder="attadeurtia@vivaldi.net" />
      </div>
      <div class="field">
        <label for="prof">Profil / parcours (Markdown)</label>
        <textarea id="prof" bind:value={profileMd} placeholder="# Mon parcours&#10;- ..."></textarea>
      </div>
      <div class="field">
        <label>CV (PDF joint aux e-mails)</label>
        <div class="cvrow">
          <span class="muted">{cvSet ? '✓ CV enregistré' : 'Aucun CV'}</span>
          <button type="button" class="btn btn-ghost" onclick={() => cvInput.click()} disabled={cvBusy}>
            {cvSet ? 'Remplacer' : 'Ajouter'}
          </button>
          {#if cvSet}
            <a class="btn btn-ghost" href="/api/cv" target="_blank" rel="noopener">Voir</a>
            <button type="button" class="btn btn-ghost" onclick={removeCV} disabled={cvBusy}>Retirer</button>
          {/if}
          <input bind:this={cvInput} type="file" accept="application/pdf" hidden onchange={uploadCV} />
        </div>
      </div>
    </section>

    <section class="card">
      <h3>En-tête de lettre</h3>
      <p class="muted small">Ces coordonnées composent l'encart en haut à gauche du PDF. L'e-mail affiché est celui de votre compte.</p>
      <div class="row">
        <div class="field">
          <label for="sn">Nom affiché</label>
          <input id="sn" bind:value={senderName} placeholder="laisser vide = nom du compte" />
        </div>
        <div class="field">
          <label for="sc">Ville (pour la date)</label>
          <input id="sc" bind:value={senderCity} placeholder="Lausanne" />
        </div>
      </div>
      <div class="row">
        <div class="field">
          <label for="sadr">Adresse</label>
          <textarea id="sadr" class="addr" bind:value={senderAddress} placeholder="12 rue des Lilas&#10;1000 Lausanne"></textarea>
        </div>
        <div class="field">
          <label for="sph">Téléphone</label>
          <input id="sph" bind:value={senderPhone} placeholder="+41 79 123 45 67" />
        </div>
      </div>
    </section>

    <section class="card">
      <h3>Mistral</h3>
      <div class="row">
        <div class="field">
          <label for="mm">Modèle</label>
          <input id="mm" bind:value={mistralModel} placeholder="mistral-small-latest" />
        </div>
        <div class="field">
          <label for="mk">Clé API ({hint(mistralKeySet)})</label>
          <input id="mk" type="password" bind:value={mistralKey} placeholder="••••••••" autocomplete="off" />
        </div>
      </div>
    </section>

    <section class="card">
      <h3>E-mail — envoi (SMTP)</h3>
      <div class="row">
        <div class="field">
          <label for="sh">Serveur</label>
          <input id="sh" bind:value={smtpHost} placeholder="smtp.exemple.net" />
        </div>
        <div class="field">
          <label for="sp">Port</label>
          <input id="sp" type="number" bind:value={smtpPort} />
        </div>
      </div>
      <div class="row">
        <div class="field">
          <label for="su">Identifiant</label>
          <input id="su" bind:value={smtpUsername} autocomplete="off" />
        </div>
        <div class="field">
          <label for="spw">Mot de passe ({hint(smtpPwSet)})</label>
          <input id="spw" type="password" bind:value={smtpPassword} autocomplete="off" />
        </div>
      </div>
      <div class="field">
        <label for="sf">Adresse d'expédition</label>
        <input id="sf" type="email" bind:value={smtpFrom} placeholder="moi@exemple.net" />
      </div>
    </section>

    <section class="card">
      <h3>E-mail — réception (IMAP)</h3>
      <div class="row">
        <div class="field">
          <label for="ih">Serveur</label>
          <input id="ih" bind:value={imapHost} placeholder="imap.exemple.net" />
        </div>
        <div class="field">
          <label for="ip">Port</label>
          <input id="ip" type="number" bind:value={imapPort} />
        </div>
      </div>
      <div class="row">
        <div class="field">
          <label for="iu">Identifiant</label>
          <input id="iu" bind:value={imapUsername} autocomplete="off" />
        </div>
        <div class="field">
          <label for="ipw">Mot de passe ({hint(imapPwSet)})</label>
          <input id="ipw" type="password" bind:value={imapPassword} autocomplete="off" />
        </div>
      </div>
    </section>

    <section class="card">
      <h3>Automatisation</h3>
      <label class="check">
        <input type="checkbox" bind:checked={autoSend} />
        Envoi automatique des e-mails
        <span class="muted">(désactivé par défaut)</span>
      </label>
      <label class="check">
        <input type="checkbox" bind:checked={followup} />
        Relance automatique
        <span class="muted">(désactivée par défaut)</span>
      </label>
      <div class="field short">
        <label for="fd">Intervalle de relance (jours)</label>
        <input id="fd" type="number" bind:value={followupDays} min="1" />
      </div>
    </section>

    <div class="actions">
      <button class="btn" type="submit" disabled={saving}>{saving ? 'Enregistrement…' : 'Enregistrer'}</button>
      {#if okMsg}<span class="ok">{okMsg}</span>{/if}
    </div>
  </form>
{/if}

<style>
  form {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  h3 {
    margin: 0 0 12px;
    font-size: 15px;
  }
  .check {
    display: flex;
    align-items: center;
    gap: 8px;
    color: var(--fg);
    margin-bottom: 10px;
    font-size: 14px;
  }
  .check input {
    width: auto;
  }
  .short {
    max-width: 220px;
  }
  .small {
    font-size: 13px;
    margin: -4px 0 12px;
  }
  textarea.addr {
    min-height: 64px;
    resize: vertical;
  }
  .cvrow {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }
  .themes {
    display: flex;
    gap: 6px;
  }
  .themes button {
    padding: 7px 16px;
    border: 1px solid var(--border);
    border-radius: 9px;
    background: transparent;
    color: var(--fg);
    cursor: pointer;
    font: inherit;
    font-size: 13px;
  }
  .themes button.active {
    background: var(--accent);
    color: #fff;
    border-color: var(--accent);
  }
  .cvrow .btn {
    text-decoration: none;
  }
  .actions {
    display: flex;
    align-items: center;
    gap: 12px;
  }
</style>
