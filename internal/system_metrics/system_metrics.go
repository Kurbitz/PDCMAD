package system_metrics

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gocarina/gocsv"
)

// SystemMetric is a struct that contains the id of a system and a slice of metrics belonging to that system
// The id is used to identify the system in the database.
// A slice of metrics is used to store the metrics of the system.
// A zero value for SystemMetric is not valid or useful.
type SystemMetric struct {
	Id      string    // Id is the id of the system
	Metrics []*Metric // Metrics is a slice of metrics belonging to the system
}

// Metric is a struct that contains the metrics of a system at a specific time.
// The fields are the metrics that are collected from the system as specified in the dataset by Westermo.
// See: https://github.com/westermo/test-system-performance-dataset/ for more information.
// Tags are used to convert parse structs to and from csv and json.
// A zero value for Metric is not valid.
type Metric struct {
	Timestamp               int64   `csv:"timestamp" json:"timestamp"`
	Load1m                  float64 `csv:"load-1m" json:"load-1m"`
	Load5m                  float64 `csv:"load-5m" json:"load-5m"`
	Load15m                 float64 `csv:"load-15m" json:"load-15m"`
	Sys_Mem_Swap_Total      int64   `csv:"sys-mem-swap-total" json:"sys-mem-swap-total"`
	Sys_Mem_Swap_Free       int64   `csv:"sys-mem-swap-free" json:"sys-mem-swap-free"`
	Sys_Mem_Free            int64   `csv:"sys-mem-free" json:"sys-mem-free"`
	Sys_Mem_Cache           int64   `csv:"sys-mem-cache" json:"sys-mem-cache"`
	Sys_Mem_Buffered        int64   `csv:"sys-mem-buffered" json:"sys-mem-buffered"`
	Sys_Mem_Available       int64   `csv:"sys-mem-available" json:"sys-mem-available"`
	Sys_Mem_Total           int64   `csv:"sys-mem-total" json:"sys-mem-total"`
	Sys_Fork_Rate           float64 `csv:"sys-fork-rate" json:"sys-fork-rate"`
	Sys_Interrupt_Rate      float64 `csv:"sys-interrupt-rate" json:"sys-interrupt-rate"`
	Sys_Context_Switch_Rate float64 `csv:"sys-context-switch-rate" json:"sys-context-switch-rate"`
	Sys_Thermal             float64 `csv:"sys-thermal" json:"sys-thermal"`
	Disk_Io_Time            float64 `csv:"disk-io-time" json:"disk-io-time"`
	Disk_Bytes_Read         float64 `csv:"disk-bytes-read" json:"disk-bytes-read"`
	Disk_Bytes_Written      float64 `csv:"disk-bytes-written" json:"disk-bytes-written"`
	Disk_Io_Read            float64 `csv:"disk-io-read" json:"disk-io-read"`
	Disk_Io_Write           float64 `csv:"disk-io-write" json:"disk-io-write"`
	Cpu_Io_Wait             float64 `csv:"cpu-iowait" json:"cpu-iowait"`
	Cpu_System              float64 `csv:"cpu-system" json:"cpu-system"`
	Cpu_User                float64 `csv:"cpu-user" json:"cpu-user"`
	Server_Up               int64   `csv:"server-up" json:"server-up"`
}

// AnomalyEvent is a struct that contains the information about an anomaly event.
// This is intended to be used to log the anomalies to a file in a generic way since the anomaly detection algorithms
// might have different information about the anomaly. Providing fields for each of the metrics would be cumbersome and
// potentially not possible if the anomaly detection algorithm does not have information about all the metrics.
// This struct is not stored in the database because the string fields are not easily visualized in Grafana.
// Tags are used to convert parse structs to and from csv.
// A zero value for AnomalyEvent is not valid.
type AnomalyEvent struct {
	Timestamp int64  `csv:"timestamp"` // Timestamp is the time when the anomaly occurred
	Host      string `csv:"host"`      // Host is the id of the system where the anomaly occurred
	Metric    string `csv:"metric"`    // Metric is the metric that triggered the anomaly (if applicable)
	Comment   string `csv:"comment"`   // Comment is a comment about the anomaly (currently used to identify the algorithm that detected the anomaly)
}

