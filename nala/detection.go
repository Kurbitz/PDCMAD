package main

import (
	"internal/influxdbapi"
	"internal/system_metrics"
	"log"
	"os"
	"os/exec"
)

type AnomalyDetectionParameters struct {
	Data system_metrics.SystemMetric
}

var SupportedAlgorithms = map[string]func(ad *AnomalyDetectionParameters) (*[]system_metrics.AnomalyDetectionOutput, error){
	"IF": isolationForest,
}

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
	cmd := exec.Command("python", "anomaly_detection/outliers.py", inputFilePath, outputFilePath)
	//Better information in case of error in script execution
	cmd.Stderr = os.Stderr
	//executes command without regards of output. If output is needed change to cmd.Output()
	if err := cmd.Run(); err != nil {
		log.Printf("Error when running anomaly detection script: %v", err)
		return nil, err
	}
	log.Println("Scipt finished executing, parsing output")
	anomalies, err := system_metrics.ParseAnomalyDetectionOutputCSV(outputFilePath, ad.Data.Id)
	if err != nil {
		log.Printf("Error when transforming output: %v", err)
		return nil, err
	}
	log.Println("Isoaltion Forest anomaly detection done!")
	return anomalies, nil
}
