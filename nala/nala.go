package main

import (
	"internal/influxdbapi"
	"internal/system_metrics"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/gocarina/gocsv"

	"github.com/gin-gonic/gin"
)

var inProgress = false

func triggerDetection(ctx *gin.Context) {
	//TODO change to environment variables
	dbapi := influxdbapi.NewInfluxDBApi(os.Getenv("INFLUXDB_TOKEN"), os.Getenv("INFLUXDB_HOST"), os.Getenv("INFLUXDB_PORT"))
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
		if err = logAnomalies("/tmp/anomalies.csv", anomalies); err != nil {
			log.Printf("Error when writing anomalies to file: %v\n", err)
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

/*
Takes AnomalyMetric struct and writes it to a log file
Logfile output: [time, host, metric, comment]
Returns error if something fails
*/
func logAnomalies(filePath string, data []system_metrics.Anomaly) error {
	outputFile, err := os.Create(filePath)
	if err != nil {
		log.Printf("Error when creating file: %v", err)
		return err
	}
	defer outputFile.Close()
	err = gocsv.MarshalFile(&data, outputFile)
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
