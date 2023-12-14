package main

// DBTOKEN: pJtsYfVHcRpuHOYcVRa7PPDy9aB1bDXTb23TJi4_vMtagT2p-s5ez9ENEAAIJ970KVqGDck_ttoEV_6W7l-Gsw==

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	simba "pdc-mad/simba/internal"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/urfave/cli/v2"
)

func main() {
	// Define the CLI Commands and flags

	simulateFlags := []cli.Flag{
		&cli.StringFlag{
			Name:    "dbtoken",
			EnvVars: []string{"INFLUXDB_TOKEN"},
			Usage:   "InfluxDB token",
			Value:   "",
		},
		&cli.StringFlag{
			Name:    "dbip",
			EnvVars: []string{"INFLUXDB_IP"},
			Usage:   "InfluxDB IP",
			Value:   "localhost",
		},
		&cli.StringFlag{
			Name:    "dbport",
			EnvVars: []string{"INFLUXDB_PORT"},
			Usage:   "InfluxDB port",
			Value:   "8086",
		},
		&cli.StringFlag{
			Name:  "duration",
			Usage: "duration",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "startat",
			Usage: "Starting line in file",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "gap",
			Usage: "Gap to now",
			Value: "",
		},
	}

	cleanFlags := []cli.Flag{
		&cli.StringFlag{
			Name:    "dbtoken",
			EnvVars: []string{"INFLUXDB_TOKEN"},
			Usage:   "InfluxDB token",
			Value:   "",
		},
		&cli.StringFlag{
			Name:    "dbip",
			EnvVars: []string{"INFLUXDB_IP"},
			Usage:   "InfluxDB IP",
			Value:   "localhost",
		},
		&cli.StringFlag{
			Name:    "dbport",
			EnvVars: []string{"INFLUXDB_PORT"},
			Usage:   "InfluxDB port",
			Value:   "8086",
		},
		&cli.StringFlag{
			Name:  "startat",
			Usage: "from where to delete relative to current time",
			Value: "0d",
		},
	}

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
				Name:  "simulate",
				Usage: "Simulate metrics from file(s)",
				Subcommands: []*cli.Command{
					{
						Name:      "fill",
						Usage:     "fill the database with data from file(s)",
						ArgsUsage: "<file1> <file2> ...",
						Action:    fill,
						Flags:     simulateFlags,
					},
					{
						Name:      "stream",
						Usage:     "stream data from file(s) in real time to the database",
						ArgsUsage: "<file1> <file2> ...",
						Action: func(ctx *cli.Context) error {
							fmt.Println("simulate stream")
							return nil
						},
						Flags: append(simulateFlags, &cli.IntFlag{
							Name:  "timemultiplier",
							Value: 1,
						}),
					},
				},
			},

			{
				Name:  "clean",
				Usage: "Clean the database",
				Flags: cleanFlags,
				Subcommands: []*cli.Command{
					{
						Name:      "bucket",
						Usage:     "clean all the data inside the specified bucket",
						ArgsUsage: "<bucket>",
						Action:    cleanBucket,
						Flags:     cleanFlags,
					},
					{
						Name:      "host",
						Usage:     "clean all the data from a host/system inside the specified bucket",
						ArgsUsage: "<host1> <host2> ...",
						Action:    cleanHost,
						Flags: append(cleanFlags, &cli.StringFlag{
							Name:  "bucket",
							Usage: "Bucket from where to delete",
							Value: "metrics",
						}),
					},
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
		// FIXME: Make these persistent flags instead. Only supported in v3 alpha right now though.
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Enable verbose output",
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}

}

