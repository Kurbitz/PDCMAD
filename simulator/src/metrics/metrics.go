package metrics

import (
	"errors"
	"os"

	"github.com/gocarina/gocsv"
)

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

func ReadFromFile(filePath string) ([]*Metric, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, errors.New("file not found")
	}
	defer file.Close()

	metrics := []*Metric{}

	if err := gocsv.UnmarshalFile(file, &metrics); err != nil {
		panic(err)
	}

	for _, metric := range metrics {
		println(metric.Timestamp, metric.Load15m)
	}

	return metrics, nil
}
