package main

import (
	"fmt"
	"log"
	"os"
	"pdc-mad/metrics"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	const numWorkers = 19
	folderPath := "../../../dataset/"
	files, err := os.ReadDir(folderPath)
	var systemMetrics = []*metrics.SystemMetric{}
	if err != nil {
		log.Fatal(err)
	}
	filePathChan := make(chan string, len(files))
	startTime := time.Now()
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".csv" {
			continue
		}
		filePath := folderPath + file.Name()
		filePathChan <- filePath
		wg.Add(1)
		go func(filePathChan <-chan string) {
			defer wg.Done()
			for filepath := range filePathChan {
				metric, _ := metrics.ReadFromFile(filepath)
				systemMetrics = append(systemMetrics, metric)
			}
		}(filePathChan)

	}
	close(filePathChan)
	wg.Wait()
	for _, sm := range systemMetrics {
		println(sm.Id)
		println(sm.Metr[10].Server_Up)

	}
	elapsedTime := time.Since(startTime)
	fmt.Printf("Time: %s", elapsedTime)

}