// ParseDurationString parses a string like 1d, 1h or 1m and returns a time.Duration
// Supports days, hours and minutes (d, h, m)
// Does not return an error if the string is empty, instead it returns 0. This is to allow for default values.
func ParseDurationString(ds string) (time.Duration, error) {
	if ds == "" {
		return 0, nil
	}
	r := regexp.MustCompile("^([0-9]+)(d|h|m)$")

	match := r.FindStringSubmatch(ds)
	if len(match) == 0 {
		return 0, fmt.Errorf("invalid time string: %s", ds)
	}

	amount, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, fmt.Errorf("invalid time string: %s", ds)
	}

	switch match[2] {
	case "d":
		return ((time.Hour * 24) * time.Duration(amount)), nil
	case "h":
		return ((time.Hour) * time.Duration(amount)), nil
	case "m":
		return (time.Minute * time.Duration(amount)), nil

	}

	return 0, fmt.Errorf("invalid time string: %s", ds)
}

// The fill command reads the files and sends them to InfluxDB
// The files are passed as arguments to the application (simba fill file1.csv file2.csv etc.)
func fill(ctx *cli.Context) error {
	// Validate ctx.Args contains at least one file
	if ctx.NArg() == 0 {
		return cli.Exit("Missing file(s)", 1)
	}
	if ctx.String("dbtoken") == "" {
		return cli.Exit("Missing InfluxDB token. See -h for help", 1)
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

	// Parse the flags
	duration, err := ParseDurationString(ctx.String("duration"))
	if err != nil {
		return cli.Exit(err, 1)
	}
	startAt, err := ParseDurationString(ctx.String("startat"))
	if err != nil {
		return cli.Exit(err, 1)
	}
	gap, err := ParseDurationString(ctx.String("gap"))
	if err != nil {
		return cli.Exit(err, 1)
	}

	var influxDBApi = simba.NewInfluxDBApi(ctx.String("dbtoken"), ctx.String("dbip"), ctx.String("dbport"))

	for _, file := range ctx.Args().Slice() {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()
			// Remove the .csv from the file name
			// FIXME: Use better ID
			id := filepath.Base(filePath)[:len(filepath.Base(filePath))-len(filepath.Ext(filePath))]
			metric, _ := simba.ReadFromFile(filePath, id)

			// Slice the metric between startAt and duration
			// If the parameters are 0, it will return all metrics, so we don't need to check for that
			metric.SliceBetween(startAt, duration)
			influxDBApi.WriteMetrics(*metric, gap)
		}(file)

	}
	wg.Wait()
	return nil
}

// The cleanBucket subcommand removes all the data inside the specified bucket
func cleanBucket(ctx *cli.Context) error {
	// Validate ctx.Args contains a bucket name
	if ctx.NArg() == 0 {
		return cli.Exit("Missing bucket name", 1)
	}

	if ctx.String("dbtoken") == "" {
		return cli.Exit("Missing InfluxDB token. See -h for help", 1)
	}

	bucket := ctx.Args().First()

	influxDBApi := simba.NewInfluxDBApi(ctx.String("dbtoken"), ctx.String("dbip"), ctx.String("dbport"))

	return influxDBApi.DeleteBucket(bucket)
}

// The cleanHost subcommand removes all the data from the desired
// hosts/systems inside the specified bucket
func cleanHost(ctx *cli.Context) error {
	// Validate ctx.Args contains at least a host/system name
	if ctx.NArg() == 0 {
		return cli.Exit("Missing host names", 1)
	}

	if ctx.String("dbtoken") == "" {
		return cli.Exit("Missing InfluxDB token. See -h for help", 1)
	}

	// Validate that a bucket has been specified
	bucket := ctx.String("bucket")
	if bucket == "" {
		return cli.Exit("Bucket not specified. See -h for help", 1)
	}

	influxDBApi := simba.NewInfluxDBApi(ctx.String("dbtoken"), ctx.String("dbip"), ctx.String("dbport"))

	var wg sync.WaitGroup

	for _, host := range ctx.Args().Slice() {
		wg.Add(1)
		go func(hostName string, bucketName string) {
			defer wg.Done()
			influxDBApi.DeleteHost(bucket, hostName)
		}(host, bucket)
	}
	wg.Wait()

	return nil
}
