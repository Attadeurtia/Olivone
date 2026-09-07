package prompt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildMessages(t *testing.T) {
	sys, usr := BuildMessages(Inputs{
		AppPrompt: "Tu rédiges des lettres.",
		Profile:   "Développeur Go, 5 ans.",
		OfferText: "Recherche dev backend.",
		JobTitle:  "Backend Engineer",
		Company:   "ACME",
	})
	if sys != "Tu rédiges des lettres." {
		t.Fatalf("système = %q", sys)
	}
	for _, want := range []string{"Backend Engineer — ACME", "Développeur Go", "Recherche dev backend"} {
		if !strings.Contains(usr, want) {
			t.Fatalf("message utilisateur, manque %q dans:\n%s", want, usr)
		}
	}
}

func TestBuildMessagesOmitsEmptySections(t *testing.T) {
	_, usr := BuildMessages(Inputs{OfferText: "x"})
	if strings.Contains(usr, "Consignes personnelles") || strings.Contains(usr, "profil") {
		t.Fatalf("les sections vides devraient être omises:\n%s", usr)
	}
}

func TestEnsureAppPromptCreatesAndReadsFile(t *testing.T) {
	dir := t.TempDir()
	c1, err := EnsureAppPrompt(dir)
	if err != nil || strings.TrimSpace(c1) == "" {
		t.Fatalf("EnsureAppPrompt: contenu vide, err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "prompt_app.md")); err != nil {
		t.Fatalf("fichier prompt_app.md non créé: %v", err)
	}
	// Le fichier prime : une édition doit être relue (éditable par fichier).
	if err := os.WriteFile(filepath.Join(dir, "prompt_app.md"), []byte("PERSO"), 0o640); err != nil {
		t.Fatal(err)
	}
	c2, _ := EnsureAppPrompt(dir)
	if c2 != "PERSO" {
		t.Fatalf("devrait relire le fichier édité, obtenu %q", c2)
	}
}
