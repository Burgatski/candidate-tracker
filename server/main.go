package main

import (
	"context"
	"database/sql"
	"errors"
	_ "embed"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/rs/cors"
	_ "modernc.org/sqlite"

	"github.com/remotely-works/frontend-challenge/server/config"
	"github.com/remotely-works/frontend-challenge/server/repository"
	"github.com/remotely-works/frontend-challenge/server/service"
	"github.com/remotely-works/frontend-challenge/server/transport"
)

//go:generate go run ./tools/generate-data

//go:embed data.json
var seedData []byte

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.Load()

	db, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		log.Error("open db", slog.String("err", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	db.SetMaxOpenConns(1) // SQLite supports one writer at a time
	db.ExecContext(ctx, `PRAGMA journal_mode=WAL`)
	db.ExecContext(ctx, `PRAGMA foreign_keys=ON`)

	if err := repository.Migrate(ctx, db); err != nil {
		log.Error("migrate", slog.String("err", err.Error()))
		os.Exit(1)
	}

	if err := repository.Seed(ctx, db, seedData); err != nil {
		log.Warn("seed skipped", slog.String("err", err.Error()))
	}

	repo := repository.New(db)
	svc := service.New(repo)
	handler := transport.NewHandler(log, svc)

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"*"},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
	})

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: c.Handler(handler),
	}

	done := make(chan struct{})
	go func() {
		log.Info("listening", slog.String("addr", ":"+cfg.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", slog.String("err", err.Error()))
		}
		close(done)
	}()

	<-ctx.Done()
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Error("shutdown", slog.String("err", err.Error()))
	}
	<-done
}
