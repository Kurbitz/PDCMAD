package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
	"golang.org/x/exp/maps"
)

// Eeach command has its own struct containing the flags passed to it.
// The flags are parsed in the Parse*Flags functions.
// This is done to avoid having to pass a lot of arguments to the functions and allow us to separate parsing from the actual logic.

// To add a new command, add a new struct here and add a new command to the App variable. Then add the logic to the corresponding command in commands.go

// DBInfo is a struct containing the information needed to connect to the database
// It is used by several commands and is defined here to avoid duplication.
type DBInfo struct {
	Token       string // InfluxDB token
	Host        string // InfluxDB hostname
	Port        string // InfluxDB port
	Org         string // InfluxDB organization
	Bucket      string // InfluxDB bucket
	Measurement string // InfluxDB measurement
}

// FillArgs is a struct containing the flags passed to the fill command
type FillArgs struct {
	DBArgs   DBInfo        // DBInfo struct containing the database information
	Duration time.Duration // Duration of the simulation
	StartAt  time.Duration // How far into the file to start the simulation
	Gap      time.Duration // How much time to leave between the last metric and now for future simulations
	Anomaly  string        // Which anomaly to use (see error_injection.go)
	Files    []string      // The CSV files of the metrics to simulate
}

// StreamArgs is a struct containing the flags passed to the stream command
type StreamArgs struct {
	DBArgs         DBInfo        // DBInfo struct containing the database information
	Duration       time.Duration // Duration of the simulation
	StartAt        time.Duration // How far into the file to start the simulation
	TimeMultiplier int           // How much to speed up the simulation
	Append         bool          // Whether to append to the latest metric or not
	Anomaly        string        // Which anomaly to use (see error_injection.go)
	File           string        // The CSV file of the metrics to simulate
}

// CleanArgs is a struct containing the flags passed to the clean command
type CleanArgs struct {
	DBArgs   DBInfo        // DBInfo struct containing the database information
	All      bool          // Whether to delete all the metrics or not
	Duration time.Duration // How far into the past to delete metrics
	Hosts    []string      // The hosts to delete metrics from
}

// Common flags for the fill and stream commands
// V2 of urfave/cli does not support shared flags so to avoid duplication we define them here and pass them to the commands
// FIXME: Use shared flags when (if) they are implemented in V3
var simulateFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  "duration",
		Usage: "How long the simulation should run. Duration string.",
		Value: "",
		Aliases: []string{
			"d",
		},
	},
	&cli.StringFlag{
		Name:  "start-at",
		Usage: "How far into the file to start the simulation. Duration string.",
		Value: "",
		Aliases: []string{
			"s",
		},
	},
	&cli.StringFlag{
		Name: "anomaly",
		// This adds the available anomalies to the help text in kind of a hacky way
		Usage: "Select which type of anomaly to use. Available: " + strings.Join(maps.Keys(AnomalyMap), ", "),
		Value: "",
		Aliases: []string{
			"a",
		},
	},
	&cli.StringFlag{
		Name:     "db-token",
		EnvVars:  []string{"INFLUXDB_TOKEN"},
		Usage:    "InfluxDB token",
		Value:    "",
		Category: "Database",
		Aliases: []string{
			"T",
		},
	},
	&cli.StringFlag{
		Name:     "db-host",
		EnvVars:  []string{"INFLUXDB_HOST"},
		Usage:    "InfluxDB hostname",
		Value:    "localhost",
		Category: "Database",
		Aliases: []string{
			"H",
		},
	},
	&cli.StringFlag{
		Name:     "db-port",
		EnvVars:  []string{"INFLUXDB_PORT"},
		Usage:    "InfluxDB port",
		Value:    "8086",
		Category: "Database",
		Aliases: []string{
			"P",
		},
	},
	&cli.StringFlag{
		Name:     "db-org",
		Usage:    "InfluxDB organization",
		EnvVars:  []string{"INFLUXDB_ORG"},
		Value:    "pdc-mad",
		Category: "Database",
		Aliases: []string{
			"O",
		},
	},
	&cli.StringFlag{
		Name:     "db-bucket",
		Usage:    "InfluxDB bucket",
		EnvVars:  []string{"INFLUXDB_BUCKET"},
		Value:    "pdc-mad",
		Category: "Database",
		Aliases: []string{
			"B",
		},
	},
}

