package main

import (
	"logmonitor-homework/display"
	"logmonitor-homework/monitoring"

	"flag"
)

var (
	//collectionInterval = flag.Int("collection-inteval", 10, "Collection Interval in Seconds")
	avgThreshold = flag.Float64("avg-threshold", 10, "Total traffic threshold for alerting (hits/s)")
	filename     = flag.String("log-filename", "/tmp/access.log", "w3c-formatted HTTP access log to be consumed actively")
)

func main() {

	flag.Parse()

	cc := make(chan monitoring.Metrics)
	ca := make(chan string)
	//cs := make(chan float64)

	monitoring := &monitoring.HttpLogMonitoring{AlertInterval: 10,
		CollectionInterval: 10,
		AlertThreshold:     *avgThreshold,
		CollectionChannel:  cc,
		AlertChannel:       ca,
		LogFilename:        *filename,
	}

	go monitoring.Monitor()

	dashboard := &display.Dashboard{AlertInterval: 10,
		CollectionInterval: 10,
		AlertThreshold:     *avgThreshold,
		CollectionChannel:  cc,
		AlertChannel:       ca}

	dashboard.DisplayStart()

}
