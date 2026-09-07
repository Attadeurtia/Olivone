package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"
)

// CookieName est le nom du cookie de session.
const CookieName = "olivone_session"

// ErrNoSession indique une session absente ou expirée.
var ErrNoSession = errors.New("session absente ou expirée")

// Manager gère les sessions en base. Seul le hash du jeton est stocké.
type Manager struct {
	DB     *sql.DB
	Secure bool // cookie Secure (activé en production / HTTPS)
}

// NewManager crée un gestionnaire de sessions.
func NewManager(db *sql.DB, secure bool) *Manager {
	return &Manager{DB: db, Secure: secure}
}

// NewToken génère un jeton de session aléatoire (opaque, non deviné).
func NewToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// Create ouvre une session pour userID et renvoie le jeton en clair (à mettre
// dans le cookie ; seul son hash est persisté).
func (m *Manager) Create(userID int64, ttl time.Duration) (string, error) {
	token, err := NewToken()
	if err != nil {
		return "", err
	}
	expires := time.Now().Add(ttl).UTC().Format(time.RFC3339)
	if _, err := m.DB.Exec(
		`INSERT INTO sessions(token_hash, user_id, expires_at) VALUES(?,?,?)`,
		hashToken(token), userID, expires,
	); err != nil {
		return "", err
	}
	return token, nil
}

// UserID valide un jeton et renvoie l'identifiant utilisateur associé.
func (m *Manager) UserID(token string) (int64, error) {
	if token == "" {
		return 0, ErrNoSession
	}
	var userID int64
	var expiresAt string
	err := m.DB.QueryRow(
		`SELECT user_id, expires_at FROM sessions WHERE token_hash = ?`, hashToken(token),
	).Scan(&userID, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNoSession
	}
	if err != nil {
		return 0, err
	}
	if exp, err := time.Parse(time.RFC3339, expiresAt); err == nil && time.Now().After(exp) {
		_ = m.Delete(token)
		return 0, ErrNoSession
	}
	_, _ = m.DB.Exec(`UPDATE sessions SET last_seen = datetime('now') WHERE token_hash = ?`, hashToken(token))
	return userID, nil
}

// Delete supprime une session (déconnexion).
func (m *Manager) Delete(token string) error {
	_, err := m.DB.Exec(`DELETE FROM sessions WHERE token_hash = ?`, hashToken(token))
	return err
}
