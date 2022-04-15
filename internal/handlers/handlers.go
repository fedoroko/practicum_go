package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/fedoroko/practicum_go/internal/errrs"
	"github.com/fedoroko/practicum_go/internal/metrics"
	"github.com/fedoroko/practicum_go/internal/storage"
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
	w.Header().Set("Content-Type", "text/html")

	data := h.r.List()
	html := "<div><ul>"
	for i := range data {
		html += "<li>" + data[i].Name() + " - " + data[i].ToString() + "</li>"
	}
	html += "</ul></div>"

	w.Write([]byte(html))
}

func (h *repoHandler) UpdateFunc(w http.ResponseWriter, r *http.Request) {
	t := chi.URLParam(r, "type")
	n := chi.URLParam(r, "name")
	v := chi.URLParam(r, "value")

	m, err := metrics.RawWithValue(t, n, v)
	if err != nil {
		switch {
		case errors.As(err, &errrs.InvalidType):
			http.Error(w, err.Error(), http.StatusNotImplemented)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}

	if err = h.r.Set(m); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(""))
}

func (h *repoHandler) UpdateJSONFunc(w http.ResponseWriter, r *http.Request) {
	m, err := metrics.FromJSON(r.Body)
	if err != nil {
		switch {
		case errors.As(err, &errrs.InvalidType):
			http.Error(w, err.Error(), http.StatusNotImplemented)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

	}

	if err = h.r.Set(m); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(""))
}

func (h *repoHandler) GetFunc(w http.ResponseWriter, r *http.Request) {
	t := chi.URLParam(r, "type")
	n := chi.URLParam(r, "name")

	m, err := metrics.Raw(t, n)
	if err != nil {
		switch {
		case errors.As(err, &errrs.InvalidType):
			e := err.Error()
			http.Error(w, e, http.StatusNotImplemented)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}

	ret, err := h.r.Get(m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(ret.ToString()))
}

func (h *repoHandler) GetJSONFunc(w http.ResponseWriter, r *http.Request) {
	m, err := metrics.FromJSON(r.Body)
	if err != nil {
		switch {
		case errors.As(err, &errrs.InvalidType):
			http.Error(w, err.Error(), http.StatusNotImplemented)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

	}
	ret, err := h.r.Get(m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(ret.ToJSON())
}
