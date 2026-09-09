package generation

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/attadeurtia/olivone/internal/prompt"
	"github.com/attadeurtia/olivone/internal/typst"
)

// Test d'intégration réel : nécessite une clé Mistral dans
// OLIVONE_MISTRAL_TEST_KEY. Ignoré sinon (aucune clé n'est stockée en dur).
func TestGenerateRealLetter(t *testing.T) {
	key := os.Getenv("OLIVONE_MISTRAL_TEST_KEY")
	if key == "" {
		t.Skip("OLIVONE_MISTRAL_TEST_KEY absent — test d'intégration ignoré")
	}

	in := prompt.Inputs{
		AppPrompt: "Tu rédiges des lettres de motivation en français, sobres et sincères. " +
			"Sortie en Markdown, commence par « Objet : ». Environ 200 mots. " +
			"N'invente aucune information absente du profil.",
		Profile: "Développeur backend, 5 ans d'expérience en Go et PostgreSQL, " +
			"goût pour les systèmes fiables et l'auto-hébergement.",
		OfferText: "ACME recherche un développeur backend Go pour concevoir des APIs " +
			"performantes et fiables. Compétences attendues : Go, SQL, Docker.",
		JobTitle: "Développeur Backend Go",
		Company:  "ACME",
	}

	letter, err := Generate(context.Background(), key, "mistral-small-latest", in)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(strings.ToLower(letter), "objet") {
		t.Errorf("la lettre devrait contenir un objet:\n%s", letter)
	}
	t.Logf("\n----- LETTRE GÉNÉRÉE -----\n%s\n--------------------------", letter)

	// Démo optionnelle : écrit le .md et rend le .pdf pour inspection.
	if out := os.Getenv("OLIVONE_DEMO_OUT"); out != "" {
		if err := os.WriteFile(out+".md", []byte(letter), 0o644); err != nil {
			t.Fatalf("écriture .md: %v", err)
		}
		if bin := os.Getenv("OLIVONE_TYPST_BIN"); bin != "" {
			if err := typst.New(bin).RenderLetter(context.Background(), typst.Letter{Markdown: letter}, out+".pdf"); err != nil {
				t.Fatalf("rendu PDF: %v", err)
			}
			t.Logf("PDF écrit : %s.pdf", out)
		}
	}
}
