package storage

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_update(t *testing.T) {
	type args struct {
		r repositories
		n string
		v string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "positive test #1",
			args: args{
				r: &dummyGaugeStorage,
				n: "Alloc",
				v: "1",
			},
			wantErr: false,
		},
		{
			name: "positive test #2",
			args: args{
				r: &dummyCounterStorage,
				n: "PollCount",
				v: "1",
			},
			wantErr: false,
		},
		{
			name: "none value",
			args: args{
				r: &dummyGaugeStorage,
				n: "Alloc",
				v: "",
			},
			wantErr: true,
		},
		{
			name: "wrong value",
			args: args{
				r: &dummyCounterStorage,
				n: "PollCounter",
				v: "none",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := update(tt.args.r, tt.args.n, tt.args.v)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}

}

func Test_collect(t *testing.T) {
	type args struct {
		r repositories
		n string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "positive test #1",
			args: args{
				r: &dummyGaugeStorage,
				n: "Alloc",
			},
			want:    "1",
			wantErr: false,
		},
		{
			name: "unknown name",
			args: args{
				r: &dummyGaugeStorage,
				n: "SubProccess",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := collect(tt.args.r, tt.args.n)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equalf(t, tt.want, got, "collect(%v, %v)", tt.args.r, tt.args.n)
		})
	}
}
