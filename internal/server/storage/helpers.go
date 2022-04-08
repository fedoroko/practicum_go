package storage

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type InvalidTypeError struct {
	Type string
}

func (e *InvalidTypeError) Error() string {
	return fmt.Sprintf("Invalid type: %v", e.Type)
}

func throwInvalidTypeError(t string) error {
	return &InvalidTypeError{Type: t}
}

type Metrics struct {
	Id    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

type input func() (*Metrics, error)

func Raw(t string, n string) input {
	return func() (*Metrics, error) {
		return &Metrics{
			Id:    n,
			MType: t,
		}, nil
	}
}

func RawWithValue(t string, n string, v string) input {
	return func() (*Metrics, error) {
		m := Metrics{
			Id:    n,
			MType: t,
		}
		switch t {
		case "counter":
			v64, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return &m, err
			}
			m.Delta = &v64
		default:
			v64, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return &m, err
			}
			m.Value = &v64
		}

		return &m, nil
	}
}

func FromJSON(b []byte) input {
	return func() (*Metrics, error) {
		m := Metrics{}
		err := json.Unmarshal(b, &m)
		if err != nil {
			return &m, err
		}

		return &m, nil
	}
}

type output func(m *Metrics) string

func Plain() output {
	return func(m *Metrics) string {
		switch m.MType {
		case "counter":
			return fmt.Sprintf("%v", *m.Delta)
		default:
			return fmt.Sprintf("%v", *m.Value)
		}
	}
}

func ToJSON() output {
	return func(m *Metrics) string {
		b, err := json.Marshal(&m)
		if err != nil {
			panic(err)
		}
		//костыль, не знаю как избавиться
		return string(b)
	}
}
