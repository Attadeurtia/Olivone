package users

import (
	"testing"

	"github.com/attadeurtia/olivone/internal/store"
)

func newSvc(t *testing.T) *Service {
	t.Helper()
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return NewService(st.DB, make([]byte, 32))
}

func TestCreateNormalizesEmailAndCredentials(t *testing.T) {
	s := newSvc(t)
	u, err := s.Create("Test@Example.com", "pw", "Test", false)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if u.Email != "test@example.com" {
		t.Fatalf("email non normalisé: %q", u.Email)
	}
	gu, hash, err := s.GetCredentials("test@example.com")
	if err != nil {
		t.Fatalf("GetCredentials: %v", err)
	}
	if gu.ID != u.ID || hash == "" {
		t.Fatal("identifiants incorrects")
	}
}

func TestDuplicateEmailRejected(t *testing.T) {
	s := newSvc(t)
	if _, err := s.Create("a@b.c", "pw", "", false); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := s.Create("a@b.c", "pw2", "", false); err != ErrEmailTaken {
		t.Fatalf("attendu ErrEmailTaken, obtenu %v", err)
	}
}

func TestEnsureAdminIdempotent(t *testing.T) {
	s := newSvc(t)
	if err := s.EnsureAdmin("admin@x.y", "pw"); err != nil {
		t.Fatalf("EnsureAdmin: %v", err)
	}
	if n, _ := s.Count(); n != 1 {
		t.Fatalf("attendu 1 utilisateur, obtenu %d", n)
	}
	// Un second appel ne doit rien créer (des utilisateurs existent déjà).
	if err := s.EnsureAdmin("autre@x.y", "pw"); err != nil {
		t.Fatalf("EnsureAdmin (2): %v", err)
	}
	if n, _ := s.Count(); n != 1 {
		t.Fatalf("attendu toujours 1 utilisateur, obtenu %d", n)
	}
}

func TestSettingsSecretRoundTripAndDefaults(t *testing.T) {
	s := newSvc(t)
	u, _ := s.Create("a@b.c", "pw", "", false)

	key := "sk-mistral-123"
	if err := s.UpdateSettings(u.ID, SettingsInput{MistralAPIKey: &key}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	st, err := s.GetSettings(u.ID)
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	if st.MistralAPIKeyEnc == "" {
		t.Fatal("la clé n'a pas été stockée")
	}
	if st.MistralAPIKeyEnc == key {
		t.Fatal("la clé est stockée en clair !")
	}
	dec, err := s.DecryptSecret(st.MistralAPIKeyEnc)
	if err != nil {
		t.Fatalf("DecryptSecret: %v", err)
	}
	if dec != key {
		t.Fatalf("clé déchiffrée incorrecte: %q", dec)
	}
	// Valeurs par défaut : envoi auto et relance désactivés (règles 12/13).
	if st.AutoSendEnabled || st.FollowupEnabled {
		t.Fatal("auto-send et relance doivent être désactivés par défaut")
	}
}
