package system_metrics

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

type AnomalyEvent struct {
	Timestamp int64  `csv:"timestamp"`
	Host      string `csv:"host"`
	Metric    string `csv:"metric"`
	Comment   string `csv:"comment"`
}

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

func (a AnomalyEvent) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"timestamp": a.Timestamp,
		"host":      a.Host,
		"metric":    a.Metric,
		"comment":   a.Comment,
	}
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
	lastTimestamp := time.Duration(sm.Metrics[len(sm.Metrics)-1].Timestamp) * time.Second
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
