<script lang="ts">
  import { api, type User } from './api'

  let list = $state<User[]>([])
  let error = $state('')
  let okMsg = $state('')
  let busy = $state(false)

  let email = $state('')
  let password = $state('')
  let displayName = $state('')
  let isAdmin = $state(false)

  async function load() {
    try {
      list = await api.get('/api/users')
    } catch (err) {
      error = err instanceof Error ? err.message : 'Chargement impossible'
    }
  }
  load()

  async function create(e: Event) {
    e.preventDefault()
    busy = true
    error = ''
    okMsg = ''
    try {
      await api.post('/api/users', {
        email,
        password,
        display_name: displayName,
        is_admin: isAdmin,
      })
      okMsg = `Compte créé : ${email}`
      email = password = displayName = ''
      isAdmin = false
      await load()
    } catch (err) {
      error = err instanceof Error ? err.message : 'Échec de la création'
    } finally {
      busy = false
    }
  }
</script>

<div class="grid">
  <section class="card">
    <h3>Utilisateurs</h3>
    {#if list.length === 0}
      <p class="muted">Aucun utilisateur.</p>
    {:else}
      <ul class="list">
        {#each list as u}
          <li>
            <span>{u.display_name || u.email}</span>
            <span class="muted">{u.email}</span>
            {#if u.is_admin}<span class="badge">admin</span>{/if}
          </li>
        {/each}
      </ul>
    {/if}
  </section>

  <section class="card">
    <h3>Créer un compte</h3>
    <form onsubmit={create}>
      <div class="field">
        <label for="ne">E-mail</label>
        <input id="ne" type="email" bind:value={email} required autocomplete="off" />
      </div>
      <div class="field">
        <label for="nn">Nom affiché</label>
        <input id="nn" bind:value={displayName} autocomplete="off" />
      </div>
      <div class="field">
        <label for="np">Mot de passe</label>
        <input id="np" type="password" bind:value={password} required autocomplete="new-password" />
      </div>
      <label class="check">
        <input type="checkbox" bind:checked={isAdmin} /> Administrateur
      </label>
      {#if error}<p class="err">{error}</p>{/if}
      {#if okMsg}<p class="ok">{okMsg}</p>{/if}
      <button class="btn" type="submit" disabled={busy}>{busy ? '…' : 'Créer'}</button>
    </form>
  </section>
</div>

<style>
  .grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
  }
  @media (max-width: 640px) {
    .grid {
      grid-template-columns: 1fr;
    }
  }
  h3 {
    margin: 0 0 12px;
    font-size: 15px;
  }
  .list {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .list li {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 0;
    border-bottom: 1px solid var(--border);
    font-size: 14px;
  }
  .badge {
    margin-left: auto;
    font-size: 12px;
    padding: 2px 8px;
    border-radius: 999px;
    background: var(--accent);
    color: #fff;
  }
  .check {
    display: flex;
    align-items: center;
    gap: 8px;
    color: var(--fg);
    margin-bottom: 12px;
    font-size: 14px;
  }
  .check input {
    width: auto;
  }
</style>
