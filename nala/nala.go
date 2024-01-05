package main

import (
	"internal/influxdbapi"
	"internal/system_metrics"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"reflect"

	"github.com/gocarina/gocsv"

	"github.com/gin-gonic/gin"
)

var inProgress = false

func triggerDetection(ctx *gin.Context) {
	//TODO change to environment variables
	dbapi := influxdbapi.NewInfluxDBApi(os.Getenv("INFLUXDB_TOKEN"), os.Getenv("INFLUXDB_HOST"), os.Getenv("INFLUXDB_PORT"), os.Getenv("INFLUXDB_ORG"), os.Getenv("INFLUXDB_BUCKET"), "anomalies")
	defer dbapi.Close()
	algorithm := ctx.Param("algorithm")
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
	if inProgress {
		ctx.String(http.StatusOK, "Anomaly detection is already in progress")
		return
	}
	callable, exists := supportedAlgorithms[algorithm]
	if !exists {
		ctx.String(http.StatusOK, "Algorithm %v is not supported", algorithm)
		return
	}
	inProgress = true
	detection, err := NewAnomalyDetection(dbapi, host, duration)
	if err != nil {
		ctx.String(http.StatusOK, "%v", err)
		inProgress = false
		return
	}

	go func() {
		defer func() {
			inProgress = false
		}()

		anomalies, err := callable(detection)
		if err != nil {
			log.Printf("Anomaly detection failed with: %v\n", err)
			return
		}
		if err = logAnomalies("/tmp/anomalies.csv", host, *anomalies); err != nil {
			log.Printf("Error when writing anomalies to file: %v\n", err)
			return
		}
		if err = dbapi.WriteAnomalies(*anomalies, host); err != nil {
			log.Printf("Error when writing anomalies to influxdb: %v\n", err)
			return
		}
		log.Println("Anomaly detection is done!")
	}()

	ctx.String(http.StatusOK, "Anomaly detection triggered!\n")
}

// Runs "testyp.py" and prints the output
func pythonSmokeTest() {

	log.Println("Running python smoke test...")
	cmd := exec.Command("python", "./testpy.py", "Python is working!")

	//executes command, listends to stdout, puts w/e into "out" var unless error
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	//Print, Need explicit typing or it prints an array with unicode numbers
	log.Print(string(out))
	log.Println("Python smoke test complete!")
}

/*
Takes AnomalyMetric struct and writes it to a log file
Logfile output: [time, host, metric, comment]
Returns error if something fails
*/
func logAnomalies(filePath string, host string, data []system_metrics.AnomalyDetectionOutput) error {
	outputArray := []system_metrics.AnomalyEvent{}
	for _, v := range data {
		r := reflect.ValueOf(v)
		for i := 1; i < r.NumField(); i++ {
			if r.Field(i).Interface() == true {
				outputArray = append(outputArray, system_metrics.AnomalyEvent{Timestamp: v.Timestamp, Host: host, Metric: r.Type().Field(i).Tag.Get("csv"), Comment: "Isolation forest"})
			}
		}
	}

	outputFile, err := os.Create(filePath)
	if err != nil {
		log.Printf("Error when creating file: %v", err)
		return err
	}
	defer outputFile.Close()
	err = gocsv.MarshalFile(&outputArray, outputFile)
	if err != nil {
		log.Printf("Error while parsing metrics from file: %v", err)
		return err
	}
	return nil
}

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

func main() {
	f, err := os.OpenFile("/var/log/nala.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	mw := io.MultiWriter(os.Stdout, f)

	log.SetOutput(mw)

	log.Println("Starting Nala...")
	pythonSmokeTest()
	checkEnv()

	router := gin.Default()

	router.GET("/nala/:algorithm/:host/:duration", triggerDetection)

	router.GET("/nala/test", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Nala is working!")
	})

	router.Run("0.0.0.0:8088")
}
