package crypto

import "testing"

func TestRoundTrip(t *testing.T) {
	key := make([]byte, 32)
	enc, err := EncryptString(key, "clé-mistral-secrète")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if enc == "clé-mistral-secrète" {
		t.Fatal("le texte ne doit pas être stocké en clair")
	}
	dec, err := DecryptString(key, enc)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if dec != "clé-mistral-secrète" {
		t.Fatalf("aller-retour incorrect: %q", dec)
	}
}

func TestEmptyStaysEmpty(t *testing.T) {
	key := make([]byte, 32)
	enc, err := EncryptString(key, "")
	if err != nil || enc != "" {
		t.Fatalf("chaîne vide: enc=%q err=%v", enc, err)
	}
}

func TestWrongKeyFails(t *testing.T) {
	k1 := make([]byte, 32)
	k2 := make([]byte, 32)
	k2[0] = 1
	enc, _ := EncryptString(k1, "x")
	if _, err := DecryptString(k2, enc); err == nil {
		t.Fatal("le déchiffrement avec une mauvaise clé devrait échouer")
	}
}
