package agent

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func Test_newStats(t *testing.T) {
	type args struct {
		cfg *config
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "positive",
			args: args{
				cfg: &config{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newStats(tt.args.cfg)
			assert.NotEqual(t, got, stats{})
		})
	}
}

func Test_stats_collect(t *testing.T) {
	type fields struct {
		metrics []metric
		count   int64
		mtx     *sync.RWMutex
		done    chan struct{}
		cfg     *config
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "positive",
			fields: fields{
				metrics: []metric{},
				cfg: &config{
					pollInterval:     1,
					shutdownInterval: 2,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stats{
				metrics: tt.fields.metrics,
				count:   tt.fields.count,
				mtx:     sync.RWMutex{},
				done:    tt.fields.done,
				cfg:     tt.fields.cfg,
			}
			empty := &stats{
				metrics: []metric{},
			}

			go s.collect()
			time.Sleep(s.cfg.shutdownInterval * time.Second)

			assert.NotEqual(t, s.metrics, empty.metrics)
		})
	}
}
