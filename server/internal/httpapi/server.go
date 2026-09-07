// Package httpapi expose l'API HTTP et sert le frontend (SPA).
// À M0, seule la route /api/health est fonctionnelle ; le reste sert la page
// placeholder ou le frontend construit.
package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/attadeurtia/olivone/internal/config"
	"github.com/attadeurtia/olivone/internal/store"
)

// Server porte les dépendances des handlers.
type Server struct {
	cfg   config.Config
	store *store.Store
}

// New construit le handler HTTP complet (routes API + service du frontend).
func New(cfg config.Config, st *store.Store) http.Handler {
	s := &Server{cfg: cfg, store: st}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", s.handleHealth)
	// Toute autre route /api/ inconnue -> 404 JSON (au lieu du fallback SPA).
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "route API inconnue"})
	})
	// Catch-all : frontend (fichier statique ou fallback index.html).
	mux.Handle("/", s.staticHandler())

	return logging(recoverMW(mux))
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
