package store

import (
	"testing"
)

func TestOpenAppliesMigrations(t *testing.T) {
	st, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer st.Close()

	// Les tables clés doivent exister.
	for _, table := range []string{"users", "user_settings", "templates", "applications", "messages", "sessions"} {
		var name string
		err := st.DB.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table,
		).Scan(&name)
		if err != nil {
			t.Fatalf("table %q absente: %v", table, err)
		}
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	st, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer st.Close()

	// Rejouer les migrations ne doit pas échouer.
	if err := Migrate(st.DB); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}

	var applied int
	if err := st.DB.QueryRow("SELECT COUNT(1) FROM schema_migrations").Scan(&applied); err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	if applied < 1 {
		t.Fatalf("attendu au moins 1 migration appliquée, obtenu %d", applied)
	}
}
