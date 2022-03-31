package storage

import (
	"errors"
	"fmt"
	"strconv"
)

type repositories interface {
	get(n string) (string, error)
	update(n string, v string) error
}

type gauge float64

type counter int64

type gaugeStorage struct {
	Fields map[string]gauge
}

type counterStorage struct {
	Fields map[string]counter
}

func (g gaugeStorage) get(n string) (string, error) {
	if v, ok := g.Fields[n]; ok {
		return fmt.Sprintf("%v", v), nil
	}

	return "", errors.New("not found")
}

func (c counterStorage) get(n string) (string, error) {
	if v, ok := c.Fields[n]; ok {
		return fmt.Sprintf("%v", v), nil
	}

	return "", errors.New("not found")
}

func (g *gaugeStorage) update(n string, v string) error {
	if g.Fields == nil {
		g.Fields = make(map[string]gauge)
	}

	v64, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return err
	}

	g.Fields[n] = gauge(v64)

	return nil
}

func (c *counterStorage) update(n string, v string) error {
	if c.Fields == nil {
		c.Fields = make(map[string]counter)
	}
	v64, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return err
	}

	c.Fields[n] = counter(v64)
	return nil
}

func collect(r repositories, n string) (string, error) {
	return r.get(n)
}

func update(r repositories, n string, v string) error {
	return r.update(n, v)
}
