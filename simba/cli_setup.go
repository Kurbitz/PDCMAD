package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/urfave/cli/v2"
)

type DBInfo struct {
	Token       string
	Host        string
	Port        string
	Org         string
	Bucket      string
	Measurement string
}

// FillArgs is a struct containing the flags passed to the fill command
type FillArgs struct {
	DBArgs   DBInfo
	Duration time.Duration
	StartAt  time.Duration
	Gap      time.Duration
	Anomaly  string
	Files    []string
}

// StreamArgs is a struct containing the flags passed to the stream command
type StreamArgs struct {
	DBArgs         DBInfo
	Duration       time.Duration
	Startat        time.Duration
	TimeMultiplier int
	Append         bool
	Anomaly        string
	File           string
}

// CleanArgs is a struct containing the flags passed to the clean command
type CleanArgs struct {
	DBArgs  DBInfo
	All     bool
	Startat time.Duration
	Hosts   []string
}

// Common flags for the simulate command
var simulateFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  "duration",
		Usage: "Duration",
		Value: "",
		Aliases: []string{
			"d",
		},
	},
	&cli.StringFlag{
		Name:  "startat",
		Usage: "Starting line in file",
		Value: "",
		Aliases: []string{
			"s",
		},
	},
	&cli.StringFlag{
		Name:  "anomaly",
		Usage: "Select which type of anomaly to use",
		Value: "",
		Aliases: []string{
			"a",
		},
	},
	&cli.StringFlag{
		Name:     "dbtoken",
		EnvVars:  []string{"INFLUXDB_TOKEN"},
		Usage:    "InfluxDB token",
		Value:    "",
		Category: "Database",
		Aliases: []string{
			"t",
		},
	},
	&cli.StringFlag{
		Name:     "dbhost",
		EnvVars:  []string{"INFLUXDB_HOST"},
		Usage:    "InfluxDB hostname",
		Value:    "localhost",
		Category: "Database",
		Aliases: []string{
			"h",
		},
	},
	&cli.StringFlag{
		Name:     "dbport",
		EnvVars:  []string{"INFLUXDB_PORT"},
		Usage:    "InfluxDB port",
		Value:    "8086",
		Category: "Database",
		Aliases: []string{
			"p",
		},
	},
	&cli.StringFlag{
		Name:     "dborg",
		Usage:    "InfluxDB organization",
		EnvVars:  []string{"INFLUXDB_ORG"},
		Value:    "pdc-mad",
		Category: "Database",
		Aliases: []string{
			"o",
		},
	},
	&cli.StringFlag{
		Name:     "dbbucket",
		Usage:    "InfluxDB bucket",
		EnvVars:  []string{"INFLUXDB_BUCKET"},
		Value:    "pdc-mad",
		Category: "Database",
		Aliases: []string{
			"b",
		},
	},
}

// Flags for the clean command
var cleanFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  "startat",
		Usage: "from where to delete relative to current time",
		Value: "",
		Aliases: []string{
			"s",
		},
	},
	&cli.BoolFlag{
		Name:  "all",
		Usage: "delete metrics from all the hosts of the bucket",
		Value: false,
	},
	&cli.StringFlag{
		Name:     "dbtoken",
		EnvVars:  []string{"INFLUXDB_TOKEN"},
		Usage:    "InfluxDB token",
		Value:    "",
		Category: "Database",
		Aliases: []string{
			"t",
		},
	},
	&cli.StringFlag{
		Name:     "dbhost",
		EnvVars:  []string{"INFLUXDB_HOST"},
		Usage:    "InfluxDB IP",
		Value:    "localhost",
		Category: "Database",
		Aliases: []string{
			"h",
		},
	},
	&cli.StringFlag{
		Name:     "dbport",
		EnvVars:  []string{"INFLUXDB_PORT"},
		Usage:    "InfluxDB port",
		Value:    "8086",
		Category: "Database",
		Aliases: []string{
			"p",
		},
	},
	&cli.StringFlag{
		Name:     "dborg",
		Usage:    "InfluxDB organization",
		EnvVars:  []string{"INFLUXDB_ORG"},
		Value:    "pdc-mad",
		Category: "Database",
		Aliases: []string{
			"o",
		},
	},
	&cli.StringFlag{
		Name:     "dbbucket",
		Usage:    "InfluxDB bucket",
		EnvVars:  []string{"INFLUXDB_BUCKET"},
		Value:    "pdc-mad",
		Category: "Database",
		Aliases: []string{
			"b",
		},
	},
}

