// Package scheduling exécute périodiquement les relances dues. Il n'agit que
// sur les candidatures des utilisateurs ayant activé la relance automatique
// (désactivée par défaut) : sans activation, c'est un no-op.
package scheduling

import (
	"context"
	"log/slog"
	"time"

	"github.com/attadeurtia/olivone/internal/applications"
)

// Scheduler déclenche les relances à intervalle régulier.
type Scheduler struct {
	apps     *applications.Service
	interval time.Duration
}

// New crée le planificateur. Intervalle par défaut : 1 heure.
func New(apps *applications.Service, interval time.Duration) *Scheduler {
	if interval <= 0 {
		interval = time.Hour
	}
	return &Scheduler{apps: apps, interval: interval}
}

// Start lance la boucle jusqu'à l'annulation du contexte.
func (s *Scheduler) Start(ctx context.Context) {
	slog.Info("planificateur de relance démarré", "intervalle", s.interval.String())
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.runOnce() // premier passage au démarrage
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runOnce()
		}
	}
}

func (s *Scheduler) runOnce() {
	due, err := s.apps.DueFollowups(time.Now())
	if err != nil {
		slog.Error("relance: recherche des candidatures dues", "err", err)
		return
	}
	for _, d := range due {
		if err := s.apps.SendFollowup(d.UserID, d.AppID); err != nil {
			slog.Warn("relance: échec", "user", d.UserID, "app", d.AppID, "err", err)
			continue
		}
		slog.Info("relance envoyée", "user", d.UserID, "app", d.AppID)
	}
}
