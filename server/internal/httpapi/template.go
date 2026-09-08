package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/attadeurtia/olivone/internal/prompt"
)

// handleGetTemplate renvoie le template de prompt personnel de l'utilisateur.
func (s *Server) handleGetTemplate(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	content, err := s.apps.Users.GetDefaultTemplate(u.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "erreur interne"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"content_md": content})
}

// handleUpdateTemplate enregistre le template de prompt personnel.
func (s *Server) handleUpdateTemplate(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	var body struct {
		ContentMD string `json:"content_md"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "requête invalide"})
		return
	}
	if err := s.apps.Users.UpdateDefaultTemplate(u.ID, body.ContentMD); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "erreur interne"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"content_md": body.ContentMD})
}

// handleGetAppPrompt renvoie le contexte applicatif global (lecture seule).
func (s *Server) handleGetAppPrompt(w http.ResponseWriter, _ *http.Request) {
	content, _ := prompt.EnsureAppPrompt(s.cfg.DataDir)
	writeJSON(w, http.StatusOK, map[string]any{"content": content})
}
