package agent

import (
	"github.com/caarlos0/env/v6"
	"log"
	"time"
)

type config struct {
	PollInterval     time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
	ReportInterval   time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	ShutdownInterval time.Duration `env:"SHUTDOWN_INTERVAL" envDefault:"500s"`
	Address          string        `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
	ContentType      string
}

type option func(*config)

func WithPollInterval(i time.Duration) option {
	return func(cfg *config) {
		cfg.PollInterval = i * time.Second
	}
}

func WithReportInterval(i time.Duration) option {
	return func(cfg *config) {
		cfg.ReportInterval = i * time.Second
	}
}

func WithShutdownInterval(i time.Duration) option {
	return func(cfg *config) {
		cfg.ShutdownInterval = i * time.Second
	}
}

func WithEndpoint(addr string) option {
	return func(cfg *config) {
		cfg.Address = addr
	}
}

const (
	ContentTypeJSON  = "application/json"
	ContentTypePlain = "text.plain"
)

func WithContentType(contentType string) option {
	return func(cfg *config) {
		cfg.ContentType = contentType
	}
}

func WithEnv() option {
	return func(cfg *config) {
		err := env.Parse(cfg)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func Run(opts ...option) {
	cfg := &config{
		PollInterval:     2 * time.Second,
		ReportInterval:   10 * time.Second,
		ShutdownInterval: 200 * time.Second,
		ContentType:      ContentTypePlain,
		Address:          "127.0.0.1:8080",
	}

	for _, o := range opts {
		o(cfg)
	}

	s := newStats(cfg)

	go s.collect()
	s.send()
}
