# Olivone

Web app auto-hébergée pour **générer, envoyer, suivre et relancer** des
candidatures et lettres de motivation. Multi-utilisateurs, déployable en Docker
derrière SWAG.

- **Génération** de lettres via l'API **Mistral** (Markdown + PDF).
- **Suivi** des candidatures avec vue agenda exportable en **iCal**.
- **Surveillance** des réponses par **IMAP** et d'un dossier local de PDF.
- **Relance** automatique (désactivée par défaut).

> Envoi d'e-mails et relance automatique sont **désactivés par défaut**.

## Stack

Go (backend + service du SPA) · Svelte 5 (frontend) · SQLite · Typst (PDF, à
partir du jalon M2) · API Mistral · écosystème emersion (IMAP/SMTP).

## Prérequis

- **Docker** + Docker Compose (voie recommandée), ou
- **Go 1.27+** et **Node 20+** pour le développement natif.

## Lancer en local

### Option A — tout en Docker (recommandé)

```bash
docker compose -f docker-compose.dev.yml up --build
```

- App : http://localhost:8080
- Mailpit (faux SMTP + UI) : http://localhost:8025

### Option B — natif (itération rapide)

```bash
# Terminal 1 : backend Go (sert le placeholder si le frontend n'est pas construit)
make server

# Terminal 2 : frontend Svelte avec proxy /api -> :8080
cd web && npm install && npm run dev   # http://localhost:5173
```

## Tests

```bash
make test        # tests Go
```

## Configuration

Copier `.env.example` en `.env` et adapter. Variables clés :

| Variable | Rôle |
|---|---|
| `OLIVONE_MASTER_KEY` | Clé (base64, 32 o) de chiffrement des secrets par utilisateur. **Obligatoire en prod.** |
| `OLIVONE_DATA_DIR` | Dossier des données (base SQLite, fichiers). |
| `OLIVONE_WEB_DIR` | Dossier du frontend construit ; vide = placeholder embarqué. |
| `OLIVONE_ADMIN_EMAIL` / `OLIVONE_ADMIN_PASSWORD` | Compte admin de bootstrap. |
| `OLIVONE_BASE_URL` | URL publique (liens e-mails / iCal). |

Les clés Mistral et identifiants e-mail ne sont **pas** ici : ils sont propres à
chaque utilisateur et stockés **chiffrés** en base.

## Structure

```
server/   Backend Go (API, base, génération, envoi, suivi, monitoring)
web/      Frontend Svelte 5
deploy/   Exemple de configuration SWAG
```

## Déploiement

Pour héberger Olivone sur ton serveur derrière SWAG, suis le guide pas à pas :
**[DEPLOY.md](DEPLOY.md)** (réseau Docker, `.env`, SWAG, sauvegardes, mises à jour).

## Sécurité

- Les secrets ne sont jamais versionnés (`.env` est ignoré par Git).
- Les secrets par utilisateur sont chiffrés au repos (AES-GCM).

## Feuille de route

✅ M0 socle · M1 auth · M2 génération (Mistral+Typst) · M3 envoi ·
M4 agenda/iCal · M6 relance · finitions/déploiement.
⏳ M5 IMAP (détection des réponses) + dossier surveillé.
Intégrations France Travail / LinkedIn : prévues architecturalement (interface
`JobSource`), non développées.

## Licence

MIT — voir [LICENSE](LICENSE).
