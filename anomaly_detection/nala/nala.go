package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"

	"github.com/gin-gonic/gin"
)

type System struct {
	ID      string
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

// ! Delete later
type Message struct {
	Msg string `json:"msg"`
}

// TODO Create a trigger endpoint
func triggerDetection(ctx *gin.Context) {
	var message = Message{Msg: "Detection triggered"}
	ctx.IndentedJSON(http.StatusOK, message)
}

// Runs "testyp.py" and prints the output
func pyCall() {
	//Sets Arguments to the command
	cmd := exec.Command("python", "./testpy.py", "Hello Python")
	//executes command, listends to stdout, puts w/e into "out" var unless error
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err) // Only gives exit 1 if error, use "cmd.Stderr = os.Stderr" (import os)
	}
	//Print, Need explicit typing or it prints an array with unicode numbers
	fmt.Println(string(out))
}

func main() {
	pyCall()
	router := gin.Default()
	router.GET("/nala/trigger", triggerDetection)
	router.Run("localhost:8088")
}
