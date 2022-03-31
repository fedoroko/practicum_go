package server

import (
	"github.com/fedoroko/practicum_go/internal/server/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
)

var address string = "127.0.0.1:8080"

func Run() {
	r := router()

	server := &http.Server{
		Addr:    address,
		Handler: r,
	}

	log.Fatal(server.ListenAndServe())
}

func router() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", handlers.IndexFunc)
	r.Get("/value/{type}/{name}", handlers.GetFunc)
	r.Post("/update/{type}/{name}/{value}", handlers.UpdateFunc)

	return r
}
