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

var supportedAlgorithms = map[string]func(ad *AnomalyDetection) ([]system_metrics.Anomaly, error){
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

func isolationForest(ad *AnomalyDetection) ([]system_metrics.Anomaly, error) {
	inputFilePath := "/tmp/go_output.csv"
	outputFilePath := "/tmp/py_output.csv"
	//Sets Arguments to the command

	// Write data to file
	if err := writeDataToFile(inputFilePath, ad.Data); err != nil {
		log.Printf("Error when writing data to file: %v", err)
		return []system_metrics.Anomaly{}, err
	}

	cmd := exec.Command("python", "anomaly_detection/outliers.py", inputFilePath, outputFilePath)
	//Better information in case of error in script execution
	cmd.Stderr = os.Stderr
	//executes command without regards of output. If output is needed change to cmd.Output()
	if err := cmd.Run(); err != nil {
		log.Printf("Error when running anomaly detection script: %v", err)
		return []system_metrics.Anomaly{}, err
	}
	//TODO outputfile path is local to my machine, change to your own
	anomalies, err := transformIFOutput("logs/dummyOutput.csv", ad.Data.Id)
	if err != nil {
		log.Printf("Error when transforming output: %v", err)
		return []system_metrics.Anomaly{}, err
	}
	return anomalies, nil
}

/*
Reads from anomaly detection output file and transforms data to anomalym struct
Returns Anomalystruct
Returns error if something fails
*/
func transformIFOutput(filename, host string) ([]system_metrics.Anomaly, error) {
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

type AnomalyMetric struct {
	Timestamp               int64 `csv:"timestamp" json:"timestamp"`
	Load1m                  bool  `csv:"load-1m" json:"load-1m"`
	Load5m                  bool  `csv:"load-5m" json:"load-5m"`
	Load15m                 bool  `csv:"load-15m" json:"load-15m"`
	Sys_Mem_Swap_Total      bool  `csv:"sys-mem-swap-total" json:"sys-mem-swap-total"`
	Sys_Mem_Swap_Free       bool  `csv:"sys-mem-swap-free" json:"sys-mem-swap-free"`
	Sys_Mem_Free            bool  `csv:"sys-mem-free" json:"sys-mem-free"`
	Sys_Mem_Cache           bool  `csv:"sys-mem-cache" json:"sys-mem-cache"`
	Sys_Mem_Buffered        bool  `csv:"sys-mem-buffered" json:"sys-mem-buffered"`
	Sys_Mem_Available       bool  `csv:"sys-mem-available" json:"sys-mem-available"`
	Sys_Mem_Total           bool  `csv:"sys-mem-total" json:"sys-mem-total"`
	Sys_Fork_Rate           bool  `csv:"sys-fork-rate" json:"sys-fork-rate"`
	Sys_Interrupt_Rate      bool  `csv:"sys-interrupt-rate" json:"sys-interrupt-rate"`
	Sys_Context_Switch_Rate bool  `csv:"sys-context-switch-rate" json:"sys-context-switch-rate"`
	Sys_Thermal             bool  `csv:"sys-thermal" json:"sys-thermal"`
	Disk_Io_Time            bool  `csv:"disk-io-time" json:"disk-io-time"`
	Disk_Bytes_Read         bool  `csv:"disk-bytes-read" json:"disk-bytes-read"`
	Disk_Bytes_Written      bool  `csv:"disk-bytes-written" json:"disk-bytes-written"`
	Disk_Io_Read            bool  `csv:"disk-io-read" json:"disk-io-read"`
	Disk_Io_Write           bool  `csv:"disk-io-write" json:"disk-io-write"`
	Cpu_Io_Wait             bool  `csv:"cpu-iowait" json:"cpu-iowait"`
	Cpu_System              bool  `csv:"cpu-system" json:"cpu-system"`
	Cpu_User                bool  `csv:"cpu-user" json:"cpu-user"`
	Server_Up               bool  `csv:"server-up" json:"server-up"`
}
