package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/attadeurtia/olivone/internal/applications"
)

const maxUploadBytes = 25 << 20 // 25 Mo

func appID(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	return id, err == nil
}

// writeAppErr traduit les erreurs métier en codes HTTP.
func writeAppErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, applications.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "candidature introuvable"})
	case errors.Is(err, applications.ErrNoMistralKey):
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}
}

func (s *Server) handleCreateApplication(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	var in applications.CreateInput

	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "formulaire invalide"})
			return
		}
		in.JobTitle = r.FormValue("job_title")
		in.Company = r.FormValue("company")
		in.RecipientEmail = r.FormValue("recipient_email")
		in.OfferText = r.FormValue("offer_text")

		if file, _, err := r.FormFile("offer_pdf"); err == nil {
			defer func() { _ = file.Close() }()
			tmp, err := os.CreateTemp("", "olivone-offer-*.pdf")
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "fichier temporaire"})
				return
			}
			_, _ = io.Copy(tmp, io.LimitReader(file, maxUploadBytes))
			_ = tmp.Close()
			in.OfferPDFPath = tmp.Name()
			defer func() { _ = os.Remove(in.OfferPDFPath) }() // no-op si déjà déplacé
		}
	} else {
		var body struct {
			JobTitle       string `json:"job_title"`
			Company        string `json:"company"`
			RecipientEmail string `json:"recipient_email"`
			OfferText      string `json:"offer_text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "requête invalide"})
			return
		}
		in.JobTitle, in.Company = body.JobTitle, body.Company
		in.RecipientEmail, in.OfferText = body.RecipientEmail, body.OfferText
	}

	app, err := s.apps.Create(r.Context(), u.ID, in)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, app)
}

func (s *Server) handleListApplications(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	list, err := s.apps.List(u.ID)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleGetApplication(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	id, ok := appID(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "id invalide"})
		return
	}
	app, err := s.apps.Get(u.ID, id)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, app)
}

func (s *Server) handleGetLetterMD(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	id, ok := appID(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "id invalide"})
		return
	}
	md, err := s.apps.LetterMarkdown(u.ID, id)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"lettre-"+strconv.FormatInt(id, 10)+".md\"")
	_, _ = io.WriteString(w, md)
}

func (s *Server) handleGetLetterPDF(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	id, ok := appID(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "id invalide"})
		return
	}
	path, err := s.apps.LetterPDFPath(u.ID, id)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	http.ServeFile(w, r, path)
}

func (s *Server) handleUpdateLetter(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	id, ok := appID(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "id invalide"})
		return
	}
	var body struct {
		Markdown string `json:"markdown"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "requête invalide"})
		return
	}
	app, err := s.apps.UpdateLetter(r.Context(), u.ID, id, body.Markdown)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, app)
}

func (s *Server) handleRegenerate(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	id, ok := appID(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "id invalide"})
		return
	}
	app, err := s.apps.Regenerate(r.Context(), u.ID, id)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, app)
}

func (s *Server) handlePatchApplication(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	id, ok := appID(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "id invalide"})
		return
	}
	var body struct {
		Status string `json:"status"`
		Notes  string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "requête invalide"})
		return
	}
	app, err := s.apps.UpdateStatus(u.ID, id, body.Status, body.Notes)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, app)
}

func (s *Server) handleDeleteApplication(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	id, ok := appID(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "id invalide"})
		return
	}
	if err := s.apps.Delete(u.ID, id); err != nil {
		writeAppErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
