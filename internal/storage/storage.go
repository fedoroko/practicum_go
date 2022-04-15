package storage

import (
	"errors"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/fedoroko/practicum_go/internal/config"
	"github.com/fedoroko/practicum_go/internal/errrs"
	"github.com/fedoroko/practicum_go/internal/metrics"
)

type Repository interface {
	Get(m metrics.Metric) (metrics.Metric, error)
	Set(m metrics.Metric) error
	List() []metrics.Metric

	restore() error
	listenAndWrite()

	Close() error
}

type gauge float64

type counter int64

type repo struct {
	G             map[string]gauge `json:"gauge"`
	gMtx          sync.RWMutex
	C             map[string]counter `json:"counter"`
	cMtx          sync.RWMutex
	storeInterval time.Duration
	producer      *producer
	consumer      *consumer
}

func (r *repo) Get(m metrics.Metric) (metrics.Metric, error) {
	switch m.Type() {
	case metrics.GaugeType:
		r.gMtx.Lock()
		defer r.gMtx.Unlock()
		if v, ok := r.G[m.Name()]; ok {
			m.SetFloat64(float64(v))
			return m, nil
		}

	case metrics.CounterType:
		r.cMtx.Lock()
		defer r.cMtx.Unlock()
		if v, ok := r.C[m.Name()]; ok {
			m.SetInt64(int64(v))
			return m, nil
		}

	default:
		return m, errrs.ThrowInvalidTypeError(m.Type())
	}

	return m, errors.New("not found")
}

func (r *repo) Set(m metrics.Metric) error {
	if r.storeInterval == 0 {
		defer r.producer.write(r)
	}

	switch m.Type() {
	case metrics.GaugeType:
		v, err := m.Float64Value()
		if err != nil {
			return err
		}

		r.gMtx.RLock()
		defer r.gMtx.RUnlock()

		r.G[m.Name()] = gauge(v)
		return nil

	case metrics.CounterType:
		v, err := m.Int64Value()
		if err != nil {
			return err
		}

		r.cMtx.RLock()
		defer r.cMtx.RUnlock()

		if cur, ok := r.C[m.Name()]; ok {
			r.C[m.Name()] = cur + counter(v)
		} else {
			r.C[m.Name()] = counter(v)
		}

		return nil
	}

	return errrs.ThrowInvalidTypeError(m.Type())
}

func (r *repo) List() []metrics.Metric {
	var ret []metrics.Metric

	r.gMtx.Lock()
	defer r.gMtx.Unlock()
	for n, v := range r.G {
		ret = append(ret, metrics.New(
			n,
			metrics.GaugeType,
			float64(v),
			0),
		)

	}

	r.cMtx.Lock()
	defer r.cMtx.Unlock()
	for n, v := range r.C {
		ret = append(ret, metrics.New(
			n,
			metrics.CounterType,
			0,
			int64(v)),
		)
	}

	return ret
}

func (r *repo) restore() error {
	defer r.consumer.close()
	err := r.consumer.read(r)
	if errors.Is(err, io.EOF) {
		err = nil
	}
	return err
}

func (r *repo) listenAndWrite() {
	if r.storeInterval == 0 {
		return
	}

	t := time.NewTicker(r.storeInterval)
	defer t.Stop()
	for range t.C {
		r.producer.write(r)
	}
}

func (r *repo) Close() error {
	return r.producer.close()
}

func repoInterface(cfg *config.ServerConfig) *repo {
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
	return &repo{
		G:             make(map[string]gauge),
		gMtx:          sync.RWMutex{},
		C:             make(map[string]counter),
		cMtx:          sync.RWMutex{},
		storeInterval: cfg.StoreInterval,
		producer:      p,
		consumer:      c,
	}
}

func New(cfg *config.ServerConfig) Repository {
	db := repoInterface(cfg)

	if cfg.Restore {
		err := db.restore()
		if err != nil {
			log.Println(err)
		}
	}

	go db.listenAndWrite()
	return db
}
