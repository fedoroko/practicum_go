package agent

import (
	"github.com/fedoroko/practicum_go/internal/config"
	"time"
)

type option func(cfg *config.AgentConfig)

func WithPollInterval(i time.Duration) option {
	return func(cfg *config.AgentConfig) {
		cfg.PollInterval = i * time.Second
	}
}

func WithReportInterval(i time.Duration) option {
	return func(cfg *config.AgentConfig) {
		cfg.ReportInterval = i * time.Second
	}
}

func WithShutdownInterval(i time.Duration) option {
	return func(cfg *config.AgentConfig) {
		cfg.ShutdownInterval = i * time.Second
	}
}

func WithEndpoint(addr string) option {
	return func(cfg *config.AgentConfig) {
		cfg.Address = addr
	}
}

const (
	ContentTypeJSON  = "application/json"
	ContentTypePlain = "text.plain"
)

func WithContentType(contentType string) option {
	return func(cfg *config.AgentConfig) {
		cfg.ContentType = contentType
	}
}

func Run(cfg *config.AgentConfig, opts ...option) {
	for _, o := range opts {
		o(cfg)
	}
	s := newStats(cfg)

	go s.collect()
	s.send()
}
