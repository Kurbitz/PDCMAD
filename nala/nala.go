package main

import (
	"fmt"
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

const (
	PATH = "../anomaly_detection/"
)

func triggerDetection(ctx *gin.Context) {
	//TODO change to environment variables
	dbapi := influxdbapi.NewInfluxDBApi("KBntTYJdaWbknRyM-CAw29iYdJmQkiK6C1vlEO3B5yuvgGJlmG4Gasps5rTRGflLq7bRSSWZSA_zdnYhpu-HXQ==", "localhost", "8086")
	defer dbapi.Close()
	host := ctx.Param("host")
	duration := ctx.Param("duration")
	//These checks might be in simba instead?
	if host == "" {
		ctx.String(http.StatusOK, "Host field is empty")
		return
	}
	if duration == "" {
		ctx.String(http.StatusOK, "Duration field is empty")
		return
	}
	if message, err := dbapi.GetMetrics(host, duration); err == nil {
		go func() {
			if err := triggerIsolationForest("outliers.py", message, host); err != nil {
				log.Printf("Anomaly detection failed with: %v\n", err)
				return
			}
			log.Println("Anomaly detection is done!")
		}()
	} else {
		ctx.String(http.StatusOK, "Error while getting metrics:\n%v", err)
		return
	}
	ctx.String(http.StatusOK, "Anomaly detection triggered!\n")
}

func writeToFile(filePath string, data system_metrics.SystemMetric) error {
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

// Runs "outliers.py" and wraps the output
func triggerIsolationForest(filename string, data system_metrics.SystemMetric, host string) error {
	if err := writeToFile("logs/go_output.csv", data); err != nil {
		return err
	}
	fullPath := PATH + filename
	inputFilePath := "../nala/logs/go_output.csv"
	outputFilePath := "../nala/logs/py_output.csv"
	//Sets Arguments to the command
	cmd := exec.Command("python", fullPath, inputFilePath, outputFilePath)
	//Better information in case of error in script execution
	cmd.Stderr = os.Stderr
	//executes command without regards of output. If output is needed change to cmd.Output()
	if err := cmd.Run(); err != nil {
		log.Printf("Error when running anomaly detection script: %v", err)
		return err
	}

	anomalies, err := transformOutput("logs/dummyOutput.csv")
	if err != nil {
		return err
	}
	logAnomalies(anomalies, host)
	return nil
}

/*
Reads from anomaly detection output file and transforms data to anomalymetric struct
Returns AnomalyMetric struct
Returns error if something fails
*/
func transformOutput(filename string) ([]AnomalyMetric, error) {
	anomalyData := []AnomalyMetric{}
	inputFile, err := os.OpenFile(filename, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Printf("Error when opening file: %v", err)
		return []AnomalyMetric{}, err
	}
	defer inputFile.Close()
	if err = gocsv.UnmarshalFile(inputFile, &anomalyData); err != nil {
		log.Printf("Error when parsing anomaly detection csv: '%v'", err)
		return []AnomalyMetric{}, err
	}
	if len(anomalyData) == 0 {
		return []AnomalyMetric{}, fmt.Errorf("Output of anomaly detection is empty")
	}
	return anomalyData, nil
}

/*
Takes AnomalyMetric struct and writes it to a log file
Logfile output: [time, host, metric, comment]
Returns error if something fails
*/
func logAnomalies(anomalies []AnomalyMetric, host string) {
	outputArray := []Anomaly{}
	for _, v := range anomalies {
		r := reflect.ValueOf(v)
		for i := 1; i < r.NumField(); i++ {
			if r.Field(i).Interface() == true {
				outputArray = append(outputArray, Anomaly{Timestamp: v.Timestamp, Host: host, Metric: r.Type().Field(i).Tag.Get("csv"), Coment: ""})
			}
		}
	}
	//TODO write output data to database
	for _, o := range outputArray {
		println(o.Timestamp, o.Host, o.Metric, o.Coment)
	}
	return
}

type Anomaly struct {
	Timestamp int64  `csv:"timestamp" json:"timestamp"`
	Host      string `csv:"host" json:"host"`
	Metric    string `csv:"metric" json:"metric"`
	Coment    string `csv:"coment" json:"coment"`
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

func main() {
	router := gin.Default()
	router.GET("/nala/IF/:host/:duration", triggerDetection)
	router.Run("localhost:8088")
}
