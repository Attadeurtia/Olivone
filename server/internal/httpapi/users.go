package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/attadeurtia/olivone/internal/users"
)

func (s *Server) handleListUsers(w http.ResponseWriter, _ *http.Request) {
	list, err := s.users.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "erreur interne"})
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		DisplayName string `json:"display_name"`
		IsAdmin     bool   `json:"is_admin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "requête invalide"})
		return
	}
	u, err := s.users.Create(in.Email, in.Password, in.DisplayName, in.IsAdmin)
	if errors.Is(err, users.ErrEmailTaken) {
		writeJSON(w, http.StatusConflict, map[string]any{"error": users.ErrEmailTaken.Error()})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, u)
}
