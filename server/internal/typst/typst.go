// Package typst rend une lettre de motivation en PDF via le binaire Typst.
//
// La mise en page reproduit une lettre classique : un encart d'expéditeur en
// haut à gauche, les coordonnées du destinataire en haut à droite, la date, un
// objet mis en évidence, puis le corps. Le corps (Markdown produit par Mistral,
// éditable par l'utilisateur) est converti en balisage Typst puis inséré dans le
// gabarit. On évite ainsi toute dépendance à un paquet Typst externe (rendu
// hors-ligne, image Docker légère).
//
// La typographie utilise Spectral (Production Type, licence SIL OFL), un serif
// sobre et lisible embarqué dans le binaire : rendu identique partout, sans
// dépendre des polices du système hôte.
package typst

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// fontsFS embarque les fichiers de police Spectral (voir fonts/OFL.txt).
//
//go:embed fonts/*.ttf
var fontsFS embed.FS

// Palette du gabarit. Teinte crème très légère du papier + un accent sobre pour
// se démarquer sans surcharger (règle : ça doit rester subtil).
const (
	colPaper = "#fdfdfa" // blanc à peine réchauffé (fond de page)
	colInk   = "#1e1d1a" // texte principal (noir chaud)
	colMuted = "#5b574e" // lignes de coordonnées
)

// Party représente un bloc de coordonnées (expéditeur ou destinataire).
type Party struct {
	Lead  string   // petite ligne au-dessus du nom (ex. « À l'attention… »)
	Name  string   // intitulé en gras (nom de la personne / de l'entreprise)
	Lines []string // lignes de coordonnées, sous le nom
}

// Letter regroupe tout ce qu'il faut pour composer la lettre.
type Letter struct {
	Sender    Party     // expéditeur (haut-gauche)
	Recipient Party     // destinataire (haut-droite)
	Subject   string    // intitulé du poste, affiché en gras à la place de « Objet »
	City      string    // ville pour « Ville, le <date> » (optionnel)
	Date      time.Time // date de la lettre ; zéro => maintenant
	Markdown  string    // corps de la lettre (commence par « Objet : … »)
}

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

