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
	Id    string  `json:"id"`
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
						Id:    "Alloc",
						MType: "gauge",
						Value: float64(currentStats.Alloc),
					},
					{
						Id:    "BuckHashSys",
						MType: "gauge",
						Value: float64(currentStats.BuckHashSys),
					},
					{
						Id:    "Frees",
						MType: "gauge",
						Value: float64(currentStats.Frees),
					},
					{
						Id:    "GCCPUFraction",
						MType: "gauge",
						Value: currentStats.GCCPUFraction,
					},
					{
						Id:    "GCSys",
						MType: "gauge",
						Value: float64(currentStats.GCSys),
					},
					{
						Id:    "HeapAlloc",
						MType: "gauge",
						Value: float64(currentStats.HeapAlloc),
					},
					{
						Id:    "HeapIdle",
						MType: "gauge",
						Value: float64(currentStats.HeapIdle),
					},
					{
						Id:    "HeapInuse",
						MType: "gauge",
						Value: float64(currentStats.HeapInuse),
					},
					{
						Id:    "HeapObjects",
						MType: "gauge",
						Value: float64(currentStats.HeapObjects),
					},
					{
						Id:    "HeapReleased",
						MType: "gauge",
						Value: float64(currentStats.HeapReleased),
					},
					{
						Id:    "HeapSys",
						MType: "gauge",
						Value: float64(currentStats.HeapSys),
					},
					{
						Id:    "LastGC",
						MType: "gauge",
						Value: float64(currentStats.LastGC),
					},
					{
						Id:    "Lookups",
						MType: "gauge",
						Value: float64(currentStats.Lookups),
					},
					{
						Id:    "MCacheInuse",
						MType: "gauge",
						Value: float64(currentStats.MCacheInuse),
					},
					{
						Id:    "MCacheSys",
						MType: "gauge",
						Value: float64(currentStats.MCacheSys),
					},
					{
						Id:    "MSpanInuse",
						MType: "gauge",
						Value: float64(currentStats.MSpanInuse),
					},
					{
						Id:    "MSpanSys",
						MType: "gauge",
						Value: float64(currentStats.MSpanSys),
					},
					{
						Id:    "Mallocs",
						MType: "gauge",
						Value: float64(currentStats.Mallocs),
					},
					{
						Id:    "NextGC",
						MType: "gauge",
						Value: float64(currentStats.NextGC),
					},
					{
						Id:    "NumForcedGC",
						MType: "gauge",
						Value: float64(currentStats.NumForcedGC),
					},
					{
						Id:    "NumGC",
						MType: "gauge",
						Value: float64(currentStats.NumGC),
					},
					{
						Id:    "OtherSys",
						MType: "gauge",
						Value: float64(currentStats.OtherSys),
					},
					{
						Id:    "PauseTotalNs",
						MType: "gauge",
						Value: float64(currentStats.PauseTotalNs),
					},
					{
						Id:    "StackInuse",
						MType: "gauge",
						Value: float64(currentStats.StackInuse),
					},
					{
						Id:    "StackSys",
						MType: "gauge",
						Value: float64(currentStats.StackSys),
					},
					{
						Id:    "Sys",
						MType: "gauge",
						Value: float64(currentStats.Sys),
					},
					{
						Id:    "TotalAlloc",
						MType: "gauge",
						Value: float64(currentStats.TotalAlloc),
					},
					{
						Id:    "RandomValue",
						MType: "gauge",
						Value: rand.Float64(),
					},
				}

				s.count += int64(len(s.metrics) - 1)
				s.metrics = append(s.metrics, metric{
					Id:    "PollCount",
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
	case ContentTypeJson:
		jsonRequest(c, cfg, m)
	default:
		plainRequest(c, cfg, m)
	}
}

func jsonRequest(c *http.Client, cfg *config, m metric) {
	url := cfg.endpoint + "/update/"
	data, err := json.Marshal(m)
	if err != nil {
		log.Fatal(err)
	}

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("Content-Type", ContentTypeJson)

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
	url := cfg.endpoint + "/update/" + m.MType + "/" + m.Id + "/"
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
