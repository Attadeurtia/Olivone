package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/attadeurtia/olivone/internal/config"
	"github.com/attadeurtia/olivone/internal/store"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	cfg := config.Config{Env: "test", Version: "test", MasterKey: make([]byte, 32)}
	return New(cfg, st)
}

func TestLoginFlow(t *testing.T) {
	s := newTestServer(t)
	if _, err := s.users.Create("user@x.y", "secret", "U", false); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Connexion.
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login",
		strings.NewReader(`{"email":"user@x.y","password":"secret"}`))
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login: code %d", rec.Code)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("aucun cookie de session posé")
	}

	// /me avec cookie.
	req2 := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	for _, c := range cookies {
		req2.AddCookie(c)
	}
	rec2 := httptest.NewRecorder()
	s.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("me: code %d", rec2.Code)
	}
	var me map[string]any
	if err := json.Unmarshal(rec2.Body.Bytes(), &me); err != nil {
		t.Fatalf("JSON /me: %v", err)
	}
	if me["email"] != "user@x.y" {
		t.Fatalf("me.email = %v", me["email"])
	}

	// /me sans cookie -> 401.
	rec3 := httptest.NewRecorder()
	s.ServeHTTP(rec3, httptest.NewRequest(http.MethodGet, "/api/auth/me", nil))
	if rec3.Code != http.StatusUnauthorized {
		t.Fatalf("attendu 401 sans cookie, obtenu %d", rec3.Code)
	}
}

func TestBadLoginRejected(t *testing.T) {
	s := newTestServer(t)
	if _, err := s.users.Create("user@x.y", "secret", "U", false); err != nil {
		t.Fatalf("Create: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login",
		strings.NewReader(`{"email":"user@x.y","password":"mauvais"}`))
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("attendu 401, obtenu %d", rec.Code)
	}
}

func TestSettingsRequiresAuth(t *testing.T) {
	s := newTestServer(t)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/settings", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("attendu 401, obtenu %d", rec.Code)
	}
}
