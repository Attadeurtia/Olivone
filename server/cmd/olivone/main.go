// Command olivone est le point d'entrée du serveur : chargement de la
// configuration, ouverture de la base SQLite, démarrage du serveur HTTP,
// puis arrêt propre sur signal (SIGINT/SIGTERM).
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/attadeurtia/olivone/internal/config"
	"github.com/attadeurtia/olivone/internal/httpapi"
	"github.com/attadeurtia/olivone/internal/store"
)

// version est renseignée à la compilation ; valeur par défaut pour le dev.
var version = "0.0.1-m0"

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	cfg, err := config.Load(version)
	if err != nil {
		slog.Error("configuration invalide", "err", err)
		os.Exit(1)
	}

	st, err := store.Open(cfg.DataDir)
	if err != nil {
		slog.Error("ouverture de la base", "err", err)
		os.Exit(1)
	}
	defer func() { _ = st.Close() }()

	handler := httpapi.New(cfg, st)
	srv := &http.Server{
		Addr:              cfg.Bind,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("démarrage",
			"bind", cfg.Bind, "env", cfg.Env, "version", version, "data", cfg.DataDir)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("serveur http", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	slog.Info("arrêt demandé, fermeture en cours…")

	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		slog.Error("arrêt du serveur", "err", err)
	}
	slog.Info("arrêté proprement")
}
