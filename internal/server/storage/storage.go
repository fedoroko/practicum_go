package storage

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

// Repository не хочется плодить методы для разных типов контента, решил экспериментировать с оцпиями
type Repository interface {
	Get(i input, o output) (string, error)
	Set(i input) error
	List() string
}

type gauge float64

type counter int64

type repo struct {
	g    map[string]gauge
	gMtx sync.RWMutex
	c    map[string]counter
	cMtx sync.RWMutex
}

func repoInterface() *repo {
	return &repo{
		g:    make(map[string]gauge),
		gMtx: sync.RWMutex{},
		c:    make(map[string]counter),
		cMtx: sync.RWMutex{},
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
		if v, ok := r.g[n]; ok {
			z := float64(v)
			m.Value = &z
			return o(m), nil
		}

	case "counter":
		r.cMtx.Lock()
		defer r.cMtx.Unlock()
		if v, ok := r.c[n]; ok {
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

		r.g[n] = gauge(*m.Value)
		return nil

	case "counter":
		r.cMtx.RLock()
		defer r.cMtx.RUnlock()

		if v, ok := r.c[n]; ok {
			r.c[n] = v + counter(*m.Delta)
		} else {
			r.c[n] = counter(*m.Delta)
		}

		return nil
	}

	return throwInvalidTypeError(m.MType)
}

func (r *repo) List() string {
	ret := "Known values:\n"
	r.gMtx.Lock()
	defer r.gMtx.Unlock()
	for n, v := range r.g {
		ret += fmt.Sprintf("%s - %v\n", n, v)
	}

	r.cMtx.Lock()
	defer r.cMtx.Unlock()
	for n, v := range r.c {
		ret += fmt.Sprintf("%s - %v\n", n, v)
	}

	return ret
}

func Init() Repository {
	db := repoInterface()
	return db
}
