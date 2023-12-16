package simba

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/urfave/cli/v2"
)

// FillArgs is a struct containing the flags passed to the fill command
type FillArgs struct {
	DBToken  string
	DBIp     string
	DBPort   string
	Duration time.Duration
	StartAt  time.Duration
	Gap      time.Duration
	Files    []string
}

// StreamArgs is a struct containing the flags passed to the stream command
type StreamArgs struct {
	DBToken        string
	DBIp           string
	DBPort         string
	Duration       time.Duration
	Startat        time.Duration
	TimeMultiplier int
	Append         bool
	File           string
}


// Common flags for the simulate command
var simulateFlags = []cli.Flag{
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

var cleanFlags = []cli.Flag{
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
		Value: "",
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
					Usage:     "fill the database with data from file(s)",
					ArgsUsage: "<file1> <file2> ...",
					Action:    invokeFill,
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
					Action:    invokeStream,
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
		DBToken:  ctx.String("dbtoken"),
		DBIp:     ctx.String("dbip"),
		DBPort:   ctx.String("dbport"),
		Duration: duration,
		StartAt:  startAt,
		Gap:      gap,
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
	file := ctx.Args().Slice()[0]
	err = ValidateFile(file)
	if err != nil {
		return nil, err
	}

	return &StreamArgs{
		DBToken:        ctx.String("dbtoken"),
		DBIp:           ctx.String("dbip"),
		DBPort:         ctx.String("dbport"),
		Duration:       duration,
		Startat:        startAt,
		TimeMultiplier: ctx.Int("timemultiplier"),
		Append:         ctx.Bool("append"),
		File:           file,
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

// The cleanBucket subcommand removes all the data inside the specified bucket
func cleanBucket(ctx *cli.Context) error {
	var start time.Duration
	var err error
	// Validate ctx.Args contains a bucket name
	if ctx.NArg() == 0 {
		return cli.Exit("Missing bucket name", 1)
	}

	if ctx.String("dbtoken") == "" {
		return cli.Exit("Missing InfluxDB token. See -h for help", 1)
	}

	bucket := ctx.Args().First()

	if ctx.String("startat") == "" {
		start = time.Now().Local().Sub(time.Unix(0, 0))
	} else {
		start, err = ParseDurationString(ctx.String("startat"))
		if err != nil {
			return cli.Exit(err, 1)
		}
	}

	influxDBApi := NewInfluxDBApi(ctx.String("dbtoken"), ctx.String("dbip"), ctx.String("dbport"))

	return influxDBApi.DeleteBucket(bucket, start)
}

// The cleanHost subcommand removes all the data from the desired
// hosts/systems inside the specified bucket
func cleanHost(ctx *cli.Context) error {
	var start time.Duration
	var err error

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

	if ctx.String("startat") == "" {
		start = time.Now().Local().Sub(time.Unix(0, 0))
	} else {
		start, err = ParseDurationString(ctx.String("startat"))
		if err != nil {
			return cli.Exit(err, 1)
		}
	}

	influxDBApi := NewInfluxDBApi(ctx.String("dbtoken"), ctx.String("dbip"), ctx.String("dbport"))

	var wg sync.WaitGroup
	for _, host := range ctx.Args().Slice() {
		wg.Add(1)
		go func(hostName string, bucketName string) {
			defer wg.Done()
			influxDBApi.DeleteHost(bucket, hostName, start)
		}(host, bucket)
	}
	wg.Wait()

	return nil
}
