// Package store gère l'accès à la base SQLite (ouverture, pragmas, migrations).
// On utilise le pilote pur Go modernc.org/sqlite (pas de CGO) pour garder un
// build et une image Docker simples et légers.
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Store enveloppe la connexion à la base.
type Store struct {
	DB *sql.DB
}

// Open crée le dossier de données si besoin, ouvre la base olivone.db,
// applique les pragmas (WAL, clés étrangères) puis les migrations.
func Open(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0o750); err != nil {
		return nil, fmt.Errorf("création du dossier de données: %w", err)
	}
	dbPath := filepath.Join(dataDir, "olivone.db")
	dsn := fmt.Sprintf(
		"file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)",
		dbPath,
	)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// À notre échelle (quelques utilisateurs), une seule connexion évite tout
	// verrou « database is locked » sans coût de performance notable.
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("connexion à la base: %w", err)
	}
	if err := Migrate(db); err != nil {
		return nil, fmt.Errorf("migrations: %w", err)
	}
	return &Store{DB: db}, nil
}

// Close ferme la connexion.
func (s *Store) Close() error { return s.DB.Close() }
