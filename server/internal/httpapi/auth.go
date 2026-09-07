package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/attadeurtia/olivone/internal/auth"
	"github.com/attadeurtia/olivone/internal/users"
)

type ctxKey int

const userCtxKey ctxKey = 0

const sessionTTL = 30 * 24 * time.Hour

func userFromContext(ctx context.Context) (users.User, bool) {
	u, ok := ctx.Value(userCtxKey).(users.User)
	return u, ok
}

// requireAuth vérifie le cookie de session et injecte l'utilisateur en contexte.
func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(auth.CookieName)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "non authentifié"})
			return
		}
		uid, err := s.auth.UserID(c.Value)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "non authentifié"})
			return
		}
		u, err := s.users.GetByID(uid)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "non authentifié"})
			return
		}
		ctx := context.WithValue(r.Context(), userCtxKey, u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// requireAdmin exige un utilisateur authentifié ET administrateur.
func (s *Server) requireAdmin(next http.Handler) http.Handler {
	return s.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, _ := userFromContext(r.Context())
		if !u.IsAdmin {
			writeJSON(w, http.StatusForbidden, map[string]any{"error": "réservé à l'administrateur"})
			return
		}
		next.ServeHTTP(w, r)
	}))
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "requête invalide"})
		return
	}
	u, hash, err := s.users.GetCredentials(in.Email)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "identifiants invalides"})
		return
	}
	ok, err := auth.VerifyPassword(hash, in.Password)
	if err != nil || !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "identifiants invalides"})
		return
	}
	token, err := s.auth.Create(u.ID, sessionTTL)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "erreur interne"})
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     auth.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.auth.Secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(sessionTTL),
	})
	writeJSON(w, http.StatusOK, u)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(auth.CookieName); err == nil {
		_ = s.auth.Delete(c.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name: auth.CookieName, Value: "", Path: "/", HttpOnly: true, MaxAge: -1,
	})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	writeJSON(w, http.StatusOK, u)
}
