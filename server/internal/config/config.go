// Package config charge la configuration depuis les variables
// d'environnement. Aucune clé API de service n'est lue ici : les secrets
// applicatifs (Mistral, e-mail) sont propres à chaque utilisateur et stockés
// chiffrés en base. Seule la clé maître de chiffrement transite par l'env.
package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
)

// Config regroupe les paramètres de démarrage du serveur.
type Config struct {
	Env           string // "dev" | "prod"
	Bind          string // ex. ":8080"
	DataDir       string // dossier des données (base SQLite, fichiers)
	WebDir        string // dossier du frontend construit ; vide => placeholder embarqué
	TypstBin      string // chemin du binaire Typst (défaut "typst" via le PATH)
	BaseURL       string // URL publique (liens dans e-mails / iCal)
	MasterKey     []byte // 32 octets, pour chiffrer les secrets par utilisateur
	AdminEmail    string // compte admin créé au premier démarrage
	AdminPassword string // mot de passe du compte admin de bootstrap
	Version       string
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// Load lit l'environnement et valide la configuration.
func Load(version string) (Config, error) {
	c := Config{
		Env:           getenv("OLIVONE_ENV", "dev"),
		Bind:          getenv("OLIVONE_BIND", ":8080"),
		DataDir:       getenv("OLIVONE_DATA_DIR", "./data"),
		WebDir:        getenv("OLIVONE_WEB_DIR", ""),
		TypstBin:      getenv("OLIVONE_TYPST_BIN", "typst"),
		BaseURL:       getenv("OLIVONE_BASE_URL", "http://localhost:8080"),
		AdminEmail:    getenv("OLIVONE_ADMIN_EMAIL", ""),
		AdminPassword: getenv("OLIVONE_ADMIN_PASSWORD", ""),
		Version:       version,
	}

	key, err := loadMasterKey(c.Env)
	if err != nil {
		return Config{}, err
	}
	c.MasterKey = key
	return c, nil
}

// loadMasterKey lit OLIVONE_MASTER_KEY (base64, 32 octets).
// En production, la clé est obligatoire. En dev, si elle est absente, une clé
// éphémère est générée : les secrets chiffrés ne survivront pas au redémarrage,
// ce qui est acceptable tant qu'aucun secret n'est enregistré.
func loadMasterKey(env string) ([]byte, error) {
	raw := os.Getenv("OLIVONE_MASTER_KEY")
	if raw == "" {
		if env == "prod" {
			return nil, fmt.Errorf("OLIVONE_MASTER_KEY est requis en production (32 octets en base64 ; générez-en une avec: openssl rand -base64 32)")
		}
		k := make([]byte, 32)
		if _, err := rand.Read(k); err != nil {
			return nil, err
		}
		slog.Warn("OLIVONE_MASTER_KEY absent : génération d'une clé éphémère (dev uniquement)")
		return k, nil
	}
	k, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("OLIVONE_MASTER_KEY doit être encodée en base64: %w", err)
	}
	if len(k) != 32 {
		return nil, fmt.Errorf("OLIVONE_MASTER_KEY doit faire 32 octets une fois décodée (obtenu %d)", len(k))
	}
	return k, nil
}
