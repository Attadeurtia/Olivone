package httpapi

import (
	"io"
	"net/http"
	"os"
)

func (s *Server) handleSendApplication(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	id, ok := appID(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "id invalide"})
		return
	}
	app, err := s.apps.Send(u.ID, id, u.DisplayName)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, app)
}

func (s *Server) handleFollowup(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	id, ok := appID(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "id invalide"})
		return
	}
	if err := s.apps.SendFollowup(u.ID, id); err != nil {
		writeAppErr(w, err)
		return
	}
	app, err := s.apps.Get(u.ID, id)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, app)
}

func (s *Server) handleUploadCV(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "formulaire invalide"})
		return
	}
	file, _, err := r.FormFile("cv")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "fichier « cv » manquant"})
		return
	}
	defer func() { _ = file.Close() }()

	tmp, err := os.CreateTemp("", "olivone-cv-*.pdf")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "fichier temporaire"})
		return
	}
	_, _ = io.Copy(tmp, io.LimitReader(file, maxUploadBytes))
	_ = tmp.Close()

	if err := s.apps.SaveCV(u.ID, tmp.Name()); err != nil {
		_ = os.Remove(tmp.Name())
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "enregistrement du CV"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"cv_set": true})
}

func (s *Server) handleGetCV(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	if !s.apps.HasCV(u.ID) {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "aucun CV"})
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	http.ServeFile(w, r, s.apps.CVPath(u.ID))
}

func (s *Server) handleDeleteCV(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	if err := s.apps.DeleteCV(u.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "suppression du CV"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"cv_set": false})
}
