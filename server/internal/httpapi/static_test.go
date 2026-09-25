package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing/fstest"

	"testing"
)

// Le routage SPA ne doit jamais servir index.html à la place d'une ressource
// statique absente : sinon un « 200 text/html » se fait mettre en cache sous une
// URL de bundle immuable et casse durablement la page (CSS/JS refusés).
func TestSPAFileServer(t *testing.T) {
	fsys := fstest.MapFS{
		"index.html":              {Data: []byte("<!doctype html><html><body>ok</body></html>")},
		"assets/index-abc123.css": {Data: []byte("body{color:red}")},
	}
	h := spaFileServer(fsys)

	get := func(p string) *http.Response {
		req := httptest.NewRequest(http.MethodGet, p, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Result()
	}

	t.Run("asset présent : 200, type CSS, cache immuable", func(t *testing.T) {
		res := get("/assets/index-abc123.css")
		if res.StatusCode != 200 {
			t.Fatalf("status = %d, veut 200", res.StatusCode)
		}
		if ct := res.Header.Get("Content-Type"); ct != "text/css; charset=utf-8" {
			t.Errorf("Content-Type = %q, veut text/css", ct)
		}
		if cc := res.Header.Get("Cache-Control"); cc != "public, max-age=31536000, immutable" {
			t.Errorf("Cache-Control = %q, veut immutable", cc)
		}
	})

	t.Run("asset absent : 404, jamais l'index HTML", func(t *testing.T) {
		res := get("/assets/index-DOESNOTEXIST.css")
		if res.StatusCode != http.StatusNotFound {
			t.Fatalf("status = %d, veut 404 (sinon un 200 text/html est mis en cache)", res.StatusCode)
		}
	})

	t.Run("route SPA sans extension : 200 index.html, no-cache", func(t *testing.T) {
		res := get("/agenda")
		if res.StatusCode != 200 {
			t.Fatalf("status = %d, veut 200", res.StatusCode)
		}
		if ct := res.Header.Get("Content-Type"); ct != "text/html; charset=utf-8" {
			t.Errorf("Content-Type = %q, veut text/html", ct)
		}
		if cc := res.Header.Get("Cache-Control"); cc != "no-cache" {
			t.Errorf("Cache-Control = %q, veut no-cache", cc)
		}
	})

	t.Run("racine : index.html en no-cache", func(t *testing.T) {
		res := get("/")
		if res.StatusCode != 200 {
			t.Fatalf("status = %d, veut 200", res.StatusCode)
		}
		if cc := res.Header.Get("Cache-Control"); cc != "no-cache" {
			t.Errorf("Cache-Control = %q, veut no-cache", cc)
		}
	})
}
