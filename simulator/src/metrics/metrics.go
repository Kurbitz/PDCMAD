package metrics

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Metric struct {
	timestamp               int64
	load1m                  float64
	load5m                  float64
	load15m                 float64
	sys_mem_swap_total      int64
	sys_mem_swap_free       int64
	sys_mem_free            int64
	sys_mem_cache           int64
	sys_mem_buffered        int64
	sys_mem_available       int64
	sys_mem_total           int64
	sys_fork_rate           float64
	sys_interrupt_rate      float64
	sys_context_switch_rate float64
	sys_thermal             float64
	disk_io_time            float64
	disk_bytes_read         float64
	disk_bytes_written      float64
	disk_io_read            float64
	disk_io_write           float64
	cpu_io_wait             float64
	cpu_system              float64
	cpu_user                float64
	server_up               int64
}

func Test() string {
	return "hello"
}

func ReadFromFile(fileName string) {
	filePath := "../../../dataset/" + fileName
	println(filePath)
	file, err := os.Open(filePath)
	if err != nil {
		println("Something went wrong reading file: %s", fileName)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var serverMetrics []Metric
	scanner.Scan()

	//This should handle all the headers
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), ",")
		var metric Metric
		var metricError [24]error
		metric.timestamp, metricError[0] = strconv.ParseInt(line[0], 10, 64)
		metric.load1m, metricError[1] = strconv.ParseFloat(line[1], 10)
		metric.load5m, metricError[2] = strconv.ParseFloat(line[2], 10)
		metric.load15m, metricError[3] = strconv.ParseFloat(line[3], 10)
		metric.sys_mem_swap_total, metricError[4] = strconv.ParseInt(line[4], 10, 64)
		metric.sys_mem_swap_free, metricError[5] = strconv.ParseInt(line[5], 10, 64)
		metric.sys_mem_free, metricError[6] = strconv.ParseInt(line[6], 10, 64)
		metric.sys_mem_cache, metricError[7] = strconv.ParseInt(line[7], 10, 64)
		metric.sys_mem_buffered, metricError[8] = strconv.ParseInt(line[8], 10, 64)
		metric.sys_mem_available, metricError[9] = strconv.ParseInt(line[9], 10, 64)
		metric.sys_mem_total, metricError[10] = strconv.ParseInt(line[10], 10, 64)
		metric.sys_fork_rate, metricError[11] = strconv.ParseFloat(line[11], 10)
		metric.sys_interrupt_rate, metricError[12] = strconv.ParseFloat(line[12], 10)
		metric.sys_context_switch_rate, metricError[13] = strconv.ParseFloat(line[13], 10)
		metric.sys_thermal, metricError[14] = strconv.ParseFloat(line[14], 10)
		metric.disk_io_time, metricError[15] = strconv.ParseFloat(line[15], 10)
		metric.disk_bytes_read, metricError[16] = strconv.ParseFloat(line[16], 10)
		metric.disk_bytes_written, metricError[17] = strconv.ParseFloat(line[17], 10)
		metric.disk_io_read, metricError[18] = strconv.ParseFloat(line[18], 10)
		metric.disk_io_write, metricError[19] = strconv.ParseFloat(line[19], 10)
		metric.cpu_io_wait, metricError[20] = strconv.ParseFloat(line[20], 10)
		metric.cpu_system, metricError[21] = strconv.ParseFloat(line[21], 10)
		metric.cpu_user, metricError[22] = strconv.ParseFloat(line[22], 10)
		metric.server_up, metricError[23] = strconv.ParseInt(line[23], 10, 64)
		for i, e := range metricError {
			if e != nil {
				println("Error occured on metric %d", i)
			}
		}
		serverMetrics = append(serverMetrics, metric)
	}
}
