// Package generation orchestre la production d'une lettre : assemblage du
// prompt (couches) puis appel à Mistral. Le rendu PDF (Typst) et le stockage
// seront ajoutés dans la suite du jalon M2.
package generation

import (
	"context"

	"github.com/attadeurtia/olivone/internal/mistral"
	"github.com/attadeurtia/olivone/internal/prompt"
)

// Generate produit une lettre de motivation en Markdown à partir des couches
// d'entrée, via l'API Mistral.
func Generate(ctx context.Context, apiKey, model string, in prompt.Inputs) (string, error) {
	system, user := prompt.BuildMessages(in)
	return mistral.New(apiKey, model).Complete(ctx, system, user)
}
