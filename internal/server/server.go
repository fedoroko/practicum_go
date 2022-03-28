package server

import (
	"net/http"
	"github.com/fedoroko/practicum_go/internal/handlers"
)

var address string = "127.0.0.1:8080"

func Run() {

	http.HandleFunc("/update/", handlers.UpdateFunc)
	server := &http.Server{
		Addr: address,
		// Handler: uHandler,
	}

	server.ListenAndServe()
}