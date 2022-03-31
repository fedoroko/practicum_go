package handlers

import (
	"github.com/fedoroko/practicum_go/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strings"
)

func IndexFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	data, err := storage.Values()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data))
}

func UpdateFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	t := strings.ToLower(chi.URLParam(r, "type"))
	n := strings.ToLower(chi.URLParam(r, "name"))
	v := strings.ToLower(chi.URLParam(r, "value"))

	err := storage.Store(t, n, v)
	if err != nil {
		switch err.(type) {
		case *storage.InvalidTypeError:
			http.Error(w, err.Error(), http.StatusNotImplemented)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func GetFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	t := strings.ToLower(chi.URLParam(r, "type"))
	n := strings.ToLower(chi.URLParam(r, "name"))

	ret, err := storage.Get(t, n)
	if err != nil {
		switch err.(type) {
		case *storage.InvalidTypeError:
			http.Error(w, err.Error(), http.StatusNotImplemented)
		default:
			http.Error(w, err.Error(), http.StatusNotFound)
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(ret))
}
