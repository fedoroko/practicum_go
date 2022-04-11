package storage

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
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
			got := Init()
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
	cfg := &config{
		Restore:       false,
		StoreInterval: 300 * time.Second,
		StoreFile:     "/tmp/devops-metrics-db.json",
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := repoInterface(cfg)
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
	type args struct {
		i input
		o output
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "positive plain",
			fields: fields{
				g: map[string]gauge{
					"alloc": gauge(1),
				},
			},
			args: args{
				i: Raw("gauge", "Alloc"),
				o: Plain(),
			},
			want:    "1",
			wantErr: false,
		},
		{
			name: "wrong type plain",
			fields: fields{
				g: map[string]gauge{
					"alloc": gauge(1),
				},
			},
			args: args{
				i: Raw("int", "Alloc"),
				o: Plain(),
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "wrong name plain",
			fields: fields{
				g: map[string]gauge{
					"alloc": gauge(1),
				},
			},
			args: args{
				i: Raw("gauge", "zlloc"),
				o: Plain(),
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "positive json",
			fields: fields{
				g: map[string]gauge{
					"alloc": gauge(1),
				},
			},
			args: args{
				i: FromMetric(&Metrics{
					ID:    "Alloc",
					MType: "gauge",
				}),
				o: ToJSON(),
			},
			want:    "{\"id\":\"Alloc\",\"type\":\"gauge\",\"value\":1}",
			wantErr: false,
		},
		{
			name: "wrong type json",
			fields: fields{
				g: map[string]gauge{
					"alloc": gauge(1),
				},
			},
			args: args{
				i: FromMetric(&Metrics{
					ID:    "Alloc",
					MType: "int",
				}),
				o: ToJSON(),
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "wrong name json",
			fields: fields{
				g: map[string]gauge{
					"alloc": gauge(1),
				},
			},
			args: args{
				i: FromMetric(&Metrics{
					ID:    "Zlloc",
					MType: "gauge",
				}),
				o: ToJSON(),
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &repo{
				G:    tt.fields.g,
				gMtx: sync.RWMutex{},
				C:    tt.fields.c,
				cMtx: sync.RWMutex{},
			}
			got, err := r.Get(tt.args.i, tt.args.o)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
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
			r := Init()
			got := r.List()
			assert.NotEqual(t, "", got)
		})
	}
}

func Test_repo_Set(t *testing.T) {
	dummy := float64(1)
	type args struct {
		i input
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "positive plain",
			args: args{
				i: RawWithValue("gauge", "Alloc", "1"),
			},
			wantErr: false,
		},
		{
			name: "wrong type plain",
			args: args{
				i: RawWithValue("int", "Alloc", "1"),
			},
			wantErr: true,
		},
		{
			name: "empty value plain",
			args: args{
				i: RawWithValue("gauge", "Alloc", ""),
			},
			wantErr: true,
		},
		{
			name: "non numeric value plain",
			args: args{
				i: RawWithValue("gauge", "Alloc", "none"),
			},
			wantErr: true,
		},
		{
			name: "positive json",
			args: args{
				i: FromMetric(&Metrics{
					ID:    "Alloc",
					MType: "gauge",
					Value: &dummy,
				}),
			},
			wantErr: false,
		},
		{
			name: "wrong type json",
			args: args{
				i: FromMetric(&Metrics{
					ID:    "Alloc",
					MType: "int",
					Value: &dummy,
				}),
			},
			wantErr: true,
		},
		{
			name: "empty value json",
			args: args{
				i: FromMetric(&Metrics{
					ID:    "Alloc",
					MType: "gauge",
					Value: nil,
				}),
			},
			wantErr: true,
		},
	}
	r := Init()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := r.Set(tt.args.i)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
