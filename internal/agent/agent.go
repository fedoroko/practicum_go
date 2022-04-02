package agent

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

type gStat struct {
	Name  string
	Value float64
	Type  string
}

func collectMemStats(pollCount int64) ([]gStat, int64) {

	var currentStats runtime.MemStats
	runtime.ReadMemStats(&currentStats)

	stats := []gStat{
		{
			"Alloc", float64(currentStats.Alloc), "gauge",
		},
		{
			"BuckHashSys", float64(currentStats.BuckHashSys), "gauge",
		},
		{
			"Frees", float64(currentStats.Frees), "gauge",
		},
		{
			"GCCPUFraction", currentStats.GCCPUFraction, "gauge",
		},
		{
			"GCSys", float64(currentStats.GCSys), "gauge",
		},
		{
			"HeapAlloc", float64(currentStats.HeapAlloc), "gauge",
		},
		{
			"HeapIdle", float64(currentStats.HeapIdle), "gauge",
		},
		{
			"HeapInuse", float64(currentStats.HeapInuse), "gauge",
		},
		{
			"HeapObjects", float64(currentStats.HeapObjects), "gauge",
		},
		{
			"HeapReleased", float64(currentStats.HeapReleased), "gauge",
		},
		{
			"HeapSys", float64(currentStats.HeapSys), "gauge",
		},
		{
			"LastGC", float64(currentStats.LastGC), "gauge",
		},
		{
			"Lookups", float64(currentStats.Lookups), "gauge",
		},
		{
			"MCacheInuse", float64(currentStats.MCacheInuse), "gauge",
		},
		{
			"MCacheSys", float64(currentStats.MCacheSys), "gauge",
		},
		{
			"MSpanInuse", float64(currentStats.MSpanInuse), "gauge",
		},
		{
			"MSpanSys", float64(currentStats.MSpanSys), "gauge",
		},
		{
			"Mallocs", float64(currentStats.Mallocs), "gauge",
		},
		{
			"NextGC", float64(currentStats.NextGC), "gauge",
		},
		{
			"NumForcedGC", float64(currentStats.NumForcedGC), "gauge",
		},
		{
			"NumGC", float64(currentStats.NumGC), "gauge",
		},
		{
			"OtherSys", float64(currentStats.OtherSys), "gauge",
		},
		{
			"PauseTotalNs", float64(currentStats.PauseTotalNs), "gauge",
		},
		{
			"StackInuse", float64(currentStats.StackInuse), "gauge",
		},
		{
			"StackSys", float64(currentStats.StackSys), "gauge",
		},
		{
			"Sys", float64(currentStats.Sys), "gauge",
		},
		{
			"TotalAlloc", float64(currentStats.TotalAlloc), "gauge",
		},
		{
			"RandomValue", rand.Float64(), "gauge",
		},
	}

	pollCount += int64(len(stats) - 1)

	return stats, pollCount
}

func sendMemStats(stats []gStat, pollCount int64) {
	client := &http.Client{}
	for _, stat := range stats {
		sendRequest(client, stat.Type, stat.Name, fmt.Sprintf("%v", stat.Value))
	}

	sendRequest(client, "counter", "PollCount", fmt.Sprintf("%v", pollCount))
}

func sendRequest(c *http.Client, t string, n string, v string) {
	endpoint := "http://127.0.0.1:8080"
	url := endpoint + "/update/" + t + "/" + n + "/" + v

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

func Run() {
	var pollInterval time.Duration = 2
	var reportInterval time.Duration = 10
	var shutdownInterval time.Duration = 60

	collectTicker := time.NewTicker(pollInterval * time.Second)
	sendTicker := time.NewTicker(reportInterval * time.Second)
	shutdownTicker := time.NewTicker(shutdownInterval * time.Minute)

	defer collectTicker.Stop()
	defer sendTicker.Stop()
	defer shutdownTicker.Stop()

	var stats []gStat
	var pollCount int64 = 0
	ch := make(chan bool, 1)
	ch <- true
	for {
		select {
		case <-collectTicker.C:
			<-ch
			stats, pollCount = collectMemStats(pollCount)
			ch <- true

		case <-sendTicker.C:
			<-ch
			sendMemStats(stats, pollCount)
			ch <- true
		case <-shutdownTicker.C:
			return
		}
	}
}
