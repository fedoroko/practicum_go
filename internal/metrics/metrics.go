package metrics

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/fedoroko/practicum_go/internal/errrs"
)

const (
	GaugeType   = "gauge"
	CounterType = "counter"
)

type Metric interface {
	Name() string
	Type() string
	Float64Value() (float64, error)
	Int64Value() (int64, error)

	SetFloat64(float64)
	SetInt64(int64)

	ToString() string
	ToJSON() ([]byte, error)
}

type metric struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func (m *metric) Name() string {
	return strings.ToLower(m.ID)
}

func (m *metric) Type() string {
	return m.MType
}

func (m *metric) Float64Value() (float64, error) {
	if m.Value == nil {
		return 0, errors.New("empty value")
	}
	return *m.Value, nil
}

func (m *metric) Int64Value() (int64, error) {
	if m.Delta == nil {
		return 0, errors.New("empty value")
	}
	return *m.Delta, nil
}

func (m *metric) SetFloat64(f float64) {
	m.Value = &f
}

func (m *metric) SetInt64(i int64) {
	m.Delta = &i
}

func (m *metric) ToString() string {
	switch m.Type() {
	case GaugeType:
		if m.Value == nil {
			return ""
		}
		return fmt.Sprintf("%v", *m.Value)
	case CounterType:
		if m.Delta == nil {
			return ""
		}
		return fmt.Sprintf("%v", *m.Delta)
	}
	return ""
}

func (m *metric) ToJSON() ([]byte, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return []byte(""), err
	}

	return b, nil
}

func RawWithValue(t string, n string, v string) (Metric, error) {
	m := &metric{
		ID:    n,
		MType: t,
	}

	//if v == "" {
	//	return m, errors.New("empty value")
	//}

	switch t {
	case GaugeType:
		f64, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return m, err
		}
		m.Value = &f64
	case CounterType:
		i64, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return m, err
		}
		m.Delta = &i64
	default:
		return m, errrs.ThrowInvalidTypeError(t)
	}
	return m, nil
}

func Raw(t string, n string) (Metric, error) {
	m := &metric{
		ID:    n,
		MType: t,
	}
	if t != GaugeType && t != CounterType {
		return m, errrs.ThrowInvalidTypeError(t)
	}

	return m, nil
}

func FromJSON(j io.Reader) (Metric, error) {
	m := metric{}
	decoder := json.NewDecoder(j)
	if err := decoder.Decode(&m); err != nil {
		return &m, err
	}

	if m.Type() != GaugeType && m.Type() != CounterType {
		return &m, errrs.ThrowInvalidTypeError(m.Type())
	}

	return &m, nil
}

func New(n string, t string, v float64, d int64) Metric {
	return &metric{
		ID:    n,
		MType: t,
		Value: &v,
		Delta: &d,
	}
}