// App is the main application
// All commands and flags are defined here
// See urfave/cli documentation for more information
var App = &cli.App{
	Name:  "simba",
	Usage: "Simulate systems producing metrics and feed them to InfluxDB",
	Description: "Simba is a tool for simulating systems producing metrics and feeding them to InfluxDB.\n" +
		"Metrics are read from CSV files and inserted into InfluxDB.\n" +
		"Anomalies can be injected into the metrics to simulate errors in the system.\n" +
		"Simba can also be used to clean metrics from the database.\n" +
		"See the documentation for more information.",
	Suggest: true,
	Authors: []*cli.Author{
		{
			Name: "The Performance Data Collection, Monitoring and Anomaly Detection Team (PDC-MAD):",
		},
		{
			Name: "Oscar Einarsson",
		},
		{
			Name: "Fredrik Nygårds",
		},
		{
			Name: "Gustav Kånåhols",
		},
		{
			Name: "Fernando Revillas",
		},
		{
			Name: "Adam Segerström",
		},
		{
			Name: "Tahira Nishat",
		},
	},
	EnableBashCompletion: true,
	// Commands are defined here
	// Add a new command by adding a new Command struct to the slice
	Commands: []*cli.Command{
		{
			Name:      "fill",
			Usage:     "Fill the database with data from file(s). Files must be in the specified CSV format",
			ArgsUsage: "<file1> <file2> ...",
			Description: "A duration string is a string like 1d, 1h or 1m.\n" +
				"Supported units are days (d), hours (h) and minutes (m).\n" +
				"Examples: 1d, 2h, 30m\n" +
				"Composite durations are not supported (e.g. 1d2h30m)",
			Action: func(ctx *cli.Context) error {
				// Parse the flags
				flags, err := ParseFillFlags(ctx)
				if err != nil {
					return cli.Exit(err, 1)
				}
				// Execute the logic
				if err := Fill(*flags); err != nil {
					return cli.Exit(err, 1)
				}
				return nil
			},
			// Append the flags to the common simulation flags
			Flags: append(simulateFlags, &cli.StringFlag{
				Name:  "gap",
				Usage: "The time to leave between the last metric and now for future simulations.",
				Value: "",
				Aliases: []string{
					"g",
				},
			}),
		},
		{
			Name:      "stream",
			Usage:     "stream data from file(s) in real time to the database",
			ArgsUsage: "<file1> <file2> ...",
			Description: "A duration string is a string like 1d, 1h or 1m.\n" +
				"Supported units are days (d), hours (h) and minutes (m).\n" +
				"Examples: 1d, 2h, 30m\n" +
				"Composite durations are not supported (e.g. 1d2h30m)",
			Action: func(ctx *cli.Context) error {
				// Parse the flags
				flags, err := ParseStreamFlags(ctx)
				if err != nil {
					return cli.Exit(err, 1)
				}
				// Execute the logic
				if err := Stream(*flags); err != nil {
					return cli.Exit(err, 1)
				}
				return nil
			},
			// Append the flags to the common simulation flags
			Flags: append(simulateFlags, &cli.IntFlag{
				Name:  "time-multiplier",
				Usage: "Increase insertion speed by a factor of n. Must be >= 1. Extreme values may cause problems, user beware.",
				Value: 1,
				Aliases: []string{
					"t",
				},
			}, &cli.BoolFlag{
				Name:  "append",
				Usage: "Append to the latest metric with the same ID. If not set, the metric will be inserted using the current (wall) time.",
				Value: false,
			}),
		},
		{
			Name:      "clean",
			Usage:     "Clean the database of data from host(s) or all hosts.",
			ArgsUsage: "<host1> <host2> ...",
			Action: func(ctx *cli.Context) error {
				//Parse the flags
				flags, err := ParseCleanFlags(ctx)
				if err != nil {
					return cli.Exit(err, 1)
				}
				// Execute the logic
				if err := Clean(*flags); err != nil {
					return cli.Exit(err, 1)
				}
				return nil
			},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "duration",
					Usage: "How far into the past to delete metrics. Duration string.",
					Value: "",
					Aliases: []string{
						"d",
					},
				},
				&cli.BoolFlag{
					Name:  "all",
					Usage: "Delete metrics from all the hosts of the bucket",
					Value: false,
				},
				&cli.StringFlag{
					Name:     "db-token",
					EnvVars:  []string{"INFLUXDB_TOKEN"},
					Usage:    "InfluxDB token",
					Value:    "",
					Category: "Database",
					Aliases: []string{
						"T",
					},
				},
				&cli.StringFlag{
					Name:     "db-host",
					EnvVars:  []string{"INFLUXDB_HOST"},
					Usage:    "InfluxDB IP",
					Value:    "localhost",
					Category: "Database",
					Aliases: []string{
						"H",
					},
				},
				&cli.StringFlag{
					Name:     "db-port",
					EnvVars:  []string{"INFLUXDB_PORT"},
					Usage:    "InfluxDB port",
					Value:    "8086",
					Category: "Database",
					Aliases: []string{
						"P",
					},
				},
				&cli.StringFlag{
					Name:     "db-org",
					Usage:    "InfluxDB organization",
					EnvVars:  []string{"INFLUXDB_ORG"},
					Value:    "pdc-mad",
					Category: "Database",
					Aliases: []string{
						"O",
					},
				},
				&cli.StringFlag{
					Name:     "db-bucket",
					Usage:    "InfluxDB bucket",
					EnvVars:  []string{"INFLUXDB_BUCKET"},
					Value:    "pdc-mad",
					Category: "Database",
					Aliases: []string{
						"B",
					},
				},
				&cli.StringFlag{
					Name:     "db-measurement",
					Usage:    "InfluxDB measurement. Use 'anomalies' to delete anomalies.",
					EnvVars:  []string{"INFLUXDB_MEASUREMENT"},
					Value:    "metrics",
					Category: "Database",
					Aliases: []string{
						"M",
					},
				},
			},
		},
	},
}

