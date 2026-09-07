// Package typst rend une lettre en PDF via le binaire Typst. Le corps Markdown
// (produit par Mistral, éditable par l'utilisateur) est converti en balisage
// Typst puis inséré dans un gabarit de lettre A4. On évite ainsi toute
// dépendance à un paquet Typst externe (rendu hors-ligne, image Docker légère).
package typst

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Renderer invoque le binaire Typst.
type Renderer struct {
	Bin string
}

// New crée un moteur de rendu. Bin vide => "typst" (résolu via le PATH).
func New(bin string) *Renderer {
	if bin == "" {
		bin = "typst"
	}
	return &Renderer{Bin: bin}
}

// RenderLetter convertit le Markdown en Typst, l'insère dans le gabarit et
// compile un PDF vers outPath.
func (r *Renderer) RenderLetter(ctx context.Context, markdown, outPath string) error {
	dir, err := os.MkdirTemp("", "olivone-typst-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(dir) }()

	doc := buildDoc(frenchDate(time.Now()), markdownToTypst(markdown))
	src := filepath.Join(dir, "letter.typ")
	if err := os.WriteFile(src, []byte(doc), 0o600); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o750); err != nil {
		return err
	}

	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, r.Bin, "compile", "--root", dir, src, outPath)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("typst compile: %v: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func buildDoc(date, body string) string {
	return `#set page(paper: "a4", margin: 2.5cm)
#set text(size: 11pt, lang: "fr")
#set par(justify: true, leading: 0.65em)
#show heading: set text(size: 12pt)

#align(right)[` + escapeTypst(date) + `]

#v(1.2em)

` + body + "\n"
}

// --- conversion Markdown -> Typst ------------------------------------------

var (
	headingRe = regexp.MustCompile(`^(#{1,6})\s+(.*)$`)
	listRe    = regexp.MustCompile(`^\s*[-*+]\s+(.*)$`)
	boldRe    = regexp.MustCompile(`\*\*(.+?)\*\*`)
)

// markdownToTypst convertit un sous-ensemble de Markdown (titres, listes, gras)
// en balisage Typst, en échappant les caractères spéciaux ailleurs.
func markdownToTypst(md string) string {
	lines := strings.Split(md, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimRight(line, " \t")
		if m := headingRe.FindStringSubmatch(line); m != nil {
			out = append(out, strings.Repeat("=", len(m[1]))+" "+inlineToTypst(m[2]))
			continue
		}
		if m := listRe.FindStringSubmatch(line); m != nil {
			out = append(out, "- "+inlineToTypst(m[1]))
			continue
		}
		out = append(out, inlineToTypst(line))
	}
	return strings.Join(out, "\n")
}

// inlineToTypst gère le gras **...** (=> *...* en Typst) et échappe le reste.
func inlineToTypst(s string) string {
	var b strings.Builder
	last := 0
	for _, loc := range boldRe.FindAllStringSubmatchIndex(s, -1) {
		b.WriteString(escapeTypst(s[last:loc[0]]))
		b.WriteString("*")
		b.WriteString(escapeTypst(s[loc[2]:loc[3]]))
		b.WriteString("*")
		last = loc[1]
	}
	b.WriteString(escapeTypst(s[last:]))
	return b.String()
}

// typstEscaper échappe les caractères spéciaux de Typst (backslash en premier).
var typstEscaper = strings.NewReplacer(
	`\`, `\\`,
	"`", "\\`",
	"#", "\\#",
	"$", "\\$",
	"*", "\\*",
	"_", "\\_",
	"<", "\\<",
	">", "\\>",
	"@", "\\@",
	"=", "\\=",
	"~", "\\~",
	"[", "\\[",
	"]", "\\]",
)

func escapeTypst(s string) string { return typstEscaper.Replace(s) }

var frMonths = [...]string{
	"janvier", "février", "mars", "avril", "mai", "juin",
	"juillet", "août", "septembre", "octobre", "novembre", "décembre",
}

func frenchDate(t time.Time) string {
	return fmt.Sprintf("%d %s %d", t.Day(), frMonths[t.Month()-1], t.Year())
}
