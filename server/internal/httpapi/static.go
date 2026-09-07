package httpapi

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// placeholderFS embarque une page minimale servie tant que le frontend n'est
// pas construit. Elle permet de lancer l'app (et de tester le thème + /api/health)
// sans étape de build côté frontend.
//
//go:embed embed/index.html
var placeholderFS embed.FS

// staticHandler sert le frontend depuis WebDir s'il existe, sinon la page
// placeholder embarquée. Dans les deux cas, le routage SPA renvoie index.html
// pour les chemins inconnus.
func (s *Server) staticHandler() http.Handler {
	if s.cfg.WebDir != "" {
		if _, err := os.Stat(filepath.Join(s.cfg.WebDir, "index.html")); err == nil {
			return spaFileServer(os.DirFS(s.cfg.WebDir))
		}
	}
	sub, _ := fs.Sub(placeholderFS, "embed")
	return spaFileServer(sub)
}

func spaFileServer(fsys fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(fsys))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, "/")
		if p == "" {
			p = "index.html"
		}
		if _, err := fs.Stat(fsys, p); err != nil {
			// Fichier absent -> fallback SPA vers index.html.
			serveIndex(w, fsys)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}

func serveIndex(w http.ResponseWriter, fsys fs.FS) {
	data, err := fs.ReadFile(fsys, "index.html")
	if err != nil {
		http.Error(w, "index.html introuvable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}
