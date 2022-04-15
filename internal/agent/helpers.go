package agent

import (
	"encoding/json"
	"log"
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
}

func newStats(cfg *config.AgentConfig) *stats {
	return &stats{
		metrics: []metrics.Metric{},
		count:   0,
		mtx:     sync.RWMutex{},
		done:    make(chan struct{}),
		cfg:     cfg,
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
					metrics.New(
						"Alloc",
						"gauge",
						float64(currentStats.Alloc),
						0,
					),

					metrics.New(
						"BuckHashSys",
						"gauge",
						float64(currentStats.BuckHashSys),
						0,
					),

					metrics.New(
						"Frees",
						"gauge",
						float64(currentStats.Frees),
						0,
					),

					metrics.New(
						"GCCPUFraction",
						"gauge",
						currentStats.GCCPUFraction,
						0,
					),

					metrics.New(
						"GCSys",
						"gauge",
						float64(currentStats.GCSys),
						0,
					),

					metrics.New(
						"HeapAlloc",
						"gauge",
						float64(currentStats.HeapAlloc),
						0,
					),

					metrics.New(
						"HeapIdle",
						"gauge",
						float64(currentStats.HeapIdle),
						0,
					),

					metrics.New(
						"HeapInuse",
						"gauge",
						float64(currentStats.HeapInuse),
						0,
					),

					metrics.New(
						"HeapObjects",
						"gauge",
						float64(currentStats.HeapObjects),
						0,
					),

					metrics.New(
						"HeapReleased",
						"gauge",
						float64(currentStats.HeapReleased),
						0,
					),

					metrics.New(
						"HeapSys",
						"gauge",
						float64(currentStats.HeapSys),
						0,
					),

					metrics.New(
						"LastGC",
						"gauge",
						float64(currentStats.LastGC),
						0,
					),

					metrics.New(
						"Lookups",
						"gauge",
						float64(currentStats.Lookups),
						0,
					),

					metrics.New(
						"MCacheInuse",
						"gauge",
						float64(currentStats.MCacheInuse),
						0,
					),

					metrics.New(
						"MCacheSys",
						"gauge",
						float64(currentStats.MCacheSys),
						0,
					),

					metrics.New(
						"MSpanInuse",
						"gauge",
						float64(currentStats.MSpanInuse),
						0,
					),

					metrics.New(
						"MSpanSys",
						"gauge",
						float64(currentStats.MSpanSys),
						0,
					),

					metrics.New(
						"Mallocs",
						"gauge",
						float64(currentStats.Mallocs),
						0,
					),

					metrics.New(
						"NextGC",
						"gauge",
						float64(currentStats.NextGC),
						0,
					),

					metrics.New(
						"NumForcedGC",
						"gauge",
						float64(currentStats.NumForcedGC),
						0,
					),

					metrics.New(
						"NumGC",
						"gauge",
						float64(currentStats.NumGC),
						0,
					),

					metrics.New(
						"OtherSys",
						"gauge",
						float64(currentStats.OtherSys),
						0,
					),

					metrics.New(
						"PauseTotalNs",
						"gauge",
						float64(currentStats.PauseTotalNs),
						0,
					),

					metrics.New(
						"StackInuse",
						"gauge",
						float64(currentStats.StackInuse),
						0,
					),

					metrics.New(
						"StackSys",
						"gauge",
						float64(currentStats.StackSys),
						0,
					),

					metrics.New(
						"Sys",
						"gauge",
						float64(currentStats.Sys),
						0,
					),

					metrics.New(
						"TotalAlloc",
						"gauge",
						float64(currentStats.TotalAlloc),
						0,
					),

					metrics.New(
						"RandomValue",
						"gauge",
						rand.Float64(),
						0,
					),
				}

				s.count += int64(len(s.metrics) - 1)
				s.metrics = append(
					s.metrics, metrics.New(
						"PollCount",
						"counter",
						0,
						s.count,
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
				for _, m := range s.metrics {
					requestHandler(client, s.cfg, m)
				}
			}()
		}
	}
}

func requestHandler(c *resty.Client, cfg *config.AgentConfig, m metrics.Metric) {
	switch cfg.ContentType {
	case ContentTypeJSON:
		jsonRequest(c, cfg, m)
	default:
		plainRequest(c, cfg, m)
	}
}

func jsonRequest(c *resty.Client, cfg *config.AgentConfig, m metrics.Metric) {
	url := "http://" + cfg.Address + "/update"
	data, err := json.Marshal(m)
	if err != nil {
		log.Println(err)
	}

	resp, err := c.R().
		SetHeader("Content-Type", ContentTypeJSON).
		SetBody(data).
		Post(url)

	if err != nil {
		log.Println(err, "resp")
	}

	if resp.StatusCode() != http.StatusOK {
		log.Fatal("Wrong Status Code")
	}
}

func plainRequest(c *resty.Client, cfg *config.AgentConfig, m metrics.Metric) {
	url := "http://" + cfg.Address + "/update/" + m.Type() + "/" + m.Name() + "/" + m.ToString()

	resp, err := c.R().
		SetHeader("Content-Type", ContentTypePlain).
		Post(url)
	if err != nil {
		log.Println(err)
	}

	if resp.StatusCode() != http.StatusOK {
		log.Println("Wrong Status Code")
	}
}
