// Package users gère les comptes et leurs réglages. À notre échelle (quelques
// utilisateurs), les comptes sont créés par un administrateur ; il n'y a pas
// d'inscription ouverte.
package users

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/attadeurtia/olivone/internal/auth"
)

// User représente un compte (sans le hash de mot de passe).
type User struct {
	ID          int64  `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	IsAdmin     bool   `json:"is_admin"`
	CreatedAt   string `json:"created_at"`
}

// Service porte l'accès base + la clé de chiffrement des secrets.
type Service struct {
	DB  *sql.DB
	Key []byte
}

// NewService crée le service utilisateurs.
func NewService(db *sql.DB, key []byte) *Service {
	return &Service{DB: db, Key: key}
}

// ErrEmailTaken est renvoyé lorsqu'une adresse est déjà utilisée.
var ErrEmailTaken = errors.New("adresse e-mail déjà utilisée")

// Count renvoie le nombre d'utilisateurs.
func (s *Service) Count() (int, error) {
	var n int
	err := s.DB.QueryRow(`SELECT COUNT(1) FROM users`).Scan(&n)
	return n, err
}

// Create ajoute un utilisateur avec sa ligne de réglages et un template par défaut.
func (s *Service) Create(email, password, displayName string, isAdmin bool) (User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || password == "" {
		return User{}, errors.New("email et mot de passe requis")
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return User{}, err
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return User{}, err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.Exec(
		`INSERT INTO users(email, password_hash, display_name, is_admin) VALUES(?,?,?,?)`,
		email, hash, displayName, boolToInt(isAdmin),
	)
	if err != nil {
		if strings.Contains(strings.ToUpper(err.Error()), "UNIQUE") {
			return User{}, ErrEmailTaken
		}
		return User{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return User{}, err
	}
	if _, err := tx.Exec(`INSERT INTO user_settings(user_id) VALUES(?)`, id); err != nil {
		return User{}, err
	}
	if _, err := tx.Exec(
		`INSERT INTO templates(user_id, name, content_md, is_default) VALUES(?, 'default', '', 1)`, id,
	); err != nil {
		return User{}, err
	}
	if err := tx.Commit(); err != nil {
		return User{}, err
	}
	return User{ID: id, Email: email, DisplayName: displayName, IsAdmin: isAdmin}, nil
}

// EnsureAdmin crée l'administrateur de bootstrap au premier démarrage, si aucun
// utilisateur n'existe et si les identifiants sont fournis dans l'environnement.
func (s *Service) EnsureAdmin(email, password string) error {
	n, err := s.Count()
	if err != nil {
		return err
	}
	if n > 0 || email == "" || password == "" {
		return nil
	}
	if _, err := s.Create(email, password, "Admin", true); err != nil {
		return fmt.Errorf("bootstrap admin: %w", err)
	}
	return nil
}

// GetByID récupère un utilisateur.
func (s *Service) GetByID(id int64) (User, error) {
	var u User
	var admin int
	err := s.DB.QueryRow(
		`SELECT id, email, display_name, is_admin, created_at FROM users WHERE id = ?`, id,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &admin, &u.CreatedAt)
	u.IsAdmin = admin == 1
	return u, err
}

// GetCredentials renvoie l'utilisateur et son hash de mot de passe (connexion).
func (s *Service) GetCredentials(email string) (User, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	var u User
	var admin int
	var hash string
	err := s.DB.QueryRow(
		`SELECT id, email, display_name, is_admin, created_at, password_hash FROM users WHERE email = ?`, email,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &admin, &u.CreatedAt, &hash)
	u.IsAdmin = admin == 1
	return u, hash, err
}

// GetDefaultTemplate renvoie le contenu du template de prompt par défaut de
// l'utilisateur (chaîne vide s'il n'y en a pas).
func (s *Service) GetDefaultTemplate(userID int64) (string, error) {
	var content string
	err := s.DB.QueryRow(
		`SELECT content_md FROM templates WHERE user_id = ? AND is_default = 1 ORDER BY id LIMIT 1`,
		userID,
	).Scan(&content)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return content, err
}

// UpdateDefaultTemplate met à jour le contenu du template de prompt par défaut
// de l'utilisateur (en crée un s'il n'existe pas).
func (s *Service) UpdateDefaultTemplate(userID int64, content string) error {
	res, err := s.DB.Exec(
		`UPDATE templates SET content_md=?, updated_at=datetime('now') WHERE user_id=? AND is_default=1`,
		content, userID,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		_, err = s.DB.Exec(
			`INSERT INTO templates(user_id, name, content_md, is_default) VALUES(?, 'default', ?, 1)`,
			userID, content,
		)
	}
	return err
}

// List renvoie tous les utilisateurs (usage admin).
func (s *Service) List() ([]User, error) {
	rows, err := s.DB.Query(`SELECT id, email, display_name, is_admin, created_at FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []User{}
	for rows.Next() {
		var u User
		var admin int
		if err := rows.Scan(&u.ID, &u.Email, &u.DisplayName, &admin, &u.CreatedAt); err != nil {
			return nil, err
		}
		u.IsAdmin = admin == 1
		out = append(out, u)
	}
	return out, rows.Err()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
