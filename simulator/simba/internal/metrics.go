package simba

import (
	"log"
	"os"
	"time"

	"github.com/gocarina/gocsv"
)

type SystemMetric struct {
	Id      string
	Metrics []*Metric
}

type Metric struct {
	Timestamp               int64   `csv:"timestamp"`
	Load1m                  float64 `csv:"load-1m"`
	Load5m                  float64 `csv:"load-5m"`
	Load15m                 float64 `csv:"load-15m"`
	Sys_Mem_Swap_Total      int64   `csv:"sys-mem-swap-total"`
	Sys_Mem_Swap_Free       int64   `csv:"sys-mem-swap-free"`
	Sys_Mem_Free            int64   `csv:"sys-mem-free"`
	Sys_Mem_Cache           int64   `csv:"sys-mem-cache"`
	Sys_Mem_Buffered        int64   `csv:"sys-mem-buffered"`
	Sys_Mem_Available       int64   `csv:"sys-mem-available"`
	Sys_Mem_Total           int64   `csv:"sys-mem-total"`
	Sys_Fork_Rate           float64 `csv:"sys-fork-rate"`
	Sys_Interrupt_Rate      float64 `csv:"sys-interrupt-rate"`
	Sys_Context_Switch_Rate float64 `csv:"sys-context-switch-rate"`
	Sys_Thermal             float64 `csv:"sys-thermal"`
	Disk_Io_Time            float64 `csv:"disk-io-time"`
	Disk_Bytes_Read         float64 `csv:"disk-bytes-read"`
	Disk_Bytes_Written      float64 `csv:"disk-bytes-written"`
	Disk_Io_Read            float64 `csv:"disk-io-read"`
	Disk_Io_Write           float64 `csv:"disk-io-write"`
	Cpu_Io_Wait             float64 `csv:"cpu-iowait"`
	Cpu_System              float64 `csv:"cpu-system"`
	Cpu_User                float64 `csv:"cpu-user"`
	Server_Up               int64   `csv:"server-up"`
}

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

// SliceBetween returns a slice of metrics between the startAt time and the duration
// If duration is 0, it will return all metrics after the startAt time
func (sm *SystemMetric) SliceBetween(startAt, duration time.Duration) {

	startIndex := 0
	endIndex := len(sm.Metrics)

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
	lastTimestamp := time.Duration(sm.Metrics[len(sm.Metrics)-1].Timestamp)
	if duration == 0 || startAt+duration >= lastTimestamp {
		sm.Metrics = sm.Metrics[startIndex:]
		return
	}

	// Go from the startat time and duration forward
	for i, m := range sm.Metrics[startIndex:] {
		if time.Second*time.Duration(m.Timestamp) >= startAt+duration {
			endIndex = i
			break
		}
	}

	sm.Metrics = sm.Metrics[startIndex : startIndex+endIndex]
}

func ReadFromFile(filePath string, id string) (*SystemMetric, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	metrics := []*Metric{}
	if err := gocsv.UnmarshalFile(file, &metrics); err != nil {
		panic(err)
	}
	var systemMetrics = SystemMetric{Id: id, Metrics: metrics}

	return &systemMetrics, nil
}
