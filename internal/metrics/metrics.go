package metrics

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/fedoroko/practicum_go/internal/errrs"
	"io"
	"strconv"
)

const (
	GaugeType   = "gauge"
	CounterType = "counter"
)

type Metric interface {
	Name() string
	Type() string
	Float64Value() float64
	Int64Value() int64

	SetFloat64(float64)
	SetInt64(int64)

	SetHash(string) error
	CheckHash(string) (bool, error)

	ToString() string
	ToJSON() []byte
}

type metric struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

func (m *metric) Name() string {
	return m.ID
}

func (m *metric) Type() string {
	return m.MType
}

func (m *metric) Float64Value() float64 {
	if m.Value == nil {
		return 0
	}
	return *m.Value
}

func (m *metric) Int64Value() int64 {
	if m.Delta == nil {
		return 0
	}
	return *m.Delta
}

func (m *metric) SetFloat64(f float64) {
	m.Value = &f
}

func (m *metric) SetInt64(i int64) {
	m.Delta = &i
}

func (m *metric) SetHash(key string) error {
	data := getHashSrc(m)

	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	hash := h.Sum(nil)
	m.Hash = hex.EncodeToString(hash)
	return nil
}

func (m *metric) CheckHash(key string) (bool, error) {
	if m.Hash == "" {
		return true, nil
	}

	data := getHashSrc(m)

	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	hash := h.Sum(nil)
	currHash, err := hex.DecodeString(m.Hash)
	if err != nil {
		return false, err
	}

	return hmac.Equal(hash, currHash), nil
}

func getHashSrc(m *metric) []byte {
	var data []byte
	switch m.Type() {
	case GaugeType:
		data = []byte(fmt.Sprintf("%s:gauge:%f", m.Name(), m.Float64Value()))
	case CounterType:
		data = []byte(fmt.Sprintf("%s:counter:%d", m.Name(), m.Int64Value()))
	}

	return data
}

func (m *metric) ToString() string {
	switch m.Type() {
	case GaugeType:
		if m.Value != nil {
			return fmt.Sprintf("%v", *m.Value)
		}
	case CounterType:
		if m.Delta != nil {
			return fmt.Sprintf("%v", *m.Delta)
		}
	}
	return ""
}

// ToJSON ожидаю, что сериализации метрики будет пердсказуемой,
// поэтому не возвращаю ошибку
func (m *metric) ToJSON() []byte {
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}

	return b
}

func RawWithValue(t string, n string, v string) (Metric, error) {
	m := &metric{
		ID:    n,
		MType: t,
	}

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

func NewOmitEmpty(n string, t string, v *float64, d *int64) Metric {
	return &metric{
		ID:    n,
		MType: t,
		Value: v,
		Delta: d,
	}
}

func PointerFromFloat64(v float64) *float64 {
	return &v
}

func PointerFromInt64(v int64) *int64 {
	return &v
}
