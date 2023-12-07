package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	simba "pdc-mad/simba/internal"

	"sync"

	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

// Check if the environment variable exists
// If not, exit the program with an error
func checkEnvVar(variableName string) {
	if _, exists := os.LookupEnv(variableName); !exists {
		log.Fatalf("Missing environment variable: %s", variableName)
	}
}

func init() {
	// Load the .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Check if all the environment variables are set
	checkEnvVar("INFLUXDB_TOKEN")
	checkEnvVar("INFLUXDB_IP")
	checkEnvVar("INFLUXDB_PORT")
	checkEnvVar("DATASET_PATH")
}

func main() {
	// Define the CLI Commands and flags
	app := &cli.App{
		Name:  "simba",
		Usage: "Simulate metrics etc.",
		Authors: []*cli.Author{
			{
				Name: "PDC-MAD",
			},
		},
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:   "populate",
				Usage:  "Batch import data",
				Action: populate,
			},
			{
				Name:  "stream",
				Usage: "simulate in real time",
				Action: func(ctx *cli.Context) error {
					fmt.Println("real time")
					return nil
				},
			},
			{
				Name:  "clean",
				Usage: "Clean the database",
				Action: func(ctx *cli.Context) error {
					fmt.Println("clean")
					return nil
				},
			},
			{
				Name:  "trigger",
				Usage: "Trigger anomaly detection",
				Action: func(ctx *cli.Context) error {
					fmt.Println("trigger")
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}

}

// populate the database with the data from the dataset
// FIXME: Add a parameters for dataset path, duration, etc.
func populate(ctx *cli.Context) error {
	var wg sync.WaitGroup
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
	return nil
}
