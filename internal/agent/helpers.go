package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

type metric struct {
	ID    string  `json:"id"`
	MType string  `json:"type"`
	Delta int64   `json:"delta"`
	Value float64 `json:"value"`
}

type stats struct {
	metrics []metric
	count   int64
	mtx     sync.RWMutex
	done    chan struct{}
	cfg     *config
}

func newStats(cfg *config) *stats {
	return &stats{
		metrics: []metric{},
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
				s.metrics = []metric{
					{
						ID:    "Alloc",
						MType: "gauge",
						Value: float64(currentStats.Alloc),
					},
					{
						ID:    "BuckHashSys",
						MType: "gauge",
						Value: float64(currentStats.BuckHashSys),
					},
					{
						ID:    "Frees",
						MType: "gauge",
						Value: float64(currentStats.Frees),
					},
					{
						ID:    "GCCPUFraction",
						MType: "gauge",
						Value: currentStats.GCCPUFraction,
					},
					{
						ID:    "GCSys",
						MType: "gauge",
						Value: float64(currentStats.GCSys),
					},
					{
						ID:    "HeapAlloc",
						MType: "gauge",
						Value: float64(currentStats.HeapAlloc),
					},
					{
						ID:    "HeapIdle",
						MType: "gauge",
						Value: float64(currentStats.HeapIdle),
					},
					{
						ID:    "HeapInuse",
						MType: "gauge",
						Value: float64(currentStats.HeapInuse),
					},
					{
						ID:    "HeapObjects",
						MType: "gauge",
						Value: float64(currentStats.HeapObjects),
					},
					{
						ID:    "HeapReleased",
						MType: "gauge",
						Value: float64(currentStats.HeapReleased),
					},
					{
						ID:    "HeapSys",
						MType: "gauge",
						Value: float64(currentStats.HeapSys),
					},
					{
						ID:    "LastGC",
						MType: "gauge",
						Value: float64(currentStats.LastGC),
					},
					{
						ID:    "Lookups",
						MType: "gauge",
						Value: float64(currentStats.Lookups),
					},
					{
						ID:    "MCacheInuse",
						MType: "gauge",
						Value: float64(currentStats.MCacheInuse),
					},
					{
						ID:    "MCacheSys",
						MType: "gauge",
						Value: float64(currentStats.MCacheSys),
					},
					{
						ID:    "MSpanInuse",
						MType: "gauge",
						Value: float64(currentStats.MSpanInuse),
					},
					{
						ID:    "MSpanSys",
						MType: "gauge",
						Value: float64(currentStats.MSpanSys),
					},
					{
						ID:    "Mallocs",
						MType: "gauge",
						Value: float64(currentStats.Mallocs),
					},
					{
						ID:    "NextGC",
						MType: "gauge",
						Value: float64(currentStats.NextGC),
					},
					{
						ID:    "NumForcedGC",
						MType: "gauge",
						Value: float64(currentStats.NumForcedGC),
					},
					{
						ID:    "NumGC",
						MType: "gauge",
						Value: float64(currentStats.NumGC),
					},
					{
						ID:    "OtherSys",
						MType: "gauge",
						Value: float64(currentStats.OtherSys),
					},
					{
						ID:    "PauseTotalNs",
						MType: "gauge",
						Value: float64(currentStats.PauseTotalNs),
					},
					{
						ID:    "StackInuse",
						MType: "gauge",
						Value: float64(currentStats.StackInuse),
					},
					{
						ID:    "StackSys",
						MType: "gauge",
						Value: float64(currentStats.StackSys),
					},
					{
						ID:    "Sys",
						MType: "gauge",
						Value: float64(currentStats.Sys),
					},
					{
						ID:    "TotalAlloc",
						MType: "gauge",
						Value: float64(currentStats.TotalAlloc),
					},
					{
						ID:    "RandomValue",
						MType: "gauge",
						Value: rand.Float64(),
					},
				}

				s.count += int64(len(s.metrics) - 1)
				s.metrics = append(s.metrics, metric{
					ID:    "PollCount",
					MType: "counter",
					Delta: s.count,
				})
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

func requestHandler(c *resty.Client, cfg *config, m metric) {
	switch cfg.ContentType {
	case ContentTypeJSON:
		jsonRequest(c, cfg, m)
	default:
		plainRequest(c, cfg, m)
	}
}

func jsonRequest(c *resty.Client, cfg *config, m metric) {
	url := "http://" + cfg.Address + "/update"
	data, err := json.Marshal(m)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := c.R().
		SetHeader("Content-Type", ContentTypeJSON).
		SetBody(data).
		Post(url)

	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode() != http.StatusOK {
		log.Fatal("Wrong Status Code")
	}
}

func plainRequest(c *resty.Client, cfg *config, m metric) {
	url := "http://" + cfg.Address + "/update/" + m.MType + "/" + m.ID + "/"
	switch m.MType {
	case "counter":
		url += fmt.Sprintf("%v", m.Delta)
	default:
		url += fmt.Sprintf("%v", m.Value)
	}

	resp, err := c.R().
		SetHeader("Content-Type", ContentTypePlain).
		Post(url)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode() != http.StatusOK {
		log.Fatal("Wrong Status Code")
	}
}
