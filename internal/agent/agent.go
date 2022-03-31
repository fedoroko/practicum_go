package agent

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

type memStat struct {
	Name  string
	Value string
	Type  string
}

var stats []memStat

var pollInterval time.Duration = 2

var reportInterval time.Duration = 10

var shutdownInterval time.Duration = 60

var endpoint = "http://127.0.0.1:8080"

var pollCount int64 = 0

func collectMemStats() {

	var currentStats runtime.MemStats
	runtime.ReadMemStats(&currentStats)

	stats = []memStat{
		{
			"Alloc", fmt.Sprintf("%v", currentStats.Alloc), "gauge",
		},
		{
			"BuckHashSys", fmt.Sprintf("%v", currentStats.BuckHashSys), "gauge",
		},
		{
			"Frees", fmt.Sprintf("%v", currentStats.Frees), "gauge",
		},
		{
			"GCCPUFraction", fmt.Sprintf("%v", currentStats.GCCPUFraction), "gauge",
		},
		{
			"GCSys", fmt.Sprintf("%v", currentStats.GCSys), "gauge",
		},
		{
			"HeapAlloc", fmt.Sprintf("%v", currentStats.HeapAlloc), "gauge",
		},
		{
			"HeapIdle", fmt.Sprintf("%v", currentStats.HeapIdle), "gauge",
		},
		{
			"HeapInuse", fmt.Sprintf("%v", currentStats.HeapInuse), "gauge",
		},
		{
			"HeapObjects", fmt.Sprintf("%v", currentStats.HeapObjects), "gauge",
		},
		{
			"HeapReleased", fmt.Sprintf("%v", currentStats.HeapReleased), "gauge",
		},
		{
			"HeapSys", fmt.Sprintf("%v", currentStats.HeapSys), "gauge",
		},
		{
			"LastGC", fmt.Sprintf("%v", currentStats.LastGC), "gauge",
		},
		{
			"Lookups", fmt.Sprintf("%v", currentStats.Lookups), "gauge",
		},
		{
			"MCacheInuse", fmt.Sprintf("%v", currentStats.MCacheInuse), "gauge",
		},
		{
			"MCacheSys", fmt.Sprintf("%v", currentStats.MCacheSys), "gauge",
		},
		{
			"MSpanInuse", fmt.Sprintf("%v", currentStats.MSpanInuse), "gauge",
		},
		{
			"MSpanSys", fmt.Sprintf("%v", currentStats.MSpanSys), "gauge",
		},
		{
			"Mallocs", fmt.Sprintf("%v", currentStats.Mallocs), "gauge",
		},
		{
			"NextGC", fmt.Sprintf("%v", currentStats.NextGC), "gauge",
		},
		{
			"NumForcedGC", fmt.Sprintf("%v", currentStats.NumForcedGC), "gauge",
		},
		{
			"NumGC", fmt.Sprintf("%v", currentStats.NumGC), "gauge",
		},
		{
			"OtherSys", fmt.Sprintf("%v", currentStats.OtherSys), "gauge",
		},
		{
			"PauseTotalNs", fmt.Sprintf("%v", currentStats.PauseTotalNs), "gauge",
		},
		{
			"StackInuse", fmt.Sprintf("%v", currentStats.StackInuse), "gauge",
		},
		{
			"StackSys", fmt.Sprintf("%v", currentStats.StackSys), "gauge",
		},
		{
			"Sys", fmt.Sprintf("%v", currentStats.Sys), "gauge",
		},
		{
			"TotalAlloc", fmt.Sprintf("%v", currentStats.TotalAlloc), "gauge",
		},
		{
			"PollCount", fmt.Sprintf("%v", pollCount), "counter",
		},
		{
			"RandomValue", fmt.Sprintf("%v", rand.Float64()), "gauge",
		},
	}
	pollCount++
}

func sendMemStats() {
	client := &http.Client{}
	for _, stat := range stats {
		url := endpoint + "/update/" + stat.Type + "/" + stat.Name + "/" + stat.Value

		request, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			log.Fatal(err)
		}
		request.Header.Set("Content-Type", "text/plain")

		response, err := client.Do(request)
		if err != nil {
			log.Fatal(err)
		}

		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			log.Fatal("Wrong Status Code")
		}
	}
}

func Run() {
	collectTicker := time.NewTicker(pollInterval * time.Second)
	sendTicker := time.NewTicker(reportInterval * time.Second)
	shutdownTicker := time.NewTicker(shutdownInterval * time.Minute)

	for {
		select {
		case <-collectTicker.C:
			collectMemStats()

		case <-sendTicker.C:
			sendMemStats()

		case <-shutdownTicker.C:
			collectTicker.Stop()
			sendTicker.Stop()
			shutdownTicker.Stop()
			break
		}
	}
}
