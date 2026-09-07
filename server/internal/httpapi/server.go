// Package httpapi expose l'API HTTP et sert le frontend (SPA).
package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/attadeurtia/olivone/internal/auth"
	"github.com/attadeurtia/olivone/internal/config"
	"github.com/attadeurtia/olivone/internal/store"
	"github.com/attadeurtia/olivone/internal/users"
)

// Server porte les dépendances des handlers et le handler HTTP assemblé.
type Server struct {
	cfg     config.Config
	store   *store.Store
	users   *users.Service
	auth    *auth.Manager
	handler http.Handler
}

// New construit le serveur (services + routes). Le résultat implémente
// http.Handler.
func New(cfg config.Config, st *store.Store) *Server {
	s := &Server{
		cfg:   cfg,
		store: st,
		users: users.NewService(st.DB, cfg.MasterKey),
		auth:  auth.NewManager(st.DB, cfg.Env == "prod"),
	}
	s.routes()
	return s
}

// ServeHTTP délègue au handler assemblé.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

// EnsureBootstrap crée l'administrateur de bootstrap au premier démarrage.
func (s *Server) EnsureBootstrap() error {
	return s.users.EnsureAdmin(s.cfg.AdminEmail, s.cfg.AdminPassword)
}

func (s *Server) routes() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", s.handleHealth)

	// Authentification.
	mux.HandleFunc("POST /api/auth/login", s.handleLogin)
	mux.Handle("POST /api/auth/logout", s.requireAuth(http.HandlerFunc(s.handleLogout)))
	mux.Handle("GET /api/auth/me", s.requireAuth(http.HandlerFunc(s.handleMe)))

	// Réglages (utilisateur courant).
	mux.Handle("GET /api/settings", s.requireAuth(http.HandlerFunc(s.handleGetSettings)))
	mux.Handle("PUT /api/settings", s.requireAuth(http.HandlerFunc(s.handleUpdateSettings)))

	// Gestion des comptes (admin).
	mux.Handle("GET /api/users", s.requireAdmin(http.HandlerFunc(s.handleListUsers)))
	mux.Handle("POST /api/users", s.requireAdmin(http.HandlerFunc(s.handleCreateUser)))

	// Route API inconnue -> 404 JSON.
	mux.HandleFunc("/api/", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "route API inconnue"})
	})
	// Frontend (statique + fallback SPA).
	mux.Handle("/", s.staticHandler())

	s.handler = logging(recoverMW(mux))
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"version": s.cfg.Version,
		"env":     s.cfg.Env,
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// --- middleware -------------------------------------------------------------

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		slog.Info("http",
			"method", r.Method, "path", r.URL.Path,
			"status", sw.status, "dur", time.Since(start).String())
	})
}

func recoverMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic récupéré", "err", rec, "path", r.URL.Path)
				writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "erreur interne"})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (s *statusWriter) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}
