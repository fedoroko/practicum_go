package storage

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fedoroko/practicum_go/internal/config"
	"github.com/fedoroko/practicum_go/internal/metrics"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		want Repository
	}{
		{
			name: "positive",
			want: &repo{
				G:    make(map[string]gauge),
				gMtx: sync.RWMutex{},
				C:    make(map[string]counter),
				cMtx: sync.RWMutex{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(config.NewServerConfig())
			assert.NotEqual(t, tt.want, got)
		})
	}
}

func Test_repoInterface(t *testing.T) {
	tests := []struct {
		name string
		want *repo
	}{
		{
			name: "positive",
			want: &repo{
				G:    make(map[string]gauge),
				gMtx: sync.RWMutex{},
				C:    make(map[string]counter),
				cMtx: sync.RWMutex{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := repoInterface(config.NewServerConfig())
			assert.Equal(t, tt.want.G, got.G)
			assert.Equal(t, tt.want.C, got.C)
		})
	}
}

func Test_repo_Get(t *testing.T) {
	type fields struct {
		g    map[string]gauge
		gMtx *sync.RWMutex
		c    map[string]counter
		cMtx *sync.RWMutex
	}
	tests := []struct {
		name     string
		fields   fields
		metric   metrics.Metric
		wantS    string
		wantJSON []byte
		wantErr  bool
	}{
		{
			name: "positive",
			fields: fields{
				g: map[string]gauge{
					"alloc": gauge(1),
				},
			},
			metric:   metrics.New("Alloc", "gauge", 1, 0),
			wantS:    "1",
			wantJSON: []byte("{\"id\":\"Alloc\",\"type\":\"gauge\",\"delta\":0,\"value\":1}"),
			wantErr:  false,
		},
		{
			name: "wrong type",
			fields: fields{
				g: map[string]gauge{
					"alloc": gauge(1),
				},
			},
			metric:   metrics.New("Alloc", "int", 0, 0),
			wantS:    "",
			wantJSON: []byte("{\"id\":\"Alloc\",\"type\":\"int\",\"delta\":0,\"value\":0}"),
			wantErr:  true,
		},
		{
			name: "wrong name",
			fields: fields{
				g: map[string]gauge{
					"alloc": gauge(1),
				},
			},
			metric:   metrics.New("zAlloc", "gauge", 0, 0),
			wantS:    "0",
			wantJSON: []byte("{\"id\":\"zAlloc\",\"type\":\"gauge\",\"delta\":0,\"value\":0}"),
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &repo{
				G:    tt.fields.g,
				gMtx: sync.RWMutex{},
				C:    tt.fields.c,
				cMtx: sync.RWMutex{},
				cfg: &config.ServerConfig{
					Key: "",
				},
			}
			got, err := r.Get(tt.metric)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantS, got.ToString())

			assert.Equal(t, tt.wantJSON, got.ToJSON())
		})
	}
}

func Test_repo_List(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "positive",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(config.NewServerConfig())
			got := r.List()
			assert.NotEqual(t, "", got)
		})
	}
}

func Test_repo_Set(t *testing.T) {
	tests := []struct {
		name    string
		metric  metrics.Metric
		wantErr bool
	}{
		{
			name:    "positive plain",
			metric:  metrics.New("Alloc", "gauge", 1, 0),
			wantErr: false,
		},
		{
			name:    "wrong type plain",
			metric:  metrics.New("Alloc", "int", 1, 0),
			wantErr: true,
		},
	}
	r := New(config.NewServerConfig())
	defer r.Close()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := r.Set(tt.metric)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
