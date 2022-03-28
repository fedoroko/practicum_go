package storage

import (
	"testing"
	"github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestMemStats_update(t *testing.T) {
	type fields struct {
		t string
		n string
		v string
	}
	tests := []struct{
		name string
		fields fields
		wantErr bool
	}{
		{
			name: "default",
			fields: fields{
				t: "gauge",
				n: "Alloc",
				v: "10",
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			fields: fields{
				t: "yolo",
				n: "Alloc",
				v: "10",
			},
			wantErr: true,
		},
		{
			name: "invalid name",
			fields: fields{
				t: "gauge",
				n: "alloc",
				v: "10",
			},
			wantErr: true,
		},
		{
			name: "invalid value",
			fields: fields{
				t: "gauge",
				n: "Alloc",
				v: "yolo",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dummyStorage := MemStats{}
			err := dummyStorage.update(tt.fields.t, tt.fields.n, tt.fields.v)
			if !tt.wantErr {
				require.NoError(t, err)
				return 
			}

			assert.NotEqual(t, dummyStorage, new(MemStats))
		})
	}
}