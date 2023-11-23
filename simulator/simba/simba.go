package main

import (
	"os"
	"path/filepath"
	simba "pdc-mad/simba/internal"

	"sync"

	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	if _, tokenExists := os.LookupEnv("INFLUXDB_TOKEN"); !tokenExists {
		panic("Missing variable in .env file: INFLUXDB_TOKEN")
	}
	if _, ipExists := os.LookupEnv("INFLUXDB_IP"); !ipExists {
		panic("Missing variable in .env file: INFLUXDB_IP")
	}
	if _, portExists := os.LookupEnv("INFLUXDB_PORT"); !portExists {
		panic("Missing variable in .env file: INFLUXDB_PORT")
	}
	if _, pathExists := os.LookupEnv("DATASET_PATH"); !pathExists {
		panic("Missing variable in .env file: DATASET_PATH")
	}
}

func main() {
	var wg sync.WaitGroup
	// FIXME: Read this from a parameter or .env file
	datasetPath := os.Getenv("DATASET_PATH")
	files, err := os.ReadDir(datasetPath)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".csv" {
			continue
		}
		filePath := datasetPath + file.Name()
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
