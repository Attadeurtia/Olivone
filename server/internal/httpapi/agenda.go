package httpapi

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/attadeurtia/olivone/internal/applications"
	"github.com/attadeurtia/olivone/internal/calendar"
)

// handleAgendaICS exporte les candidatures de l'utilisateur au format iCalendar.
func (s *Server) handleAgendaICS(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	apps, err := s.apps.List(u.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "erreur interne"})
		return
	}
	ics := calendar.Render(calendar.FromApplications(apps))
	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"olivone.ics\"")
	_, _ = io.WriteString(w, ics)
}

// handleCreateExternal ajoute une candidature faite hors logiciel.
func (s *Server) handleCreateExternal(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	var body struct {
		JobTitle       string `json:"job_title"`
		Company        string `json:"company"`
		RecipientEmail string `json:"recipient_email"`
		Status         string `json:"status"`
		SentAt         string `json:"sent_at"`
		Notes          string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "requête invalide"})
		return
	}
	app, err := s.apps.CreateExternal(u.ID, applications.ExternalInput{
		JobTitle:       body.JobTitle,
		Company:        body.Company,
		RecipientEmail: body.RecipientEmail,
		Status:         body.Status,
		SentAt:         body.SentAt,
		Notes:          body.Notes,
	})
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, app)
}
