package server

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/fedoroko/practicum_go/internal/server/handlers"
	"github.com/fedoroko/practicum_go/internal/server/storage"
)

type config struct {
	address string
}

type option func(*config)

func WithEnv() option {
	a := "127.0.0.1:8080"
	address := os.Getenv("ADDRESS")
	if address != "" {
		a = address
	}
	return func(c *config) {
		c.address = a
	}
}

func Run(opts ...option) {
	cfg := &config{
		address: "127.0.0.1:8080",
	}
	for _, o := range opts {
		o(cfg)
	}
	r := router()
	address := cfg.address
	server := &http.Server{
		Addr:    address,
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
