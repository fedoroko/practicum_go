package server

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/fedoroko/practicum_go/internal/config"
	"github.com/fedoroko/practicum_go/internal/handlers"
	"github.com/fedoroko/practicum_go/internal/storage"
)

func Run(cfg *config.ServerConfig, logger *config.Logger) {
	db := storage.New(cfg, logger)
	defer db.Close()

	r := router(&db, logger)

	server := &http.Server{
		Addr:    cfg.Address,
		Handler: r,
	}

	defer server.Close()
	go func() {
		if err := server.ListenAndServe(); err != nil {
			logger.Error().Err(err).Send()
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)
	<-sig
}

func router(db *storage.Repository, logger *config.Logger) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Compress(5))

	h := handlers.NewRepoHandler(*db, logger)

	r.Get("/", h.IndexFunc)
	r.Route("/value", func(r chi.Router) {
		r.Post("/", h.GetJSONFunc)
		r.Get("/{type}/{name}", h.GetFunc)
	})
	r.Route("/update", func(r chi.Router) {
		r.Post("/", h.UpdateJSONFunc)
		r.Post("/{type}/{name}/{value}", h.UpdateFunc)
	})
	r.Route("/ping", func(r chi.Router) {
		r.Get("/", h.PingFunc)
	})
	r.Route("/updates", func(r chi.Router) {
		r.Post("/", h.UpdatesFunc)
	})

	return r
}
