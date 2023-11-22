package main

import (
	"pdc-mad/metrics"

	"sync"
)

func main() {
	var wg sync.WaitGroup
	//resultChannel := make(chan []*metrics.Metric)
	fileName := "../../../dataset/"
	fileNumber := []string{"system-1.csv", "system-2.csv"}
	for _, fileNum := range fileNumber {
		wg.Add(1)
		go func() {
			metrics.ReadFromFile((fileName + fileNum))
			wg.Done()
		}()
	}

	wg.Wait()
}
