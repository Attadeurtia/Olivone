package httpapi

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"path"
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
// placeholder embarquée. Le routage SPA renvoie index.html pour les routes
// applicatives (chemins sans extension) ; une ressource statique absente renvoie
// 404 (jamais l'index HTML à sa place — voir spaFileServer).
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
			// Fichier absent. Pour une ressource statique (chemin avec extension,
			// ex. /assets/index-xxxx.css), renvoyer 404 : NE JAMAIS servir l'index
			// HTML à sa place. Sinon le navigateur met en cache ce « 200 text/html »
			// sous une URL de bundle au nom immuable (haché), et la page reste
			// cassée (CSS/JS refusés pour cause de mauvais type MIME).
			if path.Ext(p) != "" {
				http.NotFound(w, r)
				return
			}
			// Route SPA (sans extension, ex. /agenda) -> coquille index.html.
			serveIndex(w, fsys)
			return
		}
		// Ressources versionnées (hash dans le nom) : cache long et immuable.
		// La coquille index.html doit au contraire toujours être revalidée pour
		// ne jamais pointer vers d'anciens hachages d'assets.
		if strings.HasPrefix(p, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else if p == "index.html" {
			w.Header().Set("Cache-Control", "no-cache")
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
	// La coquille SPA ne doit jamais être servie « périmée » : toujours
	// revalider pour récupérer les bons noms de bundles après un déploiement.
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}