// AnomalyDetectionOutput is a struct that contains every metric (same fields as Metric) and whether or not it is an anomaly.
// This is intended to be the output of the anomaly detection algorithms and is what is used to store the anomalies
// in the database. It is not as generic as AnomalyEvent but we found that it was easier to work with.
// This struct is stored in the database because it contains the same fields as Metric and is easily visualized in Grafana.
// Tags are used to convert parse structs to and from csv.
// A zero value for AnomalyDetectionOutput is not valid.
type AnomalyDetectionOutput struct {
	Timestamp               int64 `csv:"timestamp"`
	Load1m                  bool  `csv:"load-1m"`
	Load5m                  bool  `csv:"load-5m"`
	Load15m                 bool  `csv:"load-15m"`
	Sys_Mem_Swap_Total      bool  `csv:"sys-mem-swap-total"`
	Sys_Mem_Swap_Free       bool  `csv:"sys-mem-swap-free"`
	Sys_Mem_Free            bool  `csv:"sys-mem-free"`
	Sys_Mem_Cache           bool  `csv:"sys-mem-cache"`
	Sys_Mem_Buffered        bool  `csv:"sys-mem-buffered"`
	Sys_Mem_Available       bool  `csv:"sys-mem-available"`
	Sys_Mem_Total           bool  `csv:"sys-mem-total"`
	Sys_Fork_Rate           bool  `csv:"sys-fork-rate"`
	Sys_Interrupt_Rate      bool  `csv:"sys-interrupt-rate"`
	Sys_Context_Switch_Rate bool  `csv:"sys-context-switch-rate"`
	Sys_Thermal             bool  `csv:"sys-thermal"`
	Disk_Io_Time            bool  `csv:"disk-io-time"`
	Disk_Bytes_Read         bool  `csv:"disk-bytes-read"`
	Disk_Bytes_Written      bool  `csv:"disk-bytes-written"`
	Disk_Io_Read            bool  `csv:"disk-io-read"`
	Disk_Io_Write           bool  `csv:"disk-io-write"`
	Cpu_Io_Wait             bool  `csv:"cpu-iowait"`
	Cpu_System              bool  `csv:"cpu-system"`
	Cpu_User                bool  `csv:"cpu-user"`
	Server_Up               bool  `csv:"server-up"`
}

// The ToMap functions are used to convert structs to maps.
// They need to be implemented for every struct that is used to store data in the database.
// This is because the influxdb api requires a map to write to the database.
// They could probably be implemented in a more generic way using reflections but this works for now.

// ToMap converts a Metric to a map[string]interface{}.
func (a AnomalyEvent) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"timestamp": a.Timestamp,
		"host":      a.Host,
		"metric":    a.Metric,
		"comment":   a.Comment,
	}
}

// ToMap converts a Metric to a map[string]interface{}.
func (m Metric) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"timestamp":               m.Timestamp,
		"load-1m":                 m.Load1m,
		"load-5m":                 m.Load5m,
		"load-15m":                m.Load15m,
		"sys-mem-swap-total":      m.Sys_Mem_Swap_Total,
		"sys-mem-swap-free":       m.Sys_Mem_Swap_Free,
		"sys-mem-free":            m.Sys_Mem_Free,
		"sys-mem-cache":           m.Sys_Mem_Cache,
		"sys-mem-buffered":        m.Sys_Mem_Buffered,
		"sys-mem-available":       m.Sys_Mem_Available,
		"sys-mem-total":           m.Sys_Mem_Total,
		"sys-fork-rate":           m.Sys_Fork_Rate,
		"sys-interrupt-rate":      m.Sys_Interrupt_Rate,
		"sys-context-switch-rate": m.Sys_Context_Switch_Rate,
		"sys-thermal":             m.Sys_Thermal,
		"disk-io-time":            m.Disk_Io_Time,
		"disk-bytes-read":         m.Disk_Bytes_Read,
		"disk-bytes-written":      m.Disk_Bytes_Written,
		"disk-io-read":            m.Disk_Io_Read,
		"disk-io-write":           m.Disk_Io_Write,
		"cpu-iowait":              m.Cpu_Io_Wait,
		"cpu-system":              m.Cpu_System,
		"cpu-user":                m.Cpu_User,
		"server-up":               m.Server_Up,
	}
}

