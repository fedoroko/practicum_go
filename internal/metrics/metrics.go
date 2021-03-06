package metrics

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

const (
	GaugeType   = "gauge"
	CounterType = "counter"
)

type Metric interface {
	Name() string
	DBName() string
	Type() string
	Float64Value() float64
	Float64Pointer() *float64
	Int64Value() int64
	Int64Pointer() *int64

	SetFloat64(float64)
	SetInt64(int64)

	SetHash(string) error
	CheckHash(string) (bool, error)
	CheckType() error

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

func (m *metric) DBName() string {
	return m.Name() + "::" + m.Type()
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

func (m *metric) Float64Pointer() *float64 {
	return m.Value
}

func (m *metric) Int64Value() int64 {
	if m.Delta == nil {
		return 0
	}
	return *m.Delta
}

func (m *metric) Int64Pointer() *int64 {
	return m.Delta
}

func (m *metric) SetFloat64(f float64) {
	m.Value = &f
}

func (m *metric) SetInt64(i int64) {
	m.Delta = &i
}

func (m *metric) SetHash(key string) error {
	if key == "" {
		return nil
	}
	data := getHashSrc(m)

	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	hash := h.Sum(nil)
	m.Hash = hex.EncodeToString(hash)
	return nil
}

func (m *metric) CheckHash(key string) (bool, error) {
	if m.Hash == "" || key == "" {
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

func (m *metric) CheckType() error {
	switch m.Type() {
	case GaugeType:
		return nil
	case CounterType:
		return nil
	default:
		return ThrowInvalidTypeError(m.Type())
	}
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

// ToJSON ????????????, ?????? ???????????????????????? ?????????????? ?????????? ??????????????????????????,
// ?????????????? ???? ?????????????????? ????????????
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
		return m, ThrowInvalidTypeError(t)
	}
	return m, nil
}

func Raw(t string, n string) (Metric, error) {
	m := &metric{
		ID:    n,
		MType: t,
	}
	if t != GaugeType && t != CounterType {
		return m, ThrowInvalidTypeError(t)
	}

	return m, nil
}

func FromJSON(j io.Reader) (Metric, error) {
	m := metric{}
	decoder := json.NewDecoder(j)
	if err := decoder.Decode(&m); err != nil {
		return &m, err
	}

	return &m, m.CheckType()
}

func ArrFromJSON(j io.Reader) ([]Metric, error) {
	metrics := make([]*metric, 0)
	decoder := json.NewDecoder(j)
	if err := decoder.Decode(&metrics); err != nil {
		return make([]Metric, 0), err
	}

	ret := make([]Metric, len(metrics))
	for i, m := range metrics {
		ret[i] = Metric(m)
	}

	return ret, nil
}

func Blank() Metric {
	return &metric{}
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
