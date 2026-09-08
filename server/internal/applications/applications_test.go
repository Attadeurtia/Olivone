package applications

import (
	"testing"
	"time"

	"github.com/attadeurtia/olivone/internal/store"
	"github.com/attadeurtia/olivone/internal/users"
)

func TestDueFollowups(t *testing.T) {
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	us := users.NewService(st.DB, make([]byte, 32))
	svc := NewService(st.DB, us, nil, t.TempDir())

	// Utilisateur 1 : relance activée.
	u1, _ := us.Create("a@b.c", "pw", "", false)
	enabled := true
	if err := us.UpdateSettings(u1.ID, users.SettingsInput{FollowupEnabled: &enabled}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	// Utilisateur 2 : relance désactivée (défaut).
	u2, _ := us.Create("d@e.f", "pw", "", false)

	past := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	future := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	ins := `INSERT INTO applications(user_id, recipient_email, status, next_followup_at) VALUES(?, 'x@y.z', 'envoyée', ?)`

	// Due pour u1 (relance activée, échéance passée).
	if _, err := st.DB.Exec(ins, u1.ID, past); err != nil {
		t.Fatal(err)
	}
	// Échéance passée mais u2 n'a pas activé la relance -> exclue.
	if _, err := st.DB.Exec(ins, u2.ID, past); err != nil {
		t.Fatal(err)
	}
	// u1 mais échéance future -> exclue.
	if _, err := st.DB.Exec(ins, u1.ID, future); err != nil {
		t.Fatal(err)
	}

	due, err := svc.DueFollowups(time.Now())
	if err != nil {
		t.Fatalf("DueFollowups: %v", err)
	}
	if len(due) != 1 {
		t.Fatalf("attendu 1 relance due, obtenu %d", len(due))
	}
	if due[0].UserID != u1.ID {
		t.Fatalf("mauvais utilisateur retourné: %d", due[0].UserID)
	}
}
