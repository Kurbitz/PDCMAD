package metrics

import (
	"bufio"
	"fmt"
	"os"
)

type Metric struct {
	timestamp               string //should it be a string?
	load1m                  float32
	load5m                  float32
	load15m                 float32
	sys_mem_swap_total      int64
	sys_mem_swap_free       int64
	sys_mem_free            int64
	sys_mem_cache           int64
	sys_mem_buffered        int64
	sys_mem_available       int64
	sys_mem_total           int64
	sys_fork_rate           float32
	sys_interrupt_rate      float32
	sys_context_switch_rate float32
	sys_thermal             float32
	disk_io_time            float32
	disk_bytes_read         float32
	disk_bytes_written      float32
	disk_io_read            float32
	disk_io_write           float32
	cpu_io_wait             float32
	cpu_system              float32
	cpu_user                float32
	server_up               int8
}

func Test() string {
	return "hello"
}

func ReadFromFile(fileName string) {
	filePath := "C:/Users/oscar/OneDrive/Dokument/GitHub/Performance_Data_Collector/dataset/" + fileName
	println(filePath)
	file, err := os.Open(filePath)
	if err != nil {
		println("Something went wrong reading file: %s", fileName)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		str := scanner.Text()
		fmt.Println(str)
	}

}