// ParseDurationString parses a string like 1d, 1h or 1m and returns a time.Duration
// Supports days, hours and minutes (d, h, m)
// Does not return an error if the string is empty, instead it returns 0. This is to allow for default values.
func ParseDurationString(ds string) (time.Duration, error) {
	if ds == "" {
		return 0, nil
	}
	// Regex to match the duration string
	// Captures the amount and the unit in different groups
	r := regexp.MustCompile("^([0-9]+)(d|h|m)$")

	// Find the matches
	match := r.FindStringSubmatch(ds)
	if len(match) == 0 {
		return 0, fmt.Errorf("invalid time string: %s", ds)
	}

	// Convert the amount to an int
	amount, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, fmt.Errorf("invalid time string: %s", ds)
	}

	// Return the duration based on the unit
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

// checkANomalyString checks if the anomalyString given exists in the AnomalyMap
// If it does not exist, it returns an error
func checkAnomalyString(anomalyString string) (string, error) {
	if anomalyString == "" {
		return anomalyString, nil
	}
	// Check if the anomaly exists in the AnomalyMap
	if _, exists := AnomalyMap[anomalyString]; exists {
		return anomalyString, nil
	}

	return anomalyString, fmt.Errorf("error injection %s is not implemented", anomalyString)
}

// ValidateFile validates that the filePath is a valid file
// Returns an error if the file does not exist, is a directory, is not a .csv file or is empty
func ValidateFile(filePath string) error {
	// Validate the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", filePath)
	}
	// Validate the file is not a directory
	if info, err := os.Stat(filePath); err == nil && info.IsDir() {
		return fmt.Errorf("file %s is a directory", filePath)
	}
	// Validate the file is a .csv files
	if filepath.Ext(filePath) != ".csv" {
		return fmt.Errorf("file %s is not a .csv file", filePath)
	}
	// Validate the file is not empty
	if info, err := os.Stat(filePath); err == nil && info.Size() == 0 {
		return fmt.Errorf("file %s is empty", filePath)
	}
	return nil
}

// GetIdFromFileName returns the ID of a metric from the file name
// The ID is the base file name without the extension
func GetIdFromFileName(file string) string {
	// Remove the file extension from the base file name
	return filepath.Base(file)[:len(filepath.Base(file))-len(filepath.Ext(file))]

}

