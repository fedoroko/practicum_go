package agent

import (
	"time"

	"github.com/fedoroko/practicum_go/internal/config"
)

type option func(cfg *config.AgentConfig)

func WithPollInterval(i time.Duration) option {
	return func(cfg *config.AgentConfig) {
		cfg.PollInterval = i
	}
}

func WithReportInterval(i time.Duration) option {
	return func(cfg *config.AgentConfig) {
		cfg.ReportInterval = i
	}
}

func WithShutdownInterval(i time.Duration) option {
	return func(cfg *config.AgentConfig) {
		cfg.ShutdownInterval = i
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

func Run(cfg *config.AgentConfig, logger *config.Logger, opts ...option) {
	for _, o := range opts {
		o(cfg)
	}
	s := newStats(cfg, logger)

	go s.collect()
	s.send()
}
