package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rs/cors"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	db, err := ReadDatabase()
	if err != nil {
		log.Error("could not read db", slog.String("err", err.Error()))
		os.Exit(1)
	}

	cors := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"*"},
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodOptions},
	})

	s := &http.Server{
		Addr:    ":8080",
		Handler: cors.Handler(NewServer(log, db)),
	}

	var done = make(chan struct{})
	go func() {
		log.Info("listenning on :8080")
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("error running server", slog.String("err", err.Error()))
		}

		close(done)
	}()

	<-ctx.Done()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := s.Shutdown(shutdownCtx); err != nil {
		log.Error("error shutting down server", slog.String("err", err.Error()))
	}

	<-done
}
