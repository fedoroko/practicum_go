package server

import (
	"log"
	"net/http"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/fedoroko/practicum_go/internal/server/handlers"
	"github.com/fedoroko/practicum_go/internal/server/storage"
)

type config struct {
	Address       string        `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
	Restore       bool          `env:"RESTORE" envDefault:"true"`
	StoreInterval time.Duration `env:"STORE_INTERVAL" envDefault:"300s"`
	StoreFile     string        `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
}

type option func(*config)

func WithEnv() option {
	return func(cfg *config) {
		err := env.Parse(cfg)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func Run(opts ...option) {
	cfg := &config{
		Address:       "127.0.0.1:8080",
		Restore:       true,
		StoreInterval: 300 * time.Second,
		StoreFile:     "/tmp/devops-metrics-db.json",
	}
	for _, o := range opts {
		o(cfg)
	}

	r := router()

	server := &http.Server{
		Addr:    cfg.Address,
		Handler: r,
	}

	log.Fatal(server.ListenAndServe())
}

func router() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)

	db := storage.Init()
	h := handlers.NewRepoHandler(db)

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