// RenderLetter compose la lettre et compile un PDF vers outPath.
func (r *Renderer) RenderLetter(ctx context.Context, l Letter, outPath string) error {
	fontDir, err := ensureFonts()
	if err != nil {
		return fmt.Errorf("polices: %w", err)
	}

	dir, err := os.MkdirTemp("", "olivone-typst-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(dir) }()

	when := l.Date
	if when.IsZero() {
		when = time.Now()
	}
	objet, body := splitObjet(l.Markdown)
	doc := buildDoc(l, frenchDate(when), objet, markdownToTypst(body))

	src := filepath.Join(dir, "letter.typ")
	if err := os.WriteFile(src, []byte(doc), 0o600); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o750); err != nil {
		return err
	}

	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, r.Bin, "compile",
		"--root", dir,
		"--font-path", fontDir,
		"--ignore-system-fonts", // rendu déterministe (dev == prod)
		src, outPath)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("typst compile: %v: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

// ensureFonts extrait les polices embarquées une seule fois par processus et
// renvoie le dossier à passer à Typst via --font-path.
var (
	fontOnce sync.Once
	fontDir  string
	fontErr  error
)

func ensureFonts() (string, error) {
	fontOnce.Do(func() {
		dir, err := os.MkdirTemp("", "olivone-fonts-*")
		if err != nil {
			fontErr = err
			return
		}
		entries, err := fontsFS.ReadDir("fonts")
		if err != nil {
			fontErr = err
			return
		}
		for _, e := range entries {
			b, err := fontsFS.ReadFile("fonts/" + e.Name())
			if err != nil {
				fontErr = err
				return
			}
			if err := os.WriteFile(filepath.Join(dir, e.Name()), b, 0o600); err != nil {
				fontErr = err
				return
			}
		}
		fontDir = dir
	})
	return fontDir, fontErr
}

// --- gabarit ----------------------------------------------------------------

func buildDoc(l Letter, date, objet, body string) string {
	var b strings.Builder

	b.WriteString(`#set page(paper: "a4", margin: (x: 2.3cm, top: 2cm, bottom: 2cm), fill: rgb("` + colPaper + `"))
#set text(font: ("Spectral", "Libertinus Serif"), size: 11.5pt, weight: 500, lang: "fr", region: "FR", fill: rgb("` + colInk + `"))
#set par(justify: true, leading: 0.72em, spacing: 1.1em)
#show heading: set text(weight: "bold")

`)

	// En-tête : expéditeur (gauche) + destinataire (droite), en texte simple.
	b.WriteString("#grid(\n  columns: (1.05fr, 0.95fr),\n  column-gutter: 1.1cm,\n  align: (left + top, right + top),\n")
	b.WriteString("  " + senderBlock(l.Sender) + ",\n")
	b.WriteString("  " + recipientBlock(l.Recipient) + ",\n)\n\n")

	// Date (ville, le …), alignée à droite.
	b.WriteString("#v(1.6em)\n#align(right)[#text(size: 10.5pt)[" + escapeTypst(dateLine(l.City, date)) + "]]\n\n")

	// Intitulé du poste, en gras (remplace la mention « Objet », sans encadré).
	title := strings.TrimSpace(l.Subject)
	if title == "" {
		title = strings.TrimSpace(objet)
	}
	if title != "" {
		b.WriteString("#v(1.4em)\n")
		b.WriteString(`#text(size: 15pt, weight: 700)[` + escapeTypst(title) + "]\n\n")
		b.WriteString("#v(1em)\n\n")
	} else {
		b.WriteString("#v(1.3em)\n\n")
	}

	// Corps.
	b.WriteString(body)
	b.WriteString("\n")
	return b.String()
}

// senderBlock : coordonnées de l'expéditeur, en texte simple (sans encadré).
func senderBlock(p Party) string {
	inner := partyContent(p, "13pt")
	return `[#[
    #set par(justify: false, leading: 0.55em)
` + inner + `
  ]]`
}

// recipientBlock : coordonnées du destinataire, en texte simple aligné à droite.
func recipientBlock(p Party) string {
	inner := partyContent(p, "11.5pt")
	return `[#[
    #set par(justify: false, leading: 0.55em)
    #set align(right)
` + inner + `
  ]]`
}

// partyContent produit les lignes Typst communes (lead, nom, coordonnées).
func partyContent(p Party, nameSize string) string {
	var parts []string
	if s := strings.TrimSpace(p.Lead); s != "" {
		parts = append(parts, `#text(size: 8.5pt, fill: rgb("`+colMuted+`"))[`+escapeTypst(s)+`]`)
	}
	if s := strings.TrimSpace(p.Name); s != "" {
		parts = append(parts, `#text(size: `+nameSize+`, weight: 700)[`+escapeTypst(s)+`]`)
	}
	var lines []string
	for _, ln := range p.Lines {
		if s := strings.TrimSpace(ln); s != "" {
			lines = append(lines, escapeTypst(s))
		}
	}
	if len(lines) > 0 {
		parts = append(parts, `#text(size: 9pt, fill: rgb("`+colMuted+`"))[`+strings.Join(lines, ` \ `)+`]`)
	}
	return strings.Join(parts, " \\\n")
}

func dateLine(city, date string) string {
	if strings.TrimSpace(city) != "" {
		return city + ", le " + date
	}
	return "Le " + date
}

// splitObjet isole la ligne « Objet : … » du reste du corps. Si absente, l'objet
// est vide et tout le Markdown constitue le corps.
var objetRe = regexp.MustCompile(`(?i)^\s*\*{0,2}\s*objet\s*[:：-]\s*(.+?)\s*\*{0,2}\s*$`)

func splitObjet(md string) (objet, body string) {
	lines := strings.Split(md, "\n")
	for i, line := range lines {
		if m := objetRe.FindStringSubmatch(line); m != nil {
			objet = strings.TrimSpace(m[1])
			rest := append(append([]string{}, lines[:i]...), lines[i+1:]...)
			body = strings.TrimLeft(strings.Join(rest, "\n"), "\n")
			return objet, body
		}
	}
	return "", md
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
