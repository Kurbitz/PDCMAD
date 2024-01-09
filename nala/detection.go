package main

import (
	"internal/influxdbapi"
	"internal/system_metrics"
	"log"
	"os"
	"os/exec"
)

// AnomalyDetectionParameters is a struct that contains all the parameters that are needed for an anomaly detection.
// It is used to pass the parameters to the anomaly detection functions.
// New fields can be added to this struct if they are needed for new anomaly detection algorithms. This way we can keep the function signatures the same.
// A zero value of this struct is not usable. Use the NewAnomalyDetectionParameters function instead.
type AnomalyDetectionParameters struct {
	Data system_metrics.SystemMetric // The metrics that will be used for the anomaly detection, contains the id of the host
}

// supportedAlgorithms is a map that maps algorithm names to anomaly detection functions.
// The algorithm names are the same as the algorithm flags that can be passed to the anomaly detection endpoint.
// To add a new algorithm, add a new entry to this map with the algorithm name as the key and the anomaly detection function as the value.
var SupportedAlgorithms = map[string]func(ad *AnomalyDetectionParameters) (*[]system_metrics.AnomalyDetectionOutput, error){
	"IF": isolationForest,
}

// NewAnomalyDetectionParameters creates a new AnomalyDetectionParameters struct.
// Gets the metrics from the database and returns them in the struct.
// Returns an error if something fails.
// If adding new parameters to the AnomalyDetectionParameters struct, they should be added here as well.
func NewAnomalyDetectionParameters(dbapi influxdbapi.InfluxDBApi, host string, duration string) (*AnomalyDetectionParameters, error) {
	log.Println("Getting metrics from influxdb")
	data, err := dbapi.GetMetrics(host, duration)
	if err != nil {
		log.Printf("Error when getting metrics from influxdb: %v", err)
		return nil, err
	}
	log.Println("Metrics received from influxdb")
	return &AnomalyDetectionParameters{
		Data: data,
	}, nil
}

// isolationForest is an anomaly detection function that uses the Isolation Forest algorithm implemented in Python.
// It writes the metrics to a csv file, then calls the Python script that runs the Isolation Forest algorithm on the csv file.
// The output of the python script is then parsed and returned.
// For more information about the details of the Isolation Forest algorithm, see the documentation of the Python script.
// Requires certain Python packages to be installed, these are listed in the requirements.txt file.
// If running the program in a Docker container, these packages will be installed automatically when building the image from the Dockerfile.
// If running the program locally, these packages will need to be installed manually, see the anomaly_detection/README.md file for more information.
// Returns an error if something fails.
func isolationForest(ad *AnomalyDetectionParameters) (*[]system_metrics.AnomalyDetectionOutput, error) {
	log.Println("Starting anomaly detection with Isolation Forest")

	inputFilePath := "/tmp/input.csv"
	outputFilePath := "/tmp/output.csv"
	log.Println("Writing data to file")

	// Write data to file
	if err := ad.Data.WriteToFile(inputFilePath); err != nil {
		log.Printf("Error when writing data to file: %v", err)
		return nil, err
	}
	log.Println("Executing anomaly detection script")

	// Set up command
	cmd := exec.Command("python", "anomaly_detection/outliers.py", inputFilePath, outputFilePath)
	// Stderr must be set to os.Stderr to get the error output from the script
	cmd.Stderr = os.Stderr

	// Executes command without regards of output. If output is needed change to cmd.Output()
	if err := cmd.Run(); err != nil {
		log.Printf("Error when running anomaly detection script: %v", err)
		return nil, err
	}
	log.Println("Scipt finished executing, parsing output")

	// Parse the file that the script wrote
	anomalies, err := system_metrics.ParseAnomalyDetectionOutputCSV(outputFilePath, ad.Data.Id)
	if err != nil {
		log.Printf("Error when transforming output: %v", err)
		return nil, err
	}
	log.Println("Isoaltion Forest anomaly detection done!")
	return anomalies, nil
}
