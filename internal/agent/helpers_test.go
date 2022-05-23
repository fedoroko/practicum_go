package agent

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/fedoroko/practicum_go/internal/config"
	"github.com/fedoroko/practicum_go/internal/metrics"
)

func Test_newStats(t *testing.T) {
	cfg := config.NewAgentConfig()
	cfg.PollInterval = time.Second * 1
	type args struct {
		cfg *config.AgentConfig
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "positive",
			args: args{
				cfg: cfg,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := config.TestLogger()
			got := newStats(tt.args.cfg, logger)
			assert.NotEqual(t, got, stats{})
		})
	}
}

func Test_stats_collect(t *testing.T) {
	cfg := config.NewAgentConfig()
	cfg.PollInterval = time.Second * 1
	type fields struct {
		metrics []metrics.Metric
		count   int64
		mtx     *sync.RWMutex
		done    chan struct{}
		cfg     *config.AgentConfig
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "positive",
			fields: fields{
				metrics: []metrics.Metric{},
				cfg:     cfg,
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
				metrics: []metrics.Metric{},
			}

			go s.collect()
			time.Sleep(s.cfg.PollInterval + time.Second*1)
			assert.NotEqual(t, s.metrics, empty.metrics)
		})
	}
}
