package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/attadeurtia/olivone/internal/generation"
	"github.com/attadeurtia/olivone/internal/pdftext"
)

// handleExtract pré-remplit poste/entreprise/destinataire à partir d'une offre
// (texte collé ou PDF). Best-effort : renvoie ce qui a pu être extrait.
func (s *Server) handleExtract(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())

	var offerText string
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "formulaire invalide"})
			return
		}
		offerText = r.FormValue("offer_text")
		if file, _, err := r.FormFile("offer_pdf"); err == nil {
			defer func() { _ = file.Close() }()
			if tmp, e := os.CreateTemp("", "olivone-extract-*.pdf"); e == nil {
				_, _ = io.Copy(tmp, io.LimitReader(file, maxUploadBytes))
				_ = tmp.Close()
				if txt, e := pdftext.ExtractText(tmp.Name()); e == nil && strings.TrimSpace(txt) != "" {
					offerText = txt
				}
				_ = os.Remove(tmp.Name())
			}
		}
	} else {
		var body struct {
			OfferText string `json:"offer_text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "requête invalide"})
			return
		}
		offerText = body.OfferText
	}

	if strings.TrimSpace(offerText) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "offre vide"})
		return
	}

	settings, err := s.apps.Users.GetSettings(u.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "erreur interne"})
		return
	}
	apiKey, _ := s.apps.Users.DecryptSecret(settings.MistralAPIKeyEnc)

	f := generation.ExtractFields(r.Context(), apiKey, settings.MistralModel, offerText)
	writeJSON(w, http.StatusOK, f)
}