// ParseFillFlags parses the flags passed to the fill command
// Returns a FillArgs struct containing the parsed flags
// Returns an error if the flags are invalid
func ParseFillFlags(ctx *cli.Context) (*FillArgs, error) {
	if ctx.String("db-token") == "" {
		return nil, fmt.Errorf("missing InfluxDB token. See -h for help")
	}
	duration, err := ParseDurationString(ctx.String("duration"))
	if err != nil {
		return nil, err
	}
	startAt, err := ParseDurationString(ctx.String("start-at"))
	if err != nil {
		return nil, err
	}
	gap, err := ParseDurationString(ctx.String("gap"))
	if err != nil {
		return nil, err
	}
	anomalyString, err := checkAnomalyString(ctx.String("anomaly"))
	if err != nil {
		return nil, err
	}

	if ctx.NArg() == 0 {
		return nil, fmt.Errorf("missing file(s). See -h for help")
	}
	// Validate the files
	files := ctx.Args().Slice()
	for _, file := range files {
		if err := ValidateFile(file); err != nil {
			return nil, err
		}
	}

	return &FillArgs{
		DBArgs: DBInfo{
			Token:       ctx.String("db-token"),
			Host:        ctx.String("db-host"),
			Port:        ctx.String("db-port"),
			Org:         ctx.String("db-org"),
			Bucket:      ctx.String("db-bucket"),
			Measurement: "metrics",
		},
		Duration: duration,
		StartAt:  startAt,
		Gap:      gap,
		Anomaly:  anomalyString,
		Files:    files,
	}, nil
}

// ParseStreamFlags parses the flags passed to the stream command
// Returns a StreamArgs struct containing the parsed flags
// Returns an error if the flags are invalid
func ParseStreamFlags(ctx *cli.Context) (*StreamArgs, error) {
	if ctx.String("db-token") == "" {
		return nil, fmt.Errorf("missing InfluxDB token. See -h for help")
	}
	duration, err := ParseDurationString(ctx.String("duration"))
	if err != nil {
		return nil, err
	}
	startAt, err := ParseDurationString(ctx.String("start-at"))
	if err != nil {
		return nil, err
	}
	if ctx.NArg() == 0 {
		return nil, fmt.Errorf("missing file. See -h for help")
	}
	if ctx.Int("time-multiplier") < 1 {
		return nil, fmt.Errorf("timemultiplier cannot be a lower than 1")
	}
	anomalyString, err := checkAnomalyString(ctx.String("anomaly"))
	if err != nil {
		return nil, err
	}
	file := ctx.Args().Slice()[0]
	err = ValidateFile(file)
	if err != nil {
		return nil, err
	}

	return &StreamArgs{
		DBArgs: DBInfo{
			Token:       ctx.String("db-token"),
			Host:        ctx.String("db-host"),
			Port:        ctx.String("db-port"),
			Org:         ctx.String("db-org"),
			Bucket:      ctx.String("db-bucket"),
			Measurement: "metrics",
		},
		Duration:       duration,
		StartAt:        startAt,
		TimeMultiplier: ctx.Int("time-multiplier"),
		Append:         ctx.Bool("append"),
		Anomaly:        anomalyString,
		File:           file,
	}, nil
}

// ParseCleanFlags parses the flags passed to the clean command
// Returns a CleanArgs struct containing the parsed flags
// Returns an error if the flags are invalid
func ParseCleanFlags(ctx *cli.Context) (*CleanArgs, error) {
	var duration time.Duration
	var err error
	if ctx.String("db-token") == "" {
		return nil, fmt.Errorf("missing InfluxDB token. See -h for help")
	}

	if ctx.String("start-at") == "" {
		duration = time.Now().Local().Sub(time.Unix(0, 0))
	} else {
		duration, err = ParseDurationString(ctx.String("start-at"))
		if err != nil {
			return nil, err
		}
	}

	if !ctx.Bool("all") && ctx.NArg() == 0 {
		return nil, fmt.Errorf("missing hostnames or --all flag. See -h for help")
	}

	hosts := ctx.Args().Slice()

	return &CleanArgs{
		DBArgs: DBInfo{
			Token:       ctx.String("db-token"),
			Host:        ctx.String("db-host"),
			Port:        ctx.String("db-port"),
			Org:         ctx.String("db-org"),
			Bucket:      ctx.String("db-bucket"),
			Measurement: ctx.String("db-measurement"),
		},
		All:      ctx.Bool("all"),
		Duration: duration,
		Hosts:    hosts,
	}, nil
}
