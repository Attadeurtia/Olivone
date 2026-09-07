package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/attadeurtia/olivone/internal/config"
	"github.com/attadeurtia/olivone/internal/store"
)

func TestHealthEndpoint(t *testing.T) {
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	defer st.Close()

	h := New(config.Config{Env: "test", Version: "test-1"}, st)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, attendu 200", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("JSON invalide: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("status = %v, attendu \"ok\"", body["status"])
	}
	if body["version"] != "test-1" {
		t.Fatalf("version = %v, attendu \"test-1\"", body["version"])
	}
}

func TestUnknownAPIRouteReturns404JSON(t *testing.T) {
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	defer st.Close()

	h := New(config.Config{Env: "test", Version: "test-1"}, st)

	req := httptest.NewRequest(http.MethodGet, "/api/inconnu", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("code = %d, attendu 404", rec.Code)
	}
}

func TestRootServesPlaceholder(t *testing.T) {
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	defer st.Close()

	// WebDir vide => placeholder embarqué.
	h := New(config.Config{Env: "test", Version: "test-1"}, st)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, attendu 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct == "" {
		t.Fatalf("Content-Type manquant")
	}
}
