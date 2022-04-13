package server

import (
	"context"
	"flag"
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
	Address       string `env:"ADDRESS"`
	Restore       bool
	StoreInterval time.Duration
	StoreFile     string
}

func parseFlags(cfg *config) {
	flag.StringVar(&cfg.Address, "a", "127.0.0.1:8080", "Host address")
	flag.BoolVar(&cfg.Restore, "r", true, "Restore previous db")
	i := flag.String("i", "300s", "Store interval")
	flag.StringVar(&cfg.StoreFile, "f", "/tmp/devops-metrics-db.json", "Store file path")
	flag.Parse()
	d, err := time.ParseDuration(*i)
	if err == nil {
		cfg.StoreInterval = d
	}
}

type option func(*config)

func WithEnv() option {
	return func(cfg *config) {
		err := env.Parse(cfg)
		if err != nil {
			log.Println(err)
		}
	}
}

func Run(opts ...option) {
	cfg := &config{}

	parseFlags(cfg)

	for _, o := range opts {
		o(cfg)
	}
	log.Println("cfg:", cfg)
	db := storage.Init(
		&storage.Config{
			Restore:       cfg.Restore,
			StoreInterval: cfg.StoreInterval,
			StoreFile:     cfg.StoreFile,
		},
	)

	defer db.Close()
	r := router(&db)
	log.Println("db created")
	server := &http.Server{
		Addr:    cfg.Address,
		Handler: r,
	}

	log.Println("run : starting server")
	defer log.Println("run : starting ended")
	defer server.Shutdown(context.Background())
	if err := server.ListenAndServe(); err != nil {
		log.Println(err)
	}

}

func router(db *storage.Repository) chi.Router {
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
