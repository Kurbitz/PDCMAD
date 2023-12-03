package main

import (
	"log"
	"os"
	"path/filepath"
	simba "pdc-mad/simba/internal"

	"sync"

	"github.com/joho/godotenv"
)

func checkEnvVar(variableName string) {
	if _, exists := os.LookupEnv(variableName); !exists {
		log.Fatalf("Missing environment variable: %s", variableName)
	}
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	checkEnvVar("INFLUXDB_TOKEN")
	checkEnvVar("INFLUXDB_IP")
	checkEnvVar("INFLUXDB_PORT")
	checkEnvVar("DATASET_PATH")
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
