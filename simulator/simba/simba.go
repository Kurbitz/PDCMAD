package main

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

type FillFlags struct {
	dbtoken  string
	dbip     string
	dbport   string
	duration time.Duration
	startat  time.Duration
	gap      time.Duration
}

type StreamFlags struct {
	dbtoken        string
	dbip           string
	dbport         string
	duration       time.Duration
	startat        time.Duration
	timeMultiplier int
	append         bool
}

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
						Flags: append(simulateFlags, &cli.StringFlag{
							Name:  "gap",
							Usage: "Gap to now",
							Value: "",
						}),
					},
					{
						Name:      "stream",
						Usage:     "stream data from file(s) in real time to the database",
						ArgsUsage: "<file1> <file2> ...",
						Action:    stream,
						Flags: append(simulateFlags, &cli.IntFlag{
							Name:  "timemultiplier",
							Value: 1,
						}, &cli.BoolFlag{
							Name:  "append",
							Value: false,
						}),
					},
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

func ValidateFile(file string) error {
	// Validate the file exists
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", file)
	}
	// Validate the file is not a directory
	if info, err := os.Stat(file); err == nil && info.IsDir() {
		return fmt.Errorf("file %s is a directory", file)
	}
	// Validate the file is a .csv files
	if filepath.Ext(file) != ".csv" {
		return fmt.Errorf("file %s is not a .csv file", file)
	}
	// Validate the file is not empty
	if info, err := os.Stat(file); err == nil && info.Size() == 0 {
		return fmt.Errorf("file %s is empty", file)
	}
	return nil
}

func ParseFillFlags(ctx *cli.Context) (*FillFlags, error) {
	if ctx.String("dbtoken") == "" {
		return nil, fmt.Errorf("missing InfluxDB token. See -h for help")
	}
	duration, err := ParseDurationString(ctx.String("duration"))
	if err != nil {
		return nil, err
	}
	startAt, err := ParseDurationString(ctx.String("startat"))
	if err != nil {
		return nil, err
	}
	gap, err := ParseDurationString(ctx.String("gap"))
	if err != nil {
		return nil, err
	}

	return &FillFlags{
		dbtoken:  ctx.String("dbtoken"),
		dbip:     ctx.String("dbip"),
		dbport:   ctx.String("dbport"),
		duration: duration,
		startat:  startAt,
		gap:      gap,
	}, nil
}

func ParseStreamFlags(ctx *cli.Context) (*StreamFlags, error) {
	if ctx.String("dbtoken") == "" {
		return nil, fmt.Errorf("missing InfluxDB token. See -h for help")
	}
	duration, err := ParseDurationString(ctx.String("duration"))
	if err != nil {
		return nil, err
	}
	startAt, err := ParseDurationString(ctx.String("startat"))
	if err != nil {
		return nil, err
	}

	return &StreamFlags{
		dbtoken:        ctx.String("dbtoken"),
		dbip:           ctx.String("dbip"),
		dbport:         ctx.String("dbport"),
		duration:       duration,
		startat:        startAt,
		timeMultiplier: ctx.Int("timemultiplier"),
		append:         ctx.Bool("append"),
	}, nil
}

// The stream command reads a single file and sends them to InfluxDB in real time
// The file is passed as an argument to the application (simba stream file.csv)
func stream(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.Exit("Missing file", 1)
	}

	// Validate the file
	file := ctx.Args().Slice()[0]
	if err := ValidateFile(file); err != nil {
		return cli.Exit(err, 1)
	}

	// Parse the flags
	flags, err := ParseStreamFlags(ctx)
	if err != nil {
		return cli.Exit(err, 1)
	}

	return nil
}

// The fill command reads the files and sends them to InfluxDB
// The files are passed as arguments to the application (simba fill file1.csv file2.csv etc.)
func fill(ctx *cli.Context) error {
	// Validate ctx.Args contains at least one file
	if ctx.NArg() == 0 {
		return cli.Exit("Missing file(s)", 1)
	}
	for _, file := range ctx.Args().Slice() {
		// Validate the file
		if err := ValidateFile(file); err != nil {
			return cli.Exit(err, 1)
		}
	}

	// Parse the flags
	flags, err := ParseFillFlags(ctx)
	if err != nil {
		return cli.Exit(err, 1)
	}

	var influxDBApi = simba.NewInfluxDBApi(flags.dbtoken, flags.dbip, flags.dbport)

	var wg sync.WaitGroup
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
			metric.SliceBetween(flags.startat, flags.duration)
			influxDBApi.WriteMetrics(*metric, flags.gap)
		}(file)

	}
	wg.Wait()
	return nil
}
