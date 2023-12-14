package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"

	"github.com/gin-gonic/gin"
)

// TODO Add the actual metrics
type metric struct {
	ID   string  `json:"id"`
	Host string  `json:"host"`
	Data float64 `json:"data"`
}

var metrics = []metric{
	{ID: "1", Host: "server9000", Data: 123.231341},
	{ID: "2", Host: "server9001", Data: 0},
	{ID: "3", Host: "server9002", Data: 124.231341},
}

func getMetrics(ctx *gin.Context) {
	ctx.IndentedJSON(http.StatusOK, metrics)
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
	router.GET("/metrics", getMetrics)
	router.Run("localhost:8088")
}
