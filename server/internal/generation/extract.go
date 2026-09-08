package generation

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/attadeurtia/olivone/internal/mistral"
)

var emailRe = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)

// Fields regroupe les champs extraits d'une offre. Ils servent à pré-remplir le
// formulaire ; l'utilisateur peut toujours les corriger.
type Fields struct {
	JobTitle       string `json:"job_title"`
	Company        string `json:"company"`
	RecipientEmail string `json:"recipient_email"`
}

// ExtractFields tente d'extraire le poste, l'entreprise et l'e-mail destinataire.
// L'e-mail est d'abord cherché par expression régulière (fiable) ; le poste et
// l'entreprise via Mistral. Best-effort : jamais d'erreur, champs vides si échec.
func ExtractFields(ctx context.Context, apiKey, model, offerText string) Fields {
	f := Fields{RecipientEmail: emailRe.FindString(offerText)}

	if strings.TrimSpace(apiKey) == "" {
		return f // sans clé Mistral, on renvoie au moins l'e-mail trouvé
	}

	system := `Tu extrais des informations d'une offre d'emploi. ` +
		`Réponds UNIQUEMENT avec un objet JSON valide, sans texte ni balises autour, ` +
		`de la forme {"job_title":"","company":"","recipient_email":""}. ` +
		`Utilise une chaîne vide pour toute information absente. N'invente rien.`

	raw, err := mistral.New(apiKey, model).Complete(ctx, system, offerText)
	if err != nil {
		return f
	}
	var parsed Fields
	if err := json.Unmarshal([]byte(extractJSONObject(raw)), &parsed); err == nil {
		if parsed.JobTitle != "" {
			f.JobTitle = parsed.JobTitle
		}
		if parsed.Company != "" {
			f.Company = parsed.Company
		}
		if f.RecipientEmail == "" && parsed.RecipientEmail != "" {
			f.RecipientEmail = parsed.RecipientEmail
		}
	}
	return f
}

// extractJSONObject isole le premier objet JSON d'une chaîne (tolère d'éventuelles
// balises de code autour de la réponse).
func extractJSONObject(s string) string {
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}
