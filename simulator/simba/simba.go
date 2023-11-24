package main

import (
	"os"
	"path/filepath"
	simba "pdc-mad/simba/internal"

	"sync"
)

func main() {
	var wg sync.WaitGroup
	// FIXME: Read this from a parameter or .env file
	folderPath := "../../../dataset/"
	files, err := os.ReadDir(folderPath)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".csv" {
			continue
		}
		filePath := folderPath + file.Name()
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()
			// Remove the .csv from the file name
			// FIXME: Use better ID
			id := filepath.Base(filePath)[:len(filepath.Base(filePath))-len(filepath.Ext(filePath))]
			metric, _ := simba.ReadFromFile(filePath, id)
			simba.WriteMetric(*metric)
		}(filePath)

	}

	wg.Wait()
}
