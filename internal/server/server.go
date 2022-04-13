package server

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/fedoroko/practicum_go/internal/config"
	"github.com/fedoroko/practicum_go/internal/server/handlers"
	"github.com/fedoroko/practicum_go/internal/server/storage"
)

type option func(serverConfig *config.ServerConfig)

func Run(cfg *config.ServerConfig, opts ...option) {
	log.Println("Run func")

	for _, o := range opts {
		o(cfg)
	}
	log.Println("cfg:", cfg)
	db := storage.New(cfg)
	log.Println("db created")
	defer db.Close()

	r := router(&db)

	server := &http.Server{
		Addr:    cfg.Address,
		Handler: r,
	}

	log.Println("run : starting server")
	defer log.Println("run : starting ended")
	defer server.Close()
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Println(err)
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

func router(db *storage.Repository) chi.Router {
	log.Println("Router func")
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)

	h := handlers.NewRepoHandler(*db)

	r.Get("/", h.IndexFunc)
	r.Route("/value", func(r chi.Router) {
		r.Post("/", h.GetJSONFunc)
		r.Get("/{type}/{name}", h.GetFunc)
	})
	r.Route("/update", func(r chi.Router) {
		r.Post("/", h.UpdateJSONFunc)
		r.Post("/{type}/{name}/{value}", h.UpdateFunc)
	})

	return r
}
