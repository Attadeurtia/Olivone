// Package calendar transforme les candidatures en événements iCalendar.
// C'est un transformateur en lecture seule : il ne modifie aucune donnée et
// dépend des modèles de tracking (applications), jamais l'inverse.
package calendar

import (
	"fmt"
	"strings"
	"time"

	"github.com/attadeurtia/olivone/internal/applications"
)

// Event est un événement « journée entière ».
type Event struct {
	UID     string
	Summary string
	Date    time.Time
}

// FromApplications construit les événements à partir des candidatures :
// date d'envoi (ou de création à défaut) et date de relance prévue.
func FromApplications(apps []applications.Application) []Event {
	var events []Event
	for _, a := range apps {
		label := a.JobTitle
		if label == "" {
			label = "(sans titre)"
		}
		if a.Company != "" {
			label += " (" + a.Company + ")"
		}

		if d, ok := parseDate(deref(a.SentAt)); ok {
			events = append(events, Event{uid(a.ID, "sent"), "Candidature envoyée — " + label, d})
		} else if d, ok := parseDate(a.CreatedAt); ok {
			events = append(events, Event{uid(a.ID, "created"), "Candidature — " + label, d})
		}
		if d, ok := parseDate(deref(a.NextFollowupAt)); ok {
			events = append(events, Event{uid(a.ID, "followup"), "Relance — " + label, d})
		}
	}
	return events
}

// Render produit un document iCalendar (VCALENDAR) valide.
func Render(events []Event) string {
	var b strings.Builder
	b.WriteString("BEGIN:VCALENDAR\r\n")
	b.WriteString("VERSION:2.0\r\n")
	b.WriteString("PRODID:-//Olivone//Candidatures//FR\r\n")
	b.WriteString("CALSCALE:GREGORIAN\r\n")

	stamp := time.Now().UTC().Format("20060102T150405Z")
	for _, e := range events {
		b.WriteString("BEGIN:VEVENT\r\n")
		fmt.Fprintf(&b, "UID:%s\r\n", e.UID)
		fmt.Fprintf(&b, "DTSTAMP:%s\r\n", stamp)
		fmt.Fprintf(&b, "DTSTART;VALUE=DATE:%s\r\n", e.Date.Format("20060102"))
		fmt.Fprintf(&b, "SUMMARY:%s\r\n", escapeText(e.Summary))
		b.WriteString("END:VEVENT\r\n")
	}
	b.WriteString("END:VCALENDAR\r\n")
	return b.String()
}

func uid(id int64, kind string) string {
	return fmt.Sprintf("olivone-%d-%s@olivone.local", id, kind)
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// parseDate accepte les formats stockés : datetime SQLite, RFC3339, date seule.
func parseDate(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{"2006-01-02 15:04:05", time.RFC3339, "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func escapeText(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, ";", `\;`)
	s = strings.ReplaceAll(s, ",", `\,`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	return s
}
