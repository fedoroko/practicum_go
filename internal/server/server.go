package server

import (
	"context"
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
			log.Fatal(err)
		}
	}
}

func Run(opts ...option) {
	cfg := &config{}

	parseFlags(cfg)

	for _, o := range opts {
		o(cfg)
	}
	r := router(cfg)

	server := &http.Server{
		Addr:    cfg.Address,
		Handler: r,
	}

	go func() {
		log.Fatal(server.ListenAndServe())
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)
	<-sig
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer os.Exit(0)

	log.Fatal(server.Shutdown(ctx))
}

func router(cfg *config) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)

	db := storage.Init(
		&storage.Config{
			Restore:       cfg.Restore,
			StoreInterval: cfg.StoreInterval,
			StoreFile:     cfg.StoreFile,
		},
	)
	defer db.Close()
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
