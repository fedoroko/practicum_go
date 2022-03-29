package storage

import (
	"strconv"
	// "fmt"
	// "errors"
)

type repositories interface {
	update()
}

type gauge float64

type counter int64

type MemStats struct {
	Alloc gauge
	BuckHashSys gauge
	Frees gauge
	GCCPUFraction gauge
	GCSys gauge
	HeapAlloc gauge
	HeapIdle gauge
	HeapInuse gauge
	HeapObjects gauge
	HeapReleased gauge
	HeapSys gauge
	LastGC gauge
	Lookups gauge
	MCacheInuse gauge
	MCacheSys gauge
	MSpanInuse gauge
	MSpanSys gauge
	Mallocs gauge
	NextGC gauge
	NumForcedGC gauge
	NumGC gauge
	OtherSys gauge
	PauseTotalNs gauge
	StackInuse gauge
	StackSys gauge
	Sys gauge
	TotalAlloc gauge
	RandomValue gauge
	PollCount counter
}

var dummyStorage MemStats

func (m MemStats) update(t string, n string, v string) error {
	
	if t == "gauge" {
		val, err := strconv.ParseFloat(v, 64)	
		if err != nil {
			return err
		}

		gval := gauge(val)
		switch n {
		case "Alloc":
			dummyStorage.Alloc = gval
		case "BuckHashSys":
			dummyStorage.BuckHashSys = gval
		case "Frees":
			dummyStorage.Frees = gval
		case "GCCPUFraction":
			dummyStorage.GCCPUFraction = gval
		case "GCSys":
			dummyStorage.GCSys = gval
		case "HeapAlloc":
			dummyStorage.HeapAlloc = gval
		case "HeapIdle":
			dummyStorage.HeapIdle = gval
		case "HeapInuse":
			dummyStorage.HeapInuse = gval
		case "HeapObjects":
			dummyStorage.HeapObjects = gval
		case "HeapReleased":
			dummyStorage.HeapReleased = gval
		case "HeapSys":
			dummyStorage.HeapSys = gval
		case "LastGC":
			dummyStorage.LastGC = gval
		case "Lookups":
			dummyStorage.Lookups = gval
		case "MCacheInuse":
			dummyStorage.MCacheInuse = gval
		case "MCacheSys":
			dummyStorage.MCacheSys = gval
		case "MSpanInuse":
			dummyStorage.MSpanInuse = gval
		case "MSpanSys":
			dummyStorage.MSpanSys = gval
		case "Mallocs":
			dummyStorage.Mallocs = gval
		case "NextGC":
			dummyStorage.NextGC = gval
		case "NumForcedGC":
			dummyStorage.NumForcedGC = gval
		case "NumGC":
			dummyStorage.NumGC = gval
		case "OtherSys":
			dummyStorage.OtherSys = gval
		case "PauseTotalNs":
			dummyStorage.PauseTotalNs = gval
		case "StackInuse":
			dummyStorage.StackInuse = gval
		case "StackSys":
			dummyStorage.StackSys = gval
		case "Sys":
			dummyStorage.Sys = gval
		case "TotalAlloc":
			dummyStorage.TotalAlloc = gval
		case "RandomValue":
			dummyStorage.RandomValue = gval
		default:
			return nil
		}
	} else {
		val, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}

		dummyStorage.PollCount = dummyStorage.PollCount + counter(val)
	}

	return nil
	

	// fmt.Println(dummyStorage)
}

func Store(t string, n string, v string) error {
	return dummyStorage.update(t, n, v)
}
