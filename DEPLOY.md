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

Olivone et SWAG doivent être sur le même réseau Docker pour que SWAG puisse
joindre le conteneur `olivone`.

```bash
docker network create swag
```

Attache aussi **SWAG** à ce réseau (dans le `docker-compose.yml` de SWAG,
ajoute `swag` sous `networks:` du service SWAG, avec en bas
`networks: { swag: { external: true, name: swag } }`), puis recrée SWAG.

> Si ton SWAG a déjà un réseau, tu peux réutiliser son nom : adapte
> `networks.swag.name` dans `docker-compose.yml` d'Olivone.

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

La conf pointe vers le conteneur `olivone` sur le port `8080` et autorise les
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

## 10. Mises à jour

```bash
git pull
docker compose up -d --build
```

Le volume `olivone_data` est conservé (tes données restent).

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
- [ ] Accès uniquement en HTTPS via SWAG (pas de port `8080` publié en prod).
- [ ] Sauvegardes régulières du volume `olivone_data`.
