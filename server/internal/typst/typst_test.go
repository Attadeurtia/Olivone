package typst

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMarkdownToTypstBoldAndHeading(t *testing.T) {
	got := markdownToTypst("# Titre\n\n**Objet : Candidature**\n\nUn paragraphe.")
	if !strings.Contains(got, "= Titre") {
		t.Errorf("titre non converti:\n%s", got)
	}
	if !strings.Contains(got, "*Objet : Candidature*") {
		t.Errorf("gras non converti:\n%s", got)
	}
}

func TestMarkdownToTypstEscapesSpecials(t *testing.T) {
	got := inlineToTypst("Coût: 100$ #tag [Votre nom]")
	for _, want := range []string{`\$`, `\#`, `\[`, `\]`} {
		if !strings.Contains(got, want) {
			t.Errorf("caractère non échappé %q dans %q", want, got)
		}
	}
}

func TestMarkdownToTypstList(t *testing.T) {
	got := markdownToTypst("- premier\n- second")
	if strings.Count(got, "- ") != 2 {
		t.Errorf("liste non convertie:\n%s", got)
	}
}

// Rendu PDF réel : nécessite le binaire Typst (OLIVONE_TYPST_BIN ou PATH).
func TestRenderLetterProducesPDF(t *testing.T) {
	bin := os.Getenv("OLIVONE_TYPST_BIN")
	if bin == "" {
		bin = "typst"
	}
	if _, err := exec.LookPath(bin); err != nil {
		if _, e := os.Stat(bin); e != nil {
			t.Skip("binaire Typst introuvable — rendu PDF ignoré")
		}
	}

	out := filepath.Join(t.TempDir(), "letter.pdf")
	md := "**Objet : Candidature au poste de développeur**\n\n" +
		"Madame, Monsieur,\n\nJe vous écris au sujet de votre offre. " +
		"Coût & symboles : 100$ #test _x_\n\nBien cordialement,\n[Votre nom]"

	if err := New(bin).RenderLetter(context.Background(), md, out); err != nil {
		t.Fatalf("RenderLetter: %v", err)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("lecture PDF: %v", err)
	}
	if !bytes.HasPrefix(b, []byte("%PDF")) {
		t.Fatalf("le fichier n'est pas un PDF (entête: %x)", b[:8])
	}
	t.Logf("PDF généré : %d octets", len(b))
}
