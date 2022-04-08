package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"
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

	pollTicker := time.NewTicker(s.cfg.pollInterval * time.Second)
	shutdownTimer := time.NewTimer(s.cfg.shutdownInterval * time.Second)
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
	client := &http.Client{}
	sendTicker := time.NewTicker(s.cfg.reportInterval * time.Second)
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

func requestHandler(c *http.Client, cfg *config, m metric) {
	switch cfg.contentType {
	case ContentTypeJSON:
		jsonRequest(c, cfg, m)
	default:
		plainRequest(c, cfg, m)
	}
}

func jsonRequest(c *http.Client, cfg *config, m metric) {
	url := cfg.endpoint + "/update"
	data, err := json.Marshal(m)
	if err != nil {
		log.Fatal(err)
	}

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("Content-Type", ContentTypeJSON)

	response, err := c.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Fatal("Wrong Status Code")
	}
}

func plainRequest(c *http.Client, cfg *config, m metric) {
	url := cfg.endpoint + "/update/" + m.MType + "/" + m.ID + "/"
	switch m.MType {
	case "counter":
		url += fmt.Sprintf("%v", m.Delta)
	default:
		url += fmt.Sprintf("%v", m.Value)
	}

	request, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("Content-Type", "text/plain")

	response, err := c.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Fatal("Wrong Status Code")
	}
}
