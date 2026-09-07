// Package promptasset embarque le prompt applicatif par défaut dans le binaire.
// Il est copié vers le volume de données au premier démarrage (data/prompt_app.md),
// où il devient éditable par simple modification de fichier.
package promptasset

import _ "embed"

//go:embed prompt_app.default.md
var Default string
