package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/fedoroko/practicum_go/internal/config"
	"github.com/fedoroko/practicum_go/internal/metrics"
)

type stats struct {
	metrics []metrics.Metric
	count   int64
	mtx     sync.RWMutex
	done    chan struct{}
	cfg     *config.AgentConfig
	logger  *config.Logger
}

func newStats(cfg *config.AgentConfig, logger *config.Logger) *stats {
	return &stats{
		metrics: []metrics.Metric{},
		count:   0,
		mtx:     sync.RWMutex{},
		done:    make(chan struct{}),
		cfg:     cfg,
		logger:  logger,
	}
}

func (s *stats) collect() {
	var currentStats runtime.MemStats

	pollTicker := time.NewTicker(s.cfg.PollInterval)
	shutdownTimer := time.NewTimer(s.cfg.ShutdownInterval)
	defer pollTicker.Stop()
	defer shutdownTimer.Stop()

	for {
		select {
		case <-shutdownTimer.C:
			s.done <- struct{}{}
			return

		case <-pollTicker.C:
			func() {
				runtime.ReadMemStats(&currentStats)
				s.mtx.RLock()
				defer s.mtx.RUnlock()
				s.metrics = []metrics.Metric{
					metrics.NewOmitEmpty(
						"Alloc",
						"gauge",
						metrics.PointerFromFloat64(float64(currentStats.Alloc)),
						nil,
					),

					metrics.NewOmitEmpty(
						"BuckHashSys",
						"gauge",
						metrics.PointerFromFloat64(float64(currentStats.BuckHashSys)),
						nil,
					),

					metrics.NewOmitEmpty(
						"Frees",
						"gauge",
						metrics.PointerFromFloat64(float64(currentStats.Frees)),
						nil,
					),

					metrics.NewOmitEmpty(
						"GCCPUFraction",
						"gauge",
						metrics.PointerFromFloat64(currentStats.GCCPUFraction),
						nil,
					),

					metrics.NewOmitEmpty(
						"GCSys",
						"gauge",
						metrics.PointerFromFloat64(float64(currentStats.GCSys)),
						nil,
					),

					metrics.NewOmitEmpty(
						"HeapAlloc",
						"gauge",
						metrics.PointerFromFloat64(float64(currentStats.HeapAlloc)),
						nil,
					),

					metrics.NewOmitEmpty(
						"HeapIdle",
						"gauge",
						metrics.PointerFromFloat64(float64(currentStats.HeapIdle)),
						nil,
					),

					metrics.NewOmitEmpty(
						"HeapInuse",
						"gauge",
						metrics.PointerFromFloat64(float64(currentStats.HeapInuse)),
						nil,
					),

					metrics.NewOmitEmpty(
						"HeapObjects",
						"gauge",
						metrics.PointerFromFloat64(float64(currentStats.HeapObjects)),
						nil,
					),

					metrics.NewOmitEmpty(
						"HeapReleased",
						"gauge",
						metrics.PointerFromFloat64(float64(currentStats.HeapReleased)),
						nil,
					),

					metrics.NewOmitEmpty(
						"HeapSys",
						"gauge",
						metrics.PointerFromFloat64(float64(currentStats.HeapSys)),
						nil,
					),

					metrics.NewOmitEmpty(
						"LastGC",
						"gauge",
						metrics.PointerFromFloat64(float64(currentStats.LastGC)),
						nil,
					),

					metrics.NewOmitEmpty(
						"Lookups",
						"gauge",
						metrics.PointerFromFloat64(float64(currentStats.Lookups)),
						nil,
					),

					metrics.NewOmitEmpty(
						"MCacheInuse",
						"gauge",
						metrics.PointerFromFloat64(float64(currentStats.MCacheInuse)),
						nil,
					),

					metrics.NewOmitEmpty(
						"MCacheSys",
						"gauge",
						metrics.PointerFromFloat64(float64(currentStats.MCacheSys)),
						nil,
					),

					metrics.NewOmitEmpty(
						"MSpanInuse",
						"gauge",
						metrics.PointerFromFloat64(float64(currentStats.MSpanInuse)),
						nil,
					),

					metrics.NewOmitEmpty(
						"MSpanSys",
						"gauge",
						metrics.PointerFromFloat64(float64(currentStats.MSpanSys)),
						nil,
					),

					metrics.NewOmitEmpty(
						"Mallocs",
						"gauge",
						metrics.PointerFromFloat64(float64(currentStats.Mallocs)),
						nil,
					),

					metrics.NewOmitEmpty(
						"NextGC",
						"gauge",
						metrics.PointerFromFloat64(float64(currentStats.NextGC)),
						nil,
					),

					metrics.NewOmitEmpty(
						"NumForcedGC",
						"gauge",
						metrics.PointerFromFloat64(float64(currentStats.NumForcedGC)),
						nil,
					),

					metrics.NewOmitEmpty(
						"NumGC",
						"gauge",
						metrics.PointerFromFloat64(float64(currentStats.NumGC)),
						nil,
					),

					metrics.NewOmitEmpty(
						"OtherSys",
						"gauge",
						metrics.PointerFromFloat64(float64(currentStats.OtherSys)),
						nil,
					),

					metrics.NewOmitEmpty(
						"PauseTotalNs",
						"gauge",
						metrics.PointerFromFloat64(float64(currentStats.PauseTotalNs)),
						nil,
					),

					metrics.NewOmitEmpty(
						"StackInuse",
						"gauge",
						metrics.PointerFromFloat64(float64(currentStats.StackInuse)),
						nil,
					),

					metrics.NewOmitEmpty(
						"StackSys",
						"gauge",
						metrics.PointerFromFloat64(float64(currentStats.StackSys)),
						nil,
					),

					metrics.NewOmitEmpty(
						"Sys",
						"gauge",
						metrics.PointerFromFloat64(float64(currentStats.Sys)),
						nil,
					),

					metrics.NewOmitEmpty(
						"TotalAlloc",
						"gauge",
						metrics.PointerFromFloat64(float64(currentStats.TotalAlloc)),
						nil,
					),

					metrics.NewOmitEmpty(
						"RandomValue",
						"gauge",
						metrics.PointerFromFloat64(rand.Float64()),
						nil,
					),
				}

				s.count += int64(len(s.metrics) - 1)
				s.metrics = append(
					s.metrics, metrics.NewOmitEmpty(
						"PollCount",
						"counter",
						nil,
						metrics.PointerFromInt64(s.count),
					),
				)
			}()
		}
	}
}

