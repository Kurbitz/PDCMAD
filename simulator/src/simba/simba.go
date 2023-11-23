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
	//resultChannel := make(chan []*metrics.Metric)
	folderPath := "../../../dataset/"
	files, err := os.ReadDir(folderPath)
	if err != nil {
		log.Fatal(err)
	}
	startTime := time.Now()
	for _, file := range files {
		if file.IsDir() || file.Name() == ".gitkeep" {
			continue
		}
		filePath := folderPath + file.Name()
		wg.Add(1)
		go func(fp string) {
			metrics.ReadFromFile(fp)
			wg.Done()
		}(filePath)
	}

	wg.Wait()
	elapsedTime := time.Since(startTime)
	fmt.Printf("Time: %s", elapsedTime)
}
