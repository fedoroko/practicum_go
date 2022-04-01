package handlers

import (
	"errors"

	"net/http"

	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/fedoroko/practicum_go/internal/server/storage"
)

type DBHandler struct {
	DB storage.Repository
}

func NewDBHandler(db storage.Repository) *DBHandler {
	return &DBHandler{
		DB: db,
	}
}

func (h *DBHandler) IndexFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	data := h.DB.Display()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data))
}

func (h *DBHandler) UpdateFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	t := strings.ToLower(chi.URLParam(r, "type"))
	n := strings.ToLower(chi.URLParam(r, "name"))
	v := strings.ToLower(chi.URLParam(r, "value"))

	var typeErr *storage.InvalidTypeError
	err := h.DB.Set(t, n, v)
	if err != nil {
		switch {
		case errors.As(err, &typeErr):
			http.Error(w, err.Error(), http.StatusNotImplemented)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (h *DBHandler) GetFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	t := strings.ToLower(chi.URLParam(r, "type"))
	n := strings.ToLower(chi.URLParam(r, "name"))

	var typeErr *storage.InvalidTypeError
	ret, err := h.DB.Get(t, n)
	if err != nil {
		switch {
		case errors.As(err, &typeErr):
			http.Error(w, err.Error(), http.StatusNotImplemented)
		default:
			http.Error(w, err.Error(), http.StatusNotFound)
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(ret))
}
