package mistral

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCompleteAgainstMock(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer clef-test" {
			t.Errorf("Authorization = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"Objet : Candidature\n\nMadame, Monsieur,"}}]}`))
	}))
	defer srv.Close()

	c := New("clef-test", "modele-test")
	c.BaseURL = srv.URL

	out, err := c.Complete(context.Background(), "système", "utilisateur")
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if !strings.Contains(out, "Objet") {
		t.Fatalf("réponse inattendue: %q", out)
	}
}

func TestCompleteMissingKey(t *testing.T) {
	c := New("", "m")
	if _, err := c.Complete(context.Background(), "s", "u"); err == nil {
		t.Fatal("une clé manquante devrait provoquer une erreur")
	}
}

func TestCompletePropagatesHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer srv.Close()

	c := New("mauvaise", "m")
	c.BaseURL = srv.URL
	if _, err := c.Complete(context.Background(), "s", "u"); err == nil {
		t.Fatal("un statut 401 devrait provoquer une erreur")
	}
}
