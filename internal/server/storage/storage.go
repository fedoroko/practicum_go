package storage

import (
	"errors"
	"fmt"

	"strconv"
)

type Repository interface {
	Get(t string, n string) (string, error)
	Set(t string, n string, v string) error
	Display() string
}

type InvalidTypeError struct {
	Type string
}

func (e *InvalidTypeError) Error() string {
	return fmt.Sprintf("Invalid type: %v", e.Type)
}

func throwInvalidTypeError(t string) error {
	return &InvalidTypeError{Type: t}
}

type gauge float64

type counter int64

type gaugeStorage struct {
	Fields map[string]gauge
}

type counterStorage struct {
	Fields map[string]counter
}

type DummyDB struct {
	G *gaugeStorage
	C *counterStorage
}

func DummyDBInterface(g *gaugeStorage, c *counterStorage) *DummyDB {
	return &DummyDB{
		G: g,
		C: c,
	}
}

func (db *DummyDB) Get(t string, n string) (string, error) {
	switch t {
	case "gauge":
		if v, ok := db.G.Fields[n]; ok {
			return fmt.Sprintf("%v", v), nil
		}
	case "counter":
		if v, ok := db.C.Fields[n]; ok {
			return fmt.Sprintf("%v", v), nil
		}
	default:
		return "", throwInvalidTypeError(t)
	}
	return "", errors.New("not found")
}

func (db *DummyDB) Set(t string, n string, v string) error {
	switch t {
	case "gauge":
		v64, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return err
		}
		db.G.Fields[n] = gauge(v64)
		return nil

	case "counter":
		v64, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}

		db.C.Fields[n] = counter(v64)
		return nil
	}

	return throwInvalidTypeError(t)
}

func (db *DummyDB) Display() string {
	ret := "Known values:\n"
	for n := range db.G.Fields {
		v, _ := db.Get("gauge", n)
		ret += fmt.Sprintf("%s - %s\n", n, v)
	}
	for n := range db.C.Fields {
		v, _ := db.Get("counter", n)
		ret += fmt.Sprintf("%s - %s\n", n, v)
	}

	return ret
}

func Init() *DummyDB {
	g := &gaugeStorage{
		Fields: make(map[string]gauge),
	}
	c := &counterStorage{
		Fields: make(map[string]counter),
	}
	db := DummyDBInterface(g, c)
	return db
}
