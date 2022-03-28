package agent

import (
	"runtime"
	"fmt"
	"time"
	"net/http"
)

type gauge float64

type counter int64

type memStat struct {
	Name string
	Value string
	Type string
}

var stats []memStat

var pollInterval time.Duration = 2

var reportInterval time.Duration = 10

var endpoint string = "http://127.0.0.1:8080"

func collectMemStats() {
	var currentStats runtime.MemStats
	runtime.ReadMemStats(&currentStats)

	stats = []memStat{
		memStat{
			"Alloc", fmt.Sprintf("%v", currentStats.Alloc), "gauge", 
		},
		memStat{
			"BuckHashSys", fmt.Sprintf("%v", currentStats.BuckHashSys), "gauge",
		},
		memStat{
			"Frees", fmt.Sprintf("%v", currentStats.Frees), "gauge",
		},
		memStat{ 
			"GCCPUFraction", fmt.Sprintf("%v", currentStats.GCCPUFraction), "gauge",
		},
		memStat{
			"GCSys", fmt.Sprintf("%v", currentStats.GCSys), "gauge",
		},
		memStat{
			"HeapAlloc", fmt.Sprintf("%v", currentStats.HeapAlloc), "gauge",
		},
		memStat{
			"HeapIdle", fmt.Sprintf("%v", currentStats.HeapIdle), "gauge",
		},
		memStat{
			"HeapInuse", fmt.Sprintf("%v", currentStats.HeapInuse), "gauge",
		},
		memStat{
			"HeapObjects", fmt.Sprintf("%v", currentStats.HeapObjects), "gauge",
		},
		memStat{
			"HeapReleased", fmt.Sprintf("%v", currentStats.HeapReleased), "gauge",
		},
		memStat{
			"HeapSys", fmt.Sprintf("%v", currentStats.HeapSys), "gauge",
		},
		memStat{
			"LastGC", fmt.Sprintf("%v", currentStats.LastGC), "gauge",
		},
		memStat{
			"Lookups", fmt.Sprintf("%v", currentStats.Lookups), "gauge",
		},
		memStat{
			"MCacheInuse", fmt.Sprintf("%v", currentStats.MCacheInuse), "gauge",
		},
		memStat{
			"MCacheSys", fmt.Sprintf("%v", currentStats.MCacheSys), "gauge",
		},
		memStat{
			"MSpanInuse", fmt.Sprintf("%v", currentStats.MSpanInuse), "gauge",
		},
		memStat{
			"MSpanSys", fmt.Sprintf("%v", currentStats.MSpanSys), "gauge",
		},
		memStat{
			"Mallocs", fmt.Sprintf("%v", currentStats.Mallocs), "gauge",
		},
		memStat{
			"NextGC", fmt.Sprintf("%v", currentStats.NextGC), "gauge",
		},
		memStat{
			"NumForcedGC", fmt.Sprintf("%v", currentStats.NumForcedGC), "gauge",
		},
		memStat{
			"NumGC", fmt.Sprintf("%v", currentStats.NumGC), "gauge",
		},
		memStat{
			"OtherSys", fmt.Sprintf("%v", currentStats.OtherSys), "gauge",
		},
		memStat{
			"PauseTotalNs", fmt.Sprintf("%v", currentStats.PauseTotalNs), "gauge",
		},
		memStat{
			"StackInuse", fmt.Sprintf("%v", currentStats.StackInuse), "gauge",
		},
		memStat{
			"StackSys", fmt.Sprintf("%v", currentStats.StackSys), "gauge",
		},
		memStat{
			"Sys", fmt.Sprintf("%v", currentStats.Sys), "gauge",
		},
		memStat{
			"TotalAlloc", fmt.Sprintf("%v", currentStats.TotalAlloc), "gauge",
		},
	}
}

func sendMemStats() {
	client := &http.Client{}
	for _, stat := range stats{
		url := endpoint + "/update/" + stat.Type + "/" + stat.Name + "/" + stat.Value
		
		request, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			fmt.Println(err)
		}
		request.Header.Set("Content-Type", "text/plain")

		response, err := client.Do(request)
		if err != nil {
			fmt.Println(err)
		}
		defer response.Body.Close()
	}
}

func RunClient() {
	collect_ticker := time.NewTicker(pollInterval * time.Second)
	send_ticker := time.NewTicker(reportInterval * time.Second)
	
	for {
		select {
		case <- collect_ticker.C:
			collectMemStats()
			// fmt.Println(stats)

		case <- send_ticker.C:
			sendMemStats()
		}
	}

}