// ToMap converts a AnomalyDetectionOutput to a map[string]interface{}.
func (am AnomalyDetectionOutput) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"timestamp":               am.Timestamp,
		"load-1m":                 am.Load1m,
		"load-5m":                 am.Load5m,
		"load-15m":                am.Load15m,
		"sys-mem-swap-total":      am.Sys_Mem_Swap_Total,
		"sys-mem-swap-free":       am.Sys_Mem_Swap_Free,
		"sys-mem-free":            am.Sys_Mem_Free,
		"sys-mem-cache":           am.Sys_Mem_Cache,
		"sys-mem-buffered":        am.Sys_Mem_Buffered,
		"sys-mem-available":       am.Sys_Mem_Available,
		"sys-mem-total":           am.Sys_Mem_Total,
		"sys-fork-rate":           am.Sys_Fork_Rate,
		"sys-interrupt-rate":      am.Sys_Interrupt_Rate,
		"sys-context-switch-rate": am.Sys_Context_Switch_Rate,
		"sys-thermal":             am.Sys_Thermal,
		"disk-io-time":            am.Disk_Io_Time,
		"disk-bytes-read":         am.Disk_Bytes_Read,
		"disk-bytes-written":      am.Disk_Bytes_Written,
		"disk-io-read":            am.Disk_Io_Read,
		"disk-io-write":           am.Disk_Io_Write,
		"cpu-iowait":              am.Cpu_Io_Wait,
		"cpu-system":              am.Cpu_System,
		"cpu-user":                am.Cpu_User,
		"server-up":               am.Server_Up,
	}
}

// SliceBetween slices the metrics between the startAt time and the duration.
// The startAt time specifies how far into the metric file we should start the slice.
// The duration specifies how long the slice should be.
// If duration is 0, it will return all metrics after the startAt time.
// Will modify the metrics slice in place.
func (sm *SystemMetric) SliceBetween(startAt, duration time.Duration) {
	// Find the first and last index of the slice
	startIndex := 0
	endIndex := len(sm.Metrics)

	// Check if the duration exceeds the length of the metric file
	if time.Duration(time.Duration.Seconds(duration+startAt)) > time.Duration(sm.Metrics[len(sm.Metrics)-1].Timestamp) {
		log.Fatal("Duration exceeds length of the metric file")
	}

	// Find the first metric that is after the startAt time
	for i, m := range sm.Metrics {
		if time.Second*time.Duration(m.Timestamp) >= startAt {
			startIndex = i
			break
		}
	}

	// If duration is 0 or the duration is longer than the last metric, return the all metrics after the startAt time
	lastTimestamp := time.Duration(sm.Metrics[len(sm.Metrics)-1].Timestamp) * time.Second
	if duration == 0 || startAt+duration >= lastTimestamp {
		sm.Metrics = sm.Metrics[startIndex:]
		return
	}

	// The last metric will be duration time after the startAt time
	for i, m := range sm.Metrics[startIndex:] {
		if time.Second*time.Duration(m.Timestamp) >= startAt+duration {
			endIndex = i
			break
		}
	}

	// Slice the metrics between the start and end index
	sm.Metrics = sm.Metrics[startIndex : startIndex+endIndex]
}

// ReadFromFile reads a CSV file of metrics and returns a SystemMetric struct.
// The CSV file should have the same format as the dataset provided by Westermo.
// Returns an error if something fails.
func ReadFromFile(filePath string, id string) (*SystemMetric, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Unmarshal the file into a slice of metrics
	metrics := []*Metric{}
	if err := gocsv.UnmarshalFile(file, &metrics); err != nil {
		panic(err)
	}

	// Create a SystemMetric struct and add the id and metrics
	var systemMetrics = SystemMetric{Id: id, Metrics: metrics}

	return &systemMetrics, nil
}

// ParseAnomalyDetectionOutputCSV parses a CSV file of AnomalyDetectionOutput structs and returns a slice of AnomalyDetectionOutput structs.
// The CSV file should have the same format as the AnomalyDetectionOutput struct.
// Returns an error if something fails.
func ParseAnomalyDetectionOutputCSV(filename, host string) (*[]AnomalyDetectionOutput, error) {
	anomalyData := []AnomalyDetectionOutput{}
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
