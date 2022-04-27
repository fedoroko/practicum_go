package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"

	"github.com/fedoroko/practicum_go/internal/config"
	"github.com/fedoroko/practicum_go/internal/metrics"
	"github.com/fedoroko/practicum_go/internal/storage"
)

type repoHandler struct {
	r      storage.Repository
	logger *config.Logger
}

func NewRepoHandler(r storage.Repository, logger *config.Logger) *repoHandler {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	subLogger := logger.With().Str("Component", "Handler").Logger()
	return &repoHandler{
		r:      r,
		logger: config.NewLogger(&subLogger),
	}
}

func (h *repoHandler) IndexFunc(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug().Msg("IndexFunc")

	w.Header().Set("Content-Type", "text/html")

	data, err := h.r.List()
	if err != nil {
		h.logger.Error().Stack().Err(err).Msg("")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html := "<div><ul>"
	for i := range data {
		html += "<li>" + data[i].Name() + " - " + data[i].ToString() + "</li>"
	}
	html += "</ul></div>"

	w.Write([]byte(html))
}

func (h *repoHandler) UpdateFunc(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug().Msg("UpdateFunc")

	t := chi.URLParam(r, "type")
	n := chi.URLParam(r, "name")
	v := chi.URLParam(r, "value")

	m, err := metrics.RawWithValue(t, n, v)
	if err != nil {
		h.logger.Error().Stack().Err(err).Msg("")

		switch {
		case errors.As(err, &metrics.InvalidType):
			http.Error(w, err.Error(), http.StatusNotImplemented)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	if err = h.r.Set(m); err != nil {
		h.logger.Error().Stack().Err(err).Msg("")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(""))
}

func (h *repoHandler) UpdateJSONFunc(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug().Msg("UpdateJSONFunc")
	m, err := metrics.FromJSON(r.Body)
	if err != nil {
		h.logger.Error().Stack().Err(err).Msg("")

		switch {
		case errors.As(err, &metrics.InvalidType):
			http.Error(w, err.Error(), http.StatusNotImplemented)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	if err = h.r.Set(m); err != nil {
		h.logger.Error().Stack().Err(err).Msg("")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(""))
}

func (h *repoHandler) GetFunc(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug().Msg("GetFunc")

	t := chi.URLParam(r, "type")
	n := chi.URLParam(r, "name")

	m, err := metrics.Raw(t, n)
	if err != nil {
		h.logger.Error().Stack().Err(err).Msg("")

		switch {
		case errors.As(err, &metrics.InvalidType):
			http.Error(w, err.Error(), http.StatusNotImplemented)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		return
	}

	ret, err := h.r.Get(m)
	if err != nil {
		h.logger.Error().Stack().Err(err).Msg("")

		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(ret.ToString()))
}

func (h *repoHandler) GetJSONFunc(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug().Msg("GetJSONFunc")

	m, err := metrics.FromJSON(r.Body)
	if err != nil {
		h.logger.Error().Stack().Err(err).Msg("")

		switch {
		case errors.As(err, &metrics.InvalidType):
			http.Error(w, err.Error(), http.StatusNotImplemented)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	ret, err := h.r.Get(m)
	if err != nil {
		h.logger.Error().Stack().Err(err).Msg("")

		switch {
		case errors.As(err, &metrics.InvalidHash):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusNotFound)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(ret.ToJSON())
}

func (h *repoHandler) PingFunc(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug().Msg("PingFunc")

	if err := h.r.Ping(); err != nil {
		h.logger.Error().Stack().Err(err).Msg("")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(""))
}

func (h *repoHandler) UpdatesFunc(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug().Msg("UpdatesFunc")

	ms, err := metrics.ArrFromJSON(r.Body)
	if err != nil {
		h.logger.Error().Stack().Err(err).Msg("")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = h.r.SetBatch(ms); err != nil {
		h.logger.Error().Stack().Err(err).Msg("")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(""))
}
