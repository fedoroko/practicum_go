package storage

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"sync"
	"testing"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name string
		want Repository
	}{
		{
			name: "positive",
			want: &repo{
				g:    make(map[string]gauge),
				gMtx: sync.RWMutex{},
				c:    make(map[string]counter),
				cMtx: sync.RWMutex{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Init(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Init() = %v, want %v", got, tt.want)
			}
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
				g:    make(map[string]gauge),
				gMtx: sync.RWMutex{},
				c:    make(map[string]counter),
				cMtx: sync.RWMutex{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := repoInterface(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("repoInterface() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_repo_Get(t *testing.T) {
	type fields struct {
		g    map[string]gauge
		gMtx sync.RWMutex
		c    map[string]counter
		cMtx sync.RWMutex
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
				i: FromJSON([]byte("{\"id\":\"Alloc\",\"type\":\"gauge\"}")),
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
				i: FromJSON([]byte("{\"id\":\"Alloc\",\"type\":\"int\"}")),
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
				i: FromJSON([]byte("{\"id\":\"zlloc\",\"type\":\"gauge\"}")),
				o: ToJSON(),
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &repo{
				g:    tt.fields.g,
				gMtx: tt.fields.gMtx,
				c:    tt.fields.c,
				cMtx: tt.fields.cMtx,
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
				i: FromJSON([]byte("{\"id\":\"Alloc\",\"type\":\"gauge\",\"value\":1}")),
			},
			wantErr: false,
		},
		{
			name: "wrong type json",
			args: args{
				i: FromJSON([]byte("{\"id\":\"Alloc\",\"type\":\"int\",\"value\":1}")),
			},
			wantErr: true,
		},
		{
			name: "empty value json",
			args: args{
				i: FromJSON([]byte("{\"id\":\"Alloc\",\"type\":\"gauge\"}")),
			},
			wantErr: true,
		},
		{
			name: "non numeric value json",
			args: args{
				i: FromJSON([]byte("{\"id\":\"Alloc\",\"type\":\"gauge\",\"value\":\"none\"}")),
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
