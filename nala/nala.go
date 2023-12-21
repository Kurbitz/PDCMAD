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

// TODO Create a trigger endpoint
func triggerDetection(ctx *gin.Context) {
	dbapi := influxdbapi.NewInfluxDBApi("KBntTYJdaWbknRyM-CAw29iYdJmQkiK6C1vlEO3B5yuvgGJlmG4Gasps5rTRGflLq7bRSSWZSA_zdnYhpu-HXQ==", "localhost", "8086")
	defer dbapi.Close()
	host := ctx.Param("host")
	duration := ctx.Param("duration")
	if message, err := dbapi.GetMetrics(host, duration); err == nil {
		go func() {
			triggerIsolationForest("outliers.py", message, host)
			log.Println("Anomaly detection is done!")
		}()
	} else {
		ctx.String(http.StatusOK, "Error while getting metrics:\n%v", err)
		return
	}
	ctx.String(http.StatusOK, "Anomaly detection triggered!")
}

// Runs "outliers.py" and wraps the output
func triggerIsolationForest(filename string, data system_metrics.SystemMetric, host string) {
	//Sets Arguments to the command
	outputFile, err := os.Create("logs/go_output.csv")
	if err != nil {
		log.Println(err)
	}
	defer outputFile.Close()
	gocsv.MarshalFile(&data.Metrics, outputFile)
	if err != nil {
		log.Println(err)
	}
	fullPath := PATH + filename
	inputFilePath := "../nala/logs/go_output.csv"
	outputFilePath := "../nala/logs/py_output.csv"
	cmd := exec.Command(PATH+"/bin/python", fullPath, inputFilePath, outputFilePath)
	cmd.Stderr = os.Stderr
	anomalyData := system_metrics.SystemMetric{Id: data.Id}
	//executes command, listends to stdout, puts w/e into "out" var unless error
	if out, err := cmd.Output(); err != nil {
		log.Println(err)
		log.Println(string(out))
	} else {
		fmt.Println(string(out))
		inputFile, err := os.OpenFile("logs/py_output.csv", os.O_RDONLY, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
		defer inputFile.Close()
		if err = gocsv.UnmarshalFile(inputFile, &anomalyData.Metrics); err != nil {
			log.Printf("Error when parsing anomaly detection csv: '%v'", err)
		}
		log.Println(anomalyData.Metrics[0].Cpu_System)
		//TODO wrap anomaly output
		//anomalies, err := transformOutput(outputFilePath)
		anomalies, err := transformOutput("./dummyOutput.csv")
		if err != nil {
			log.Println(err)
		}
		logAnomalies(anomalies, host)
	}

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
		log.Println(err)
	}
	defer inputFile.Close()
	if err = gocsv.UnmarshalFile(inputFile, &anomalyData); err != nil {
		log.Printf("Error when parsing anomaly detection csv: '%v'", err)
	}
	log.Println(anomalyData[0])
	return anomalyData, nil
}

/*
Takes AnomalyMetric struct and writes it to a log file
Logfile output: [time, host, metric, comment]
Returns error if something fails
*/
func logAnomalies(anomalies []AnomalyMetric, host string) {
	if anomalies == nil || host == "" {
		panic("anomalies or host is missing")
	}
	outputArray := []Output{}
	for _, v := range anomalies { //len(anomalies)
		r := reflect.ValueOf(v)
		for i := 1; i < r.NumField(); i++ {
			if r.Field(i).Interface() == true {
				outputArray = append(outputArray, Output{Timestamp: v.Timestamp, Host: host, Metric: r.Type().Field(i).Name, Coment: ""})
			}
		}
	}
	for _, o := range outputArray {
		println(o.Timestamp, o.Host, o.Metric, o.Coment)
	}
	return
}

type Output struct {
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
