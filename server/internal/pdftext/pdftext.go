// Package pdftext extrait le texte d'un PDF « texte » (non scanné) en Go pur.
// Pour les PDF scannés (image), l'extraction renverra peu ou pas de texte ;
// un secours OCR via Mistral pourra être ajouté ultérieurement.
package pdftext

import (
	"io"
	"strings"

	"github.com/ledongthuc/pdf"
)

// ExtractText renvoie le texte brut d'un fichier PDF.
func ExtractText(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	reader, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	var b strings.Builder
	if _, err := io.Copy(&b, reader); err != nil {
		return "", err
	}
	return b.String(), nil
}
