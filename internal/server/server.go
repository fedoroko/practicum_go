package server

import (
	"github.com/fedoroko/practicum_go/internal/server/storage"
	"log"

	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/fedoroko/practicum_go/internal/server/handlers"
)

func Run() {
	r := router()
	address := "127.0.0.1:8080"
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
	h := handlers.NewDBHandler(db)

	r.Get("/", h.IndexFunc)
	r.Get("/value/{type}/{name}", h.GetFunc)
	r.Post("/update/{type}/{name}/{value}", h.UpdateFunc)

	return r
}
