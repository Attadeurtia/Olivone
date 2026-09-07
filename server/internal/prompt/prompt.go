// Package prompt assemble les couches d'entrée de la génération : le prompt
// applicatif (contexte global, éditable par fichier) sert de message système ;
// le template personnel, le profil et l'offre forment le message utilisateur.
package prompt

import (
	"os"
	"path/filepath"
	"strings"

	promptasset "github.com/attadeurtia/olivone/prompt"
)

// Inputs regroupe les couches d'entrée de la génération.
type Inputs struct {
	AppPrompt string // contexte applicatif (message système)
	Template  string // consignes personnelles de l'utilisateur
	Profile   string // profil / parcours
	OfferText string // texte de l'offre d'emploi
	JobTitle  string
	Company   string
	Recipient string
}

// EnsureAppPrompt copie le prompt applicatif par défaut vers dataDir/prompt_app.md
// au premier démarrage, puis renvoie le contenu du fichier. Éditer ce fichier
// modifie le comportement sans reconstruire le binaire.
func EnsureAppPrompt(dataDir string) (string, error) {
	path := filepath.Join(dataDir, "prompt_app.md")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.WriteFile(path, []byte(promptasset.Default), 0o640); err != nil {
			return "", err
		}
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// BuildMessages produit le message système (contexte applicatif) et le message
// utilisateur (sections template + profil + offre). Les sections vides sont omises.
func BuildMessages(in Inputs) (system, user string) {
	system = strings.TrimSpace(in.AppPrompt)

	var b strings.Builder
	section := func(title, body string) {
		body = strings.TrimSpace(body)
		if body == "" {
			return
		}
		b.WriteString("## ")
		b.WriteString(title)
		b.WriteString("\n")
		b.WriteString(body)
		b.WriteString("\n\n")
	}

	poste := strings.TrimSpace(in.JobTitle)
	if c := strings.TrimSpace(in.Company); c != "" {
		if poste != "" {
			poste += " — " + c
		} else {
			poste = c
		}
	}

	section("Poste visé", poste)
	section("Mon profil et mon parcours", in.Profile)
	section("Consignes personnelles", in.Template)
	section("Offre d'emploi", in.OfferText)

	return system, strings.TrimSpace(b.String())
}