// App is the main application
// All commands and flags are defined here
var App = &cli.App{
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
					Usage:     "Fill the database with data from file(s)",
					ArgsUsage: "<file1> <file2> ...",
					Action:    invokeFill,
					Flags: append(simulateFlags, &cli.StringFlag{
						Name:  "gap",
						Usage: "Gap to now",
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
					Action:    invokeStream,
					Flags: append(simulateFlags, &cli.IntFlag{
						Name:  "timemultiplier",
						Usage: "Increase insertion speed",
						Value: 1,
						Aliases: []string{
							"m",
						},
					}, &cli.BoolFlag{
						Name:  "append",
						Usage: "Insert from the latest metric",
						Value: false,
					}),
				},
			},
		},
		{
			Name:      "clean",
			Usage:     "Clean the database",
			ArgsUsage: "<host1> <host2> ...",
			Action:    invokeClean,
			Flags:     cleanFlags,
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

// checkANomalyString checks if the anomalyString given exists in the AnomalyMap
// if it does not exists it returns an error
// if it does exist or is empty the anomalyString gets returned
func checkAnomalyString(anomalyString string) (string, error) {
	if anomalyString == "" {
		return anomalyString, nil
	}
	if _, exists := AnomalyMap[anomalyString]; exists {
		return anomalyString, nil
	}

	return anomalyString, fmt.Errorf("error injection %s is not implemented", anomalyString)
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

func ParseFillFlags(ctx *cli.Context) (*FillArgs, error) {
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
			Token:       ctx.String("dbtoken"),
			Host:        ctx.String("dbhost"),
			Port:        ctx.String("dbport"),
			Org:         ctx.String("dborg"),
			Bucket:      ctx.String("dbbucket"),
			Measurement: "metrics",
		},
		Duration: duration,
		StartAt:  startAt,
		Gap:      gap,
		Anomaly:  anomalyString,
		Files:    files,
	}, nil
}

func ParseStreamFlags(ctx *cli.Context) (*StreamArgs, error) {
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
	if ctx.NArg() == 0 {
		return nil, fmt.Errorf("missing file. See -h for help")
	}
	if ctx.Int("timemultiplier") < 1 {
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
			Token:       ctx.String("dbtoken"),
			Host:        ctx.String("dbhost"),
			Port:        ctx.String("dbport"),
			Org:         ctx.String("dborg"),
			Bucket:      ctx.String("dbbucket"),
			Measurement: "metrics",
		},
		Duration:       duration,
		Startat:        startAt,
		TimeMultiplier: ctx.Int("timemultiplier"),
		Append:         ctx.Bool("append"),
		Anomaly:        anomalyString,
		File:           file,
	}, nil
}

// FIXME: Use better ID
func GetIdFromFileName(file string) string {
	// Remove the file extension from the base file name
	return filepath.Base(file)[:len(filepath.Base(file))-len(filepath.Ext(file))]

}

func ParseCleanFlags(ctx *cli.Context) (*CleanArgs, error) {
	var startAt time.Duration
	var err error
	if ctx.String("dbtoken") == "" {
		return nil, fmt.Errorf("missing InfluxDB token. See -h for help")
	}

	if ctx.String("startat") == "" {
		startAt = time.Now().Local().Sub(time.Unix(0, 0))
	} else {
		startAt, err = ParseDurationString(ctx.String("startat"))
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
			Token:       ctx.String("dbtoken"),
			Host:        ctx.String("dbhost"),
			Port:        ctx.String("dbport"),
			Org:         ctx.String("dborg"),
			Bucket:      ctx.String("dbbucket"),
			Measurement: "metrics",
		},
		All:     ctx.Bool("all"),
		Startat: startAt,
		Hosts:   hosts,
	}, nil
}

// The invokeStream command reads a single file and sends them to InfluxDB in real time
// The file is passed as an argument to the application (simba invokeStream file.csv)
func invokeStream(ctx *cli.Context) error {
	// Parse the flags
	flags, err := ParseStreamFlags(ctx)
	if err != nil {
		return cli.Exit(err, 1)
	}
	if err := Stream(*flags); err != nil {
		return cli.Exit(err, 1)
	}

	return nil
}

// The invokeFill command reads the files and sends them to InfluxDB
// The files are passed as arguments to the application (simba invokeFill file1.csv file2.csv etc.)
func invokeFill(ctx *cli.Context) error {
	// Parse the flags
	flags, err := ParseFillFlags(ctx)
	if err != nil {
		return cli.Exit(err, 1)
	}
	if err := Fill(*flags); err != nil {
		return cli.Exit(err, 1)
	}
	return nil
}

func invokeClean(ctx *cli.Context) error {
	//Parse the flags
	flags, err := ParseCleanFlags(ctx)
	if err != nil {
		return cli.Exit(err, 1)
	}
	if err := Clean(*flags); err != nil {
		return cli.Exit(err, 1)
	}
	return nil
}
