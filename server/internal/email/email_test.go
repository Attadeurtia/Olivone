package email

import (
	"strings"
	"testing"
)

func TestNewMessageIDFormat(t *testing.T) {
	id := NewMessageID("moi@exemple.net")
	if !strings.HasSuffix(id, "@exemple.net") {
		t.Fatalf("le domaine devrait provenir de l'expéditeur: %q", id)
	}
	if strings.ContainsAny(id, "<> ") {
		t.Fatalf("le Message-ID ne doit pas contenir de chevrons/espaces: %q", id)
	}
}

func TestNewMessageIDIsUnique(t *testing.T) {
	if NewMessageID("a@b.c") == NewMessageID("a@b.c") {
		t.Fatal("deux Message-ID successifs doivent différer")
	}
}

func TestNewMessageIDFallbackDomain(t *testing.T) {
	id := NewMessageID("sans-arobase")
	if !strings.HasSuffix(id, "@olivone.local") {
		t.Fatalf("domaine de repli attendu: %q", id)
	}
}

func TestSendRequiresHost(t *testing.T) {
	if _, err := Send(Config{}, Message{To: "x@y.z"}); err == nil {
		t.Fatal("un hôte SMTP vide devrait provoquer une erreur")
	}
}
