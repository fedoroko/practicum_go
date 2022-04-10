package agent

import (
	"fmt"
	"os"
	"strconv"
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

func WithEnv() option {
	a := "http://127.0.0.1:8080"
	p := time.Duration(2)
	r := time.Duration(10)
	address := os.Getenv("ADDRESS")
	pollInterval := os.Getenv("POLL_INTERVAL")
	reportInterval := os.Getenv("REPORT_INTERVAL")
	if address != "" {
		a = "http://" + address
	}
	if pollInterval != "" {
		i, err := strconv.ParseInt(pollInterval, 10, 64)
		if err == nil {
			p = time.Duration(i)
		}
	}
	if reportInterval != "" {
		i, err := strconv.ParseInt(reportInterval, 10, 64)
		if err == nil {
			r = time.Duration(i)
		}
	}
	return func(c *config) {
		c.endpoint = a
		c.pollInterval = p
		c.reportInterval = r
	}
}

func Run(opts ...option) {
	cfg := &config{
		pollInterval:     2,
		reportInterval:   10,
		shutdownInterval: 200,
		contentType:      ContentTypePlain,
		endpoint:         "http://127.0.0.1:8080",
	}

	for _, o := range opts {
		o(cfg)
	}
	fmt.Println(cfg)
	s := newStats(cfg)

	go s.collect()
	s.send()
}
