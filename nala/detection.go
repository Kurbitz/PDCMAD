package main

import (
	"fmt"
	"internal/influxdbapi"
	"internal/system_metrics"
	"log"
	"os"
	"os/exec"

	"github.com/gocarina/gocsv"
)

type AnomalyDetection struct {
	Duration string
	Data     system_metrics.SystemMetric
}

var supportedAlgorithms = map[string]func(ad *AnomalyDetection) (*[]system_metrics.AnomalyDetectionOutput, error){
	"IF": isolationForest,
}

func CheckSupportedAlgorithm(algorithm string) bool {
	_, ok := supportedAlgorithms[algorithm]
	return ok
}

func NewAnomalyDetection(dbapi influxdbapi.InfluxDBApi, host string, duration string) (*AnomalyDetection, error) {
	data, err := dbapi.GetMetrics(host, duration)
	if err != nil {
		log.Printf("Error when getting metrics from influxdb: %v", err)
		return nil, err
	}

	return &AnomalyDetection{
		Duration: duration,
		Data:     data,
	}, nil
}

func isolationForest(ad *AnomalyDetection) (*[]system_metrics.AnomalyDetectionOutput, error) {
	inputFilePath := "/tmp/go_output.csv"
	outputFilePath := "/tmp/py_output.csv"
	//Sets Arguments to the command

	// Write data to file
	if err := writeDataToFile(inputFilePath, ad.Data); err != nil {
		log.Printf("Error when writing data to file: %v", err)
		return nil, err
	}

	cmd := exec.Command("python", "anomaly_detection/outliers.py", inputFilePath, outputFilePath)
	//Better information in case of error in script execution
	cmd.Stderr = os.Stderr
	//executes command without regards of output. If output is needed change to cmd.Output()
	if err := cmd.Run(); err != nil {
		log.Printf("Error when running anomaly detection script: %v", err)
		return nil, err
	}
	//TODO outputfile path is local to my machine, change to your own
	anomalies, err := parseIFOutput(outputFilePath, ad.Data.Id)
	if err != nil {
		log.Printf("Error when transforming output: %v", err)
		return nil, err
	}
	return anomalies, nil
}

/*
Reads from anomaly detection output file and transforms data to anomalym struct
Returns Anomalystruct
Returns error if something fails
*/
func parseIFOutput(filename, host string) (*[]system_metrics.AnomalyDetectionOutput, error) {
	anomalyData := []system_metrics.AnomalyDetectionOutput{}
	inputFile, err := os.OpenFile(filename, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Printf("Error when opening file: %v", err)
		return nil, err
	}
	defer inputFile.Close()
	if err = gocsv.UnmarshalFile(inputFile, &anomalyData); err != nil {
		log.Printf("Error when parsing anomaly detection csv: '%v'", err)
		return nil, err
	}
	if len(anomalyData) == 0 {
		return nil, fmt.Errorf("output of anomaly detection is empty")
	}

	return &anomalyData, nil
}

func writeDataToFile(filePath string, data system_metrics.SystemMetric) error {
	outputFile, err := os.Create(filePath)
	if err != nil {
		log.Printf("Error when creating file: %v", err)
		return err
	}
	defer outputFile.Close()
	err = gocsv.MarshalFile(&data.Metrics, outputFile)
	if err != nil {
		log.Printf("Error while parsing metrics from file: %v", err)
		return err
	}
	return nil
}
