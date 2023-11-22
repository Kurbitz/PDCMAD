package main

import (
	"log"
	"os"
	"pdc-mad/metrics"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	//resultChannel := make(chan []*metrics.Metric)
	folderPath := "../../../dataset/"
	files, err := os.ReadDir(folderPath)
	if err != nil {
		log.Fatal(err)
	}
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
}
