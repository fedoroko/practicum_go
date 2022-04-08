package agent

import (
	"time"
)

type config struct {
	pollInterval     time.Duration
	reportInterval   time.Duration
	shutdownInterval time.Duration
	contentType      string
	endpoint         string
}

type option func(*config)

func WithPollInterval(i time.Duration) option {
	return func(c *config) {
		c.pollInterval = i
	}
}

func WithReportInterval(i time.Duration) option {
	return func(c *config) {
		c.reportInterval = i
	}
}

func WithShutdownInterval(i time.Duration) option {
	return func(c *config) {
		c.shutdownInterval = i
	}
}

func WithEndpoint(addr string) option {
	return func(c *config) {
		c.endpoint = addr
	}
}

const (
	ContentTypeJSON  = "application/json"
	ContentTypePlain = "text.plain"
)

func WithContentType(contentType string) option {
	return func(c *config) {
		c.contentType = contentType
	}
}

func Run(opts ...option) {
	cfg := &config{
		pollInterval:     2,
		reportInterval:   10,
		shutdownInterval: 200,
		contentType:      ContentTypePlain,
		endpoint:         "http://localhost:8080",
	}

	for _, o := range opts {
		o(cfg)
	}

	s := newStats(cfg)

	go s.collect()
	s.send()
}
