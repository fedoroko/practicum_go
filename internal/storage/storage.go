package storage

import (
	"errors"
	"io"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/fedoroko/practicum_go/internal/config"
	"github.com/fedoroko/practicum_go/internal/metrics"
)

type Repository interface {
	Get(metrics.Metric) (metrics.Metric, error)
	Set(metrics.Metric) error
	SetBatch([]metrics.Metric) error
	List() ([]metrics.Metric, error)

	Ping() error
	Close() error
}

type gauge float64

type counter int64

type repo struct {
	G        map[string]gauge `json:"gauge"`
	gMtx     sync.RWMutex
	C        map[string]counter `json:"counter"`
	cMtx     sync.RWMutex
	cfg      *config.ServerConfig
	producer *producer
	consumer *consumer
	logger   *config.Logger
}

func (r *repo) Get(m metrics.Metric) (metrics.Metric, error) {
	switch m.Type() {
	case metrics.GaugeType:
		r.gMtx.RLock()
		defer r.gMtx.RUnlock()
		v, ok := r.G[m.Name()]
		if !ok {
			return m, errors.New("not found")
		}
		m.SetFloat64(float64(v))

	case metrics.CounterType:
		r.cMtx.RLock()
		defer r.cMtx.RUnlock()
		v, ok := r.C[m.Name()]
		if !ok {
			return m, errors.New("not found")
		}
		m.SetInt64(int64(v))

	default:
		return m, metrics.ThrowInvalidTypeError(m.Type())
	}

	if err := m.SetHash(r.cfg.Key); err != nil {
		return m, err
	}

	return m, nil
}

func (r *repo) Set(m metrics.Metric) error {
	if ok, _ := m.CheckHash(r.cfg.Key); !ok {
		return metrics.ThrowInvalidHashError()
	}

	if r.cfg.StoreInterval == 0 {
		defer r.producer.write(r)
	}

	switch m.Type() {
	case metrics.GaugeType:
		r.gMtx.Lock()
		defer r.gMtx.Unlock()

		r.G[m.Name()] = gauge(m.Float64Value())

	case metrics.CounterType:
		r.cMtx.Lock()
		defer r.cMtx.Unlock()

		if cur, ok := r.C[m.Name()]; ok {
			r.C[m.Name()] = cur + counter(m.Int64Value())
		} else {
			r.C[m.Name()] = counter(m.Int64Value())
		}

	default:
		return metrics.ThrowInvalidTypeError(m.Type())
	}

	return nil
}

func (r *repo) SetBatch(ms []metrics.Metric) error {
	for _, m := range ms {
		if err := m.CheckType(); err != nil {
			return err
		}

		if ok, _ := m.CheckHash(r.cfg.Key); !ok {
			return metrics.ThrowInvalidHashError()
		}

		if err := r.Set(m); err != nil {
			return err
		}
	}

	return nil
}

func (r *repo) List() ([]metrics.Metric, error) {
	var ret []metrics.Metric

	r.gMtx.RLock()
	defer r.gMtx.RUnlock()
	for n, v := range r.G {
		ret = append(ret, metrics.New(
			n,
			metrics.GaugeType,
			float64(v),
			0),
		)
	}

	r.cMtx.RLock()
	defer r.cMtx.RUnlock()
	for n, v := range r.C {
		ret = append(ret, metrics.New(
			n,
			metrics.CounterType,
			0,
			int64(v)),
		)
	}

	return ret, nil
}

func (r *repo) restore() error {
	r.logger.Info().Msg("Restoring DB")
	defer r.consumer.close()
	err := r.consumer.read(r)
	if errors.Is(err, io.EOF) {
		err = nil
	}
	return err
}

func (r *repo) listenAndWrite() {
	if r.cfg.StoreInterval == 0 {
		return
	}

	t := time.NewTicker(r.cfg.StoreInterval)
	defer t.Stop()
	for range t.C {
		if err := r.producer.write(r); err != nil {
			r.logger.Error().Stack().Err(err).Msg("")
		}
	}
}

func (r *repo) Ping() error {
	return nil
}

func (r *repo) Close() error {
	r.logger.Info().Msg("DB: closed")
	return r.producer.close()
}

func repoInterface(cfg *config.ServerConfig, logger *config.Logger) *repo {
	flag := 0
	if cfg.StoreInterval == 0 {
		flag = os.O_SYNC
	}

	p, err := newProducer(cfg.StoreFile, flag)
	if err != nil {
		panic(err)
	}

	c, err := newConsumer(cfg.StoreFile)
	if err != nil {
		panic(err)
	}

	subLogger := logger.With().Str("Component", "DUMMY-DB").Logger()
	return &repo{
		G:        make(map[string]gauge),
		gMtx:     sync.RWMutex{},
		C:        make(map[string]counter),
		cMtx:     sync.RWMutex{},
		cfg:      cfg,
		producer: p,
		consumer: c,
		logger:   config.NewLogger(&subLogger),
	}
}

func New(cfg *config.ServerConfig, logger *config.Logger) Repository {
	if cfg.Database != "" {
		logger.Info().Msg("DB: postgres")
		return postgresInterface(cfg, logger)
	}

	log.Info().Msg("DB: dummy")
	db := repoInterface(cfg, logger)

	if cfg.Restore {
		err := db.restore()
		if err != nil {
			log.Error().Err(err).Send()
		}
	}

	go db.listenAndWrite()
	return db
}
