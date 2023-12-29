package main

import (
	"fmt"
	"internal/influxdbapi"
	"internal/system_metrics"
	"log"
	"os"
	"os/exec"
	"reflect"

	"github.com/gocarina/gocsv"
)

type AnomalyDetection struct {
	Duration string
	Data     system_metrics.SystemMetric
}

var supportedAlgorithms = map[string]func(ad *AnomalyDetection) error{
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

func isolationForest(ad *AnomalyDetection) error {
	inputFilePath := "/tmp/go_output.csv"
	outputFilePath := "/tmp/py_output.csv"
	//Sets Arguments to the command

	cmd := exec.Command("python", "anomaly_detection/outliers.py", inputFilePath, outputFilePath)
	//Better information in case of error in script execution
	cmd.Stderr = os.Stderr
	//executes command without regards of output. If output is needed change to cmd.Output()
	if err := cmd.Run(); err != nil {
		log.Printf("Error when running anomaly detection script: %v", err)
		return err
	}
	return nil
}

/*
Reads from anomaly detection output file and transforms data to anomalym struct
Returns Anomalystruct
Returns error if something fails
*/
func transformOutput(filename, host string) ([]system_metrics.Anomaly, error) {
	anomalyData := []AnomalyMetric{}
	inputFile, err := os.OpenFile(filename, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Printf("Error when opening file: %v", err)
		return []system_metrics.Anomaly{}, err
	}
	defer inputFile.Close()
	if err = gocsv.UnmarshalFile(inputFile, &anomalyData); err != nil {
		log.Printf("Error when parsing anomaly detection csv: '%v'", err)
		return []system_metrics.Anomaly{}, err
	}
	if len(anomalyData) == 0 {
		return []system_metrics.Anomaly{}, fmt.Errorf("output of anomaly detection is empty")
	}
	outputArray := []system_metrics.Anomaly{}
	for _, v := range anomalyData {
		r := reflect.ValueOf(v)
		for i := 1; i < r.NumField(); i++ {
			if r.Field(i).Interface() == true {
				outputArray = append(outputArray, system_metrics.Anomaly{Timestamp: v.Timestamp, Host: host, Metric: r.Type().Field(i).Tag.Get("csv"), Comment: "Isolation forest"})
			}
		}
	}
	return outputArray, nil
}
