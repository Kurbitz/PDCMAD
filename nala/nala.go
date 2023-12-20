package main

import (
	"fmt"
	"internal/influxdbapi"
	"internal/system_metrics"
	"log"
	"net/http"
	"os"
	"os/exec"

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
			triggerIsolationForest("outliers.py", message)
			log.Println("Anomaly detection is done!")
		}()
	} else {
		ctx.String(http.StatusOK, "Error while getting metrics:\n%v", err)
		return
	}
	ctx.String(http.StatusOK, "Anomaly detection triggered!")
}

// Runs "outliers.py" and wraps the output
func triggerIsolationForest(filename string, data system_metrics.SystemMetric) {
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
		gocsv.UnmarshalFile(inputFile, &anomalyData.Metrics)
		log.Println(anomalyData.Metrics[0].Cpu_System)
		//TODO wrap anomaly output

	}

}

func main() {
	router := gin.Default()
	router.GET("/nala/IF/:host/:duration", triggerDetection)
	router.Run("localhost:8088")
}
