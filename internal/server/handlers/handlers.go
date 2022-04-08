package handlers

import (
	"errors"
	"github.com/fedoroko/practicum_go/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
)

type repoHandler struct {
	r storage.Repository
}

func NewRepoHandler(r storage.Repository) *repoHandler {
	return &repoHandler{
		r: r,
	}
}

func (h *repoHandler) IndexFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	data := h.r.List()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data))
}

func (h *repoHandler) UpdateFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	t := chi.URLParam(r, "type")
	n := chi.URLParam(r, "name")
	v := chi.URLParam(r, "value")

	err := h.r.Set(
		storage.RawWithValue(t, n, v),
	)

	var typeErr *storage.InvalidTypeError

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

func (h *repoHandler) UpdateJSONFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	err = h.r.Set(
		storage.FromJSON(b),
	)

	var typeErr *storage.InvalidTypeError

	if err != nil {
		switch {
		case errors.As(err, &typeErr):
			http.Error(w, err.Error(), http.StatusNotImplemented)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(""))
}

func (h *repoHandler) GetFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	t := chi.URLParam(r, "type")
	n := chi.URLParam(r, "name")

	var typeErr *storage.InvalidTypeError

	ret, err := h.r.Get(
		storage.Raw(t, n),
		storage.Plain(),
	)
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

func (h *repoHandler) GetJSONFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	ret, err := h.r.Get(
		storage.FromJSON(b),
		storage.ToJSON(),
	)

	var typeErr *storage.InvalidTypeError
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
