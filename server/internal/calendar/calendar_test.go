package calendar

import (
	"strings"
	"testing"

	"github.com/attadeurtia/olivone/internal/applications"
)

func TestRenderValidVCalendar(t *testing.T) {
	sent := "2026-09-07 10:00:00"
	followup := "2026-09-21T10:00:00Z"
	apps := []applications.Application{
		{ID: 1, JobTitle: "Dev Go", Company: "ACME", SentAt: &sent, NextFollowupAt: &followup},
	}
	ics := Render(FromApplications(apps))

	for _, want := range []string{
		"BEGIN:VCALENDAR", "END:VCALENDAR", "BEGIN:VEVENT", "END:VEVENT",
		"DTSTART;VALUE=DATE:20260907", "Candidature envoyée",
		"DTSTART;VALUE=DATE:20260921", "Relance",
	} {
		if !strings.Contains(ics, want) {
			t.Errorf("le document iCal ne contient pas %q:\n%s", want, ics)
		}
	}
}

func TestFallbackToCreatedDate(t *testing.T) {
	created := "2026-09-01 08:00:00"
	apps := []applications.Application{{ID: 2, JobTitle: "X", CreatedAt: created}}
	ics := Render(FromApplications(apps))
	if !strings.Contains(ics, "DTSTART;VALUE=DATE:20260901") {
		t.Errorf("devrait utiliser la date de création à défaut d'envoi:\n%s", ics)
	}
}

func TestEscapeText(t *testing.T) {
	if got := escapeText("A; B, C"); got != `A\; B\, C` {
		t.Errorf("échappement iCal incorrect: %q", got)
	}
}
