# Déploiement d'Olivone (Docker + SWAG)

Guide pas à pas pour héberger Olivone sur ton serveur, derrière SWAG (reverse
proxy + HTTPS).

## Prérequis

- Docker + Docker Compose sur le serveur.
- SWAG déjà en place (avec ton domaine et Let's Encrypt).
- Un sous-domaine pointant vers ton serveur, ex. `olivone.mondomaine.tld`.

## 1. Récupérer le code

```bash
git clone https://github.com/Attadeurtia/Olivone.git
cd Olivone
```

## 2. Réseau Docker partagé avec SWAG

Olivone et SWAG doivent partager un réseau Docker pour que SWAG joigne le
conteneur `olivone` par son nom. Trouve d'abord le nom de ton conteneur SWAG :

```bash
docker ps --format '{{.Names}}  {{.Image}}' | grep -i swag
```

### Option A — créer un réseau dédié `swag` (attendu par le compose)

```bash
docker network create swag
docker network connect swag NOM_DU_CONTENEUR_SWAG   # attache SWAG au réseau
```

### Option B — réutiliser le réseau existant de SWAG (recommandé)

```bash
docker inspect NOM_DU_CONTENEUR_SWAG -f '{{json .NetworkSettings.Networks}}'
```

Puis mets ce nom de réseau dans `.env` (rien à créer ni connecter) :

```bash
SWAG_NETWORK=swag-net    # remplace par le réseau affiché ci-dessus
```

> Sans cette étape, `docker compose up` échoue avec
> « network swag declared as external, but could not be found ».

## 3. Configurer les secrets (`.env`)

```bash
cp .env.example .env
```

Génère une clé maître (chiffrement des secrets par utilisateur) :

```bash
openssl rand -base64 32
```

Édite `.env` :

- `OLIVONE_BASE_URL=https://olivone.mondomaine.tld`
- `OLIVONE_MASTER_KEY=` ← la valeur générée ci-dessus
- `OLIVONE_ADMIN_EMAIL=` et `OLIVONE_ADMIN_PASSWORD=` ← compte admin de départ

> ⚠️ Sauvegarde `OLIVONE_MASTER_KEY` en lieu sûr. Si tu la perds ou la
> changes, les clés Mistral et mots de passe e-mail enregistrés deviennent
> **illisibles** (il faudra les re-saisir).

## 4. Configuration SWAG (proxy)

Copie l'exemple fourni et adapte le sous-domaine si besoin :

```bash
cp deploy/swag/olivone.subdomain.conf.sample \
   /chemin/vers/swag/config/nginx/proxy-confs/olivone.subdomain.conf
```

La conf pointe vers le conteneur `olivone` sur le port `8791` et autorise les
uploads jusqu'à 25 Mo (PDF d'offres / CV).

## 5. Construire et démarrer

```bash
docker compose up -d --build
```

Vérifie que le conteneur est sain :

```bash
docker compose ps          # STATUS doit afficher "healthy"
docker compose logs -f olivone
```

Recharge SWAG (ou attends son rechargement auto) puis ouvre
`https://olivone.mondomaine.tld`.

## 6. Première connexion & sécurisation

1. Connecte-toi avec `OLIVONE_ADMIN_EMAIL` / `OLIVONE_ADMIN_PASSWORD`.
2. Dans **Utilisateurs**, crée les comptes des autres utilisateurs (2–3 max).
3. Le mot de passe admin de bootstrap n'est utile qu'au premier démarrage ;
   tu peux ensuite retirer `OLIVONE_ADMIN_PASSWORD` de `.env`.

## 7. Configurer chaque utilisateur

Dans **Réglages**, chaque utilisateur renseigne (ses secrets sont **chiffrés**
en base, jamais renvoyés en clair) :

- **Clé API Mistral** (obligatoire pour générer les lettres).
- **E-mail — SMTP** (serveur, port, identifiant, mot de passe, expéditeur).
- **E-mail — IMAP** (pour plus tard : détection des réponses, jalon M5).
- **CV** (PDF joint aux e-mails).
- **En-tête de lettre** (nom, adresse, téléphone, ville) : compose l'encart
  d'expéditeur du PDF. Optionnel — le nom retombe sur le nom du compte et
  l'e-mail sur l'adresse du compte.
- **Profil / parcours**, et dans **Prompt** son template personnel.

## 8. Automatisations (désactivées par défaut)

Dans **Réglages → Automatisation** :

- **Envoi automatique** : OFF par défaut. Laissé OFF, rien ne part sans ton
  clic « Envoyer ».
- **Relance automatique** : OFF par défaut. Une fois activée, le planificateur
  envoie une relance à l'échéance (intervalle configurable, 14 j par défaut).

## 9. Sauvegardes

Toutes les données (base SQLite + PDF + lettres) sont dans le volume Docker
`olivone_data`. Pour une sauvegarde cohérente, arrête brièvement le conteneur :

```bash
docker compose stop olivone
docker run --rm -v olivone_olivone_data:/data -v "$PWD":/backup alpine \
  tar czf /backup/olivone-backup-$(date +%F).tar.gz -C /data .
docker compose start olivone
```

Restauration : décompresse l'archive dans le volume avec la commande inverse
(`tar xzf ... -C /data`), conteneur arrêté.

## 10. Mettre à jour (nouvelles fonctionnalités depuis GitHub)

Quand tu pousses de nouvelles fonctionnalités sur GitHub, mets à jour le
serveur ainsi, dans le dossier `Olivone` :

```bash
# (recommandé) sauvegarde d'abord — voir section 9
git pull                          # récupère le nouveau code
docker compose up -d --build      # reconstruit l'image et recrée le conteneur
```

- Le volume `olivone_data` est **conservé** : utilisateurs, candidatures,
  réglages, lettres et PDF restent en place.
- Les **migrations de base** éventuelles s'appliquent automatiquement au
  démarrage (ajout de tables/colonnes) — rien à faire.
- **Interruption** : quelques secondes (l'image se construit d'abord, puis le
  conteneur est remplacé).

Vérifie ensuite :

```bash
docker compose ps                 # STATUS "healthy"
docker compose logs -f olivone
```

Revenir en arrière : `git checkout <commit-précédent>` puis
`docker compose up -d --build` (le volume et tes données récentes restent).

## 11. Dépannage

- **Le conteneur redémarre / `unhealthy`** : `docker compose logs olivone`.
  Une erreur `OLIVONE_MASTER_KEY est requis` = variable manquante dans `.env`.
- **502 depuis SWAG** : olivone et SWAG ne sont pas sur le même réseau, ou le
  nom du conteneur/upstream ne correspond pas à la conf.
- **Les lettres ne se génèrent pas** : clé Mistral absente dans Réglages.

## Check-list sécurité

- [ ] `.env` n'est pas committé (il est ignoré par Git).
- [ ] `OLIVONE_MASTER_KEY` forte, générée aléatoirement, sauvegardée à part.
- [ ] Mot de passe admin fort ; comptes créés uniquement pour tes utilisateurs.
- [ ] Accès uniquement en HTTPS via SWAG (pas de port `8791` publié en prod).
- [ ] Sauvegardes régulières du volume `olivone_data`.
