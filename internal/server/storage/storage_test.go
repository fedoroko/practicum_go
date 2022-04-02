package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDummyDBInterface(t *testing.T) {
	type args struct {
		g *gaugeStorage
		c *counterStorage
	}
	tests := []struct {
		name string
		args args
		want *DummyDB
	}{
		{
			name: "positive",
			args: args{
				g: &gaugeStorage{
					Fields: map[string]gauge{},
				},
				c: &counterStorage{
					Fields: map[string]counter{},
				},
			},
			want: &DummyDB{
				G: &gaugeStorage{
					Fields: map[string]gauge{},
				},
				C: &counterStorage{
					Fields: map[string]counter{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, DummyDBInterface(tt.args.g, tt.args.c), "DummyDBInterface(%v, %v)", tt.args.g, tt.args.c)
		})
	}
}

func TestDummyDB_Display(t *testing.T) {
	type fields struct {
		G *gaugeStorage
		C *counterStorage
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "positive",
			fields: fields{
				G: &gaugeStorage{
					Fields: map[string]gauge{
						"alloc": gauge(1),
					},
				},
				C: &counterStorage{
					Fields: map[string]counter{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &DummyDB{
				G: tt.fields.G,
				C: tt.fields.C,
			}
			assert.NotEqual(t, "", db.Display(), "Display()")
		})
	}
}

func TestDummyDB_Get(t *testing.T) {
	type fields struct {
		G *gaugeStorage
		C *counterStorage
	}
	type args struct {
		t string
		n string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "positive",
			fields: fields{
				G: &gaugeStorage{
					Fields: map[string]gauge{
						"alloc": gauge(1),
					},
				},
				C: &counterStorage{
					Fields: map[string]counter{},
				},
			},
			args: args{
				t: "gauge",
				n: "alloc",
			},
			want:    "1",
			wantErr: false,
		},
		{
			name: "wrong type",
			fields: fields{
				G: &gaugeStorage{
					Fields: map[string]gauge{
						"alloc": gauge(1),
					},
				},
				C: &counterStorage{
					Fields: map[string]counter{},
				},
			},
			args: args{
				t: "int",
				n: "alloc",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "wrong key",
			fields: fields{
				G: &gaugeStorage{
					Fields: map[string]gauge{
						"alloc": gauge(1),
					},
				},
				C: &counterStorage{
					Fields: map[string]counter{},
				},
			},
			args: args{
				t: "gauge",
				n: "bulloc",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &DummyDB{
				G: tt.fields.G,
				C: tt.fields.C,
			}
			got, err := db.Get(tt.args.t, tt.args.n)
			assert.Equal(t, tt.want, got)
			if !tt.wantErr {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDummyDB_Set(t *testing.T) {
	type fields struct {
		G *gaugeStorage
		C *counterStorage
	}
	type args struct {
		t string
		n string
		v string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "positive",
			fields: fields{
				G: &gaugeStorage{
					Fields: map[string]gauge{},
				},
				C: &counterStorage{
					Fields: map[string]counter{},
				},
			},
			args: args{
				t: "gauge",
				n: "alloc",
				v: "1",
			},
			wantErr: false,
		},
		{
			name: "wrong type",
			fields: fields{
				G: &gaugeStorage{
					Fields: map[string]gauge{},
				},
				C: &counterStorage{
					Fields: map[string]counter{},
				},
			},
			args: args{
				t: "int",
				n: "alloc",
				v: "1",
			},
			wantErr: true,
		},
		{
			name: "wrong value",
			fields: fields{
				G: &gaugeStorage{
					Fields: map[string]gauge{},
				},
				C: &counterStorage{
					Fields: map[string]counter{},
				},
			},
			args: args{
				t: "gauge",
				n: "alloc",
				v: "none",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &DummyDB{
				G: tt.fields.G,
				C: tt.fields.C,
			}
			err := db.Set(tt.args.t, tt.args.n, tt.args.v)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestInit(t *testing.T) {
	tests := []struct {
		name string
		want *DummyDB
	}{
		{
			name: "positive",
			want: &DummyDB{
				G: &gaugeStorage{
					Fields: map[string]gauge{},
				},
				C: &counterStorage{
					Fields: map[string]counter{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, Init(), "Init()")
		})
	}
}
