package auth

import "testing"

func TestHashAndVerify(t *testing.T) {
	h, err := HashPassword("mot-de-passe-costaud")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if h == "mot-de-passe-costaud" {
		t.Fatal("le mot de passe ne doit pas être stocké en clair")
	}
	ok, err := VerifyPassword(h, "mot-de-passe-costaud")
	if err != nil || !ok {
		t.Fatalf("le bon mot de passe devrait être accepté: ok=%v err=%v", ok, err)
	}
	bad, err := VerifyPassword(h, "mauvais")
	if err != nil {
		t.Fatalf("VerifyPassword: %v", err)
	}
	if bad {
		t.Fatal("un mauvais mot de passe ne doit pas être accepté")
	}
}
