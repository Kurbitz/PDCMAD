package main

import (
	"internal/influxdbapi"
	"internal/system_metrics"
	"log"
	"net/http"
	"os"
	"os/exec"
	"reflect"

	"github.com/gocarina/gocsv"

	"github.com/gin-gonic/gin"
)

// inProgress is a global status variable that is used to check if an anomaly detection is already in progress.
// FIXME: This is not thread safe, but it should be fine for what we are doing. We could probably use a mutex to make it thread safe.
var inProgress = false

func triggerDetection(ctx *gin.Context) {
	log.Println("Anomaly detection request received!")
	algorithm := ctx.Param("algorithm")
	host := ctx.Param("host")
	duration := ctx.Param("duration")

	if host == "" {
		ctx.String(http.StatusBadRequest, "Host field is empty")
		log.Println("Host field is empty")
		return
	}
	if duration == "" {
		ctx.String(http.StatusBadRequest, "Duration field is empty")
		log.Println("Duration field is empty")
		return
	}
	if inProgress {
		ctx.String(http.StatusConflict, "Anomaly detection is already in progress")
		log.Println("Anomaly detection is already in progress")
		return
	}

	// Check if the algorithm is supported by looking it up in the supportedAlgorithms map
	detection, exists := SupportedAlgorithms[algorithm]
	if !exists {
		ctx.String(http.StatusBadRequest, "Algorithm %v is not supported", algorithm)
		log.Printf("Algorithm %v is not supported", algorithm)
		return
	}

	// Set inProgress to true so that we can't trigger another anomaly detection while one is already running
	inProgress = true
	log.Printf("Starting anomaly detection for %v\n", host)

	// Create the influxdb api
	dbapi := influxdbapi.NewInfluxDBApi(os.Getenv("INFLUXDB_TOKEN"), os.Getenv("INFLUXDB_HOST"), os.Getenv("INFLUXDB_PORT"), os.Getenv("INFLUXDB_ORG"), os.Getenv("INFLUXDB_BUCKET"), "metrics")
	defer dbapi.Close()

	// Create the parameters for the anomaly detection
	// This is done here so that we can return an error if the parameters are invalid
	// The data from the databse will be fetched here
	parameters, err := NewAnomalyDetectionParameters(dbapi, host, duration)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "%v", err)
		inProgress = false
		return
	}

	// Start a goroutine that will run the anomaly detection
	go func() {
		// Make sure to set inProgress to false when the function returns
		defer func() {
			inProgress = false
		}()

		// Trigger the anomaly detection, this may take a while and can fail in unknown ways all depending on the algorithm
		anomalies, err := detection(parameters)
		if err != nil {
			log.Printf("Anomaly detection failed with: %v\n", err)
			return
		}
		log.Println("Logging anomalies to file")
		if err = logAnomalies("/tmp/anomalies.csv", host, algorithm, *anomalies); err != nil {
			log.Printf("Error when writing anomalies to file: %v\n", err)
			return
		}

		// Make sure to set the measurement to anomalies before writing to influxdb
		dbapi.Measurement = "anomalies"
		log.Println("Writing anomalies to influxdb")
		if err = dbapi.WriteAnomalies(*anomalies, host, algorithm); err != nil {
			log.Printf("Error when writing anomalies to influxdb: %v\n", err)
			return
		}
		log.Println("Anomaly detection is done!")
	}()

	// If everything went well, return a 200 OK
	ctx.String(http.StatusOK, "Anomaly detection triggered!\n")
}

// pythonSmokeTest runs a simple python script to check if the Python environment is working
// If the Python environment is not working, the program will exit.
func pythonSmokeTest() {
	log.Println("Running python smoke test...")

	// Define command to run
	cmd := exec.Command("python", "./testpy.py", "Python is working!")

	// Executes command, waits for it to finish and returns output
	// If the commands terminates with an error code, err will be set
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	// Print the Python output
	// output needs to be converted to string
	log.Print(string(out))
	log.Println("Python smoke test complete!")
}

// logAnomalies takes a slice of AnomalyDetectionOutput structs, converts it to AnomalyEvent structs and writes it to a log file.
// The format of the log file is defined by the AnomalyEvent struct.
// The log file is written to the path specified by filePath.
// If the file does not exist, it will be created, if it does exist, it will be appended to.
// Returns an error if any of the steps fail.
func logAnomalies(filePath string, host string, algorithm string, data []system_metrics.AnomalyDetectionOutput) error {
	// Convert the AnomalyDetectionOutput structs to AnomalyEvent structs
	outputArray := []system_metrics.AnomalyEvent{}
	for _, v := range data {
		// Use reflection to get the fields of the struct
		r := reflect.ValueOf(v)

		// Iterate over the fields and if any value is true, add it to the output array as an AnomalyEvent
		// This means any single AnomalyDetectionOutput struct can result in multiple AnomalyEvent structs
		// Skip the first field (Timestamp) because it is not a bool
		for i := 1; i < r.NumField(); i++ {
			if r.Field(i).Interface() == true {
				// Get the name of the CSV tag of the field, this is the name of the metric
				outputArray = append(outputArray, system_metrics.AnomalyEvent{Timestamp: v.Timestamp, Host: host, Metric: r.Type().Field(i).Tag.Get("csv"), Comment: algorithm})
			}
		}
	}

	// Open the file for writing and append to it if it exists
	outputFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Error when creating file: %v", err)
		return err
	}
	defer outputFile.Close()

	// Write the AnomalyEvent structs to the file
	err = gocsv.MarshalFile(&outputArray, outputFile)
	if err != nil {
		log.Printf("Error while parsing metrics from file: %v", err)
		return err
	}
	return nil
}

// checkEnv checks if all the required environment variables are set.
// If any of the required environment variables are not set, the program will exit.
func checkEnv() {
	log.Println("Checking environment variables...")

	if _, exists := os.LookupEnv("INFLUXDB_HOST"); !exists {
		log.Fatal("INFLUXDB_HOST is not set")
	}
	if _, exists := os.LookupEnv("INFLUXDB_PORT"); !exists {
		log.Fatal("INFLUXDB_PORT is not set")
	}
	if _, exists := os.LookupEnv("INFLUXDB_TOKEN"); !exists {
		log.Fatal("INFLUXDB_TOKEN is not set")
	}
	if _, exists := os.LookupEnv("INFLUXDB_ORG"); !exists {
		log.Fatal("INFLUXDB_TOKEN is not set")
	}
	if _, exists := os.LookupEnv("INFLUXDB_BUCKET"); !exists {
		log.Fatal("INFLUXDB_TOKEN is not set")
	}
	log.Println("Environment variables are set!")
}

// setupEndpoints sets up the API endpoints for the router.
func setupEndpoints(router *gin.Engine) {
	router.GET("/nala/:algorithm/:host/:duration", triggerDetection)

	router.GET("/nala/test", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Nala is working!")
	})

	router.GET("/nala/status", func(ctx *gin.Context) {
		responseText := ""
		if inProgress {
			responseText = "Anomaly detection in progress"
		} else {
			responseText = "No anomaly detection running"
		}
		ctx.String(http.StatusOK, responseText)
	})
}

func main() {
	log.Println("Starting Nala...")
	// Run some startup checks
	pythonSmokeTest()
	checkEnv()

	// Create the router and setup the endpoints
	router := gin.Default()
	setupEndpoints(router)

	// Start the router
	// Needs 0.0.0.0 to bind to any interface
	router.Run("0.0.0.0:8088")
}