func (s *stats) send() {
	client := resty.New()
	client.
		SetRetryCount(3).
		SetRetryWaitTime(20 * time.Second).
		SetRetryMaxWaitTime(100 * time.Second)

	sendTicker := time.NewTicker(s.cfg.ReportInterval)
	defer sendTicker.Stop()

	for {
		select {
		case <-s.done:
			return
		case <-sendTicker.C:
			func() {
				s.mtx.Lock()
				defer s.mtx.Unlock()
				if err := butchRequest(client, s.cfg, s.logger, s.metrics); err != nil {
					s.logger.Error().Stack().Err(err).Msg("")
				}
			}()
		}
	}
}

func requestHandler(c *resty.Client, cfg *config.AgentConfig, logger *config.Logger, m metrics.Metric) {
	switch cfg.ContentType {
	case ContentTypeJSON:
		jsonRequest(c, cfg, logger, m)
	default:
		plainRequest(c, cfg, logger, m)
	}
}

func jsonRequest(c *resty.Client, cfg *config.AgentConfig, logger *config.Logger, m metrics.Metric) {
	url := "http://" + cfg.Address + "/update"

	if err := m.SetHash(cfg.Key); err != nil {
		logger.Error().Stack().Err(err).Msg("")
	}

	data, err := json.Marshal(m)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("")
	}

	resp, err := c.R().
		SetHeader("Content-Type", ContentTypeJSON).
		SetBody(data).
		Post(url)

	if err != nil {
		logger.Error().Stack().Err(err).Msg("")
	}

	if resp.StatusCode() != http.StatusOK {
		logger.Fatal().Stack().Int("status code", resp.StatusCode()).Msg("response status code not 200")
	}
}

func plainRequest(c *resty.Client, cfg *config.AgentConfig, logger *config.Logger, m metrics.Metric) {
	url := "http://" + cfg.Address + "/update/" + m.Type() + "/" + m.Name() + "/" + m.ToString()

	resp, err := c.R().
		SetHeader("Content-Type", ContentTypePlain).
		Post(url)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("")
	}

	if resp.StatusCode() != http.StatusOK {
		logger.Fatal().Stack().Int("status code", resp.StatusCode()).Msg("response status code not 200")
	}
}

func butchRequest(c *resty.Client, cfg *config.AgentConfig, logger *config.Logger, metrics []metrics.Metric) error {
	url := "http://" + cfg.Address + "/updates"
	var data bytes.Buffer
	encoder := json.NewEncoder(&data)
	for _, m := range metrics {
		if err := m.SetHash(cfg.Key); err != nil {
			return err
		}
	}

	if err := encoder.Encode(metrics); err != nil {
		return err
	}
	logger.Debug().Str("Data:", data.String()).Send()
	resp, err := c.R().
		SetHeader("Content-Type", ContentTypeJSON).
		SetBody(data.Bytes()).
		Post(url)

	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return errors.New("wrong status code: " + fmt.Sprintf("%d", resp.StatusCode()))
	}

	return nil
}
