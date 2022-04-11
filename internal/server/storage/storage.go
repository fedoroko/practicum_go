package storage

import (
	"errors"
	"fmt"
	"github.com/caarlos0/env/v6"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// Repository не хочется плодить методы для разных типов контента, решил экспериментировать с оцпиями
type Repository interface {
	Get(i input, o output) (string, error)
	Set(i input) error
	List() string
	restore() error
	listenAndWrite()
}

type gauge float64

type counter int64

type repo struct {
	G             map[string]gauge `json:"gauge"`
	gMtx          sync.RWMutex
	C             map[string]counter `json:"counter"`
	cMtx          sync.RWMutex
	storeInterval time.Duration
	storeFile     string
	producer      *producer
}

func repoInterface(cfg *config) *repo {
	flag := 0
	if cfg.StoreInterval == 0 {
		flag = os.O_SYNC
	}

	p, err := newProducer(cfg.StoreFile, flag)
	if err != nil {
		log.Fatal(err)
	}
	return &repo{
		G:             make(map[string]gauge),
		gMtx:          sync.RWMutex{},
		C:             make(map[string]counter),
		cMtx:          sync.RWMutex{},
		storeInterval: cfg.StoreInterval,
		storeFile:     cfg.StoreFile,
		producer:      p,
	}
}

func (r *repo) Get(i input, o output) (string, error) {
	m, err := i()
	if err != nil {
		return "", err
	}
	n := strings.ToLower(m.ID)
	switch m.MType {
	case "gauge":
		r.gMtx.Lock()
		defer r.gMtx.Unlock()
		if v, ok := r.G[n]; ok {
			z := float64(v)
			m.Value = &z
			return o(m), nil
		}

	case "counter":
		r.cMtx.Lock()
		defer r.cMtx.Unlock()
		if v, ok := r.C[n]; ok {
			z := int64(v)
			m.Delta = &z
			return o(m), nil
		}

	default:
		return "", throwInvalidTypeError(m.MType)
	}

	return "", errors.New("not found")
}

func (r *repo) Set(i input) error {
	if r.storeInterval == 0 {
		defer r.producer.write(r)
	}
	m, err := i()
	if err != nil {
		return err
	}

	n := strings.ToLower(m.ID)
	switch m.MType {
	case "gauge":
		if m.Value == nil {
			return errors.New("bad")
		}
		r.gMtx.RLock()
		defer r.gMtx.RUnlock()

		r.G[n] = gauge(*m.Value)
		return nil

	case "counter":
		r.cMtx.RLock()
		defer r.cMtx.RUnlock()

		if v, ok := r.C[n]; ok {
			r.C[n] = v + counter(*m.Delta)
		} else {
			r.C[n] = counter(*m.Delta)
		}

		return nil
	}

	return throwInvalidTypeError(m.MType)
}

func (r *repo) List() string {
	ret := "Known values:\n"
	r.gMtx.Lock()
	defer r.gMtx.Unlock()
	for n, v := range r.G {
		ret += fmt.Sprintf("%s - %v\n", n, v)
	}

	r.cMtx.Lock()
	defer r.cMtx.Unlock()
	for n, v := range r.C {
		ret += fmt.Sprintf("%s - %v\n", n, v)
	}

	return ret
}

func (r *repo) restore() error {
	if r.storeFile == "" {
		return errors.New("empty file path")
	}

	c, err := newConsumer(r.storeFile)
	if err != nil {
		return err
	}
	defer c.close()
	err = c.read(r)
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
		log.Fatal(r.producer.write(r))
	}
}

type config struct {
	Restore       bool          `env:"RESTORE" envDefault:"true"`
	StoreInterval time.Duration `env:"STORE_INTERVAL" envDefault:"300s"`
	StoreFile     string        `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
}

func Init() Repository {
	cfg := &config{
		Restore:       true,
		StoreInterval: 300 * time.Second,
		StoreFile:     "/tmp/devops-metrics-db.json",
	}
	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(err)
	}
	db := repoInterface(cfg)

	if cfg.Restore {
		err = db.restore()
		if err != nil {
			log.Fatal(err)
		}
	}

	go db.listenAndWrite()
	return db
}
