package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	simba "pdc-mad/simba/internal"
	"sync"

	"github.com/urfave/cli/v2"
)

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
				Name:      "populate",
				Usage:     "Batch import data",
				Action:    populate,
				ArgsUsage: "<file1> <file2> ...",
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

// FIXME: Add a parameters for, duration, etc.
// The populate command reads the files and sends them to InfluxDB
// The files are passed as arguments to the application (simba populate file1.csv file2.csv etc.)
func populate(ctx *cli.Context) error {
	// Validate ctx.Args contains at least one file
	if ctx.NArg() == 0 {
		return cli.Exit("Missing file(s)", 1)
	}
	for _, file := range ctx.Args().Slice() {
		// Validate the files exist
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return cli.Exit(fmt.Sprintf("File %s does not exist", file), 1)
		}
		// Validate the files are not directories
		if info, err := os.Stat(file); err == nil && info.IsDir() {
			return cli.Exit(fmt.Sprintf("File %s is a directory", file), 1)
		}
		// Validate the files are .csv files
		if filepath.Ext(file) != ".csv" {
			return cli.Exit(fmt.Sprintf("File %s is not a .csv file", file), 1)
		}
		// Validate the files are not empty
		if info, err := os.Stat(file); err == nil && info.Size() == 0 {
			return cli.Exit(fmt.Sprintf("File %s is empty", file), 1)
		}

	}

	var wg sync.WaitGroup

	for _, file := range ctx.Args().Slice() {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()
			// Remove the .csv from the file name
			// FIXME: Use better ID
			id := filepath.Base(filePath)[:len(filepath.Base(filePath))-len(filepath.Ext(filePath))]
			metric, _ := simba.ReadFromFile(filePath, id)
			simba.WriteMetric(*metric)
		}(file)

	}
	wg.Wait()
	return nil
}
