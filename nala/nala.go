package main

import (
	"net/http"

	"internal/influxdbapi"

	"github.com/gin-gonic/gin"
)

// TODO Create a trigger endpoint
func triggerDetection(ctx *gin.Context) {
	dbapi := influxdbapi.NewInfluxDBApi("KBntTYJdaWbknRyM-CAw29iYdJmQkiK6C1vlEO3B5yuvgGJlmG4Gasps5rTRGflLq7bRSSWZSA_zdnYhpu-HXQ==", "localhost", "8086")
	defer dbapi.Close()
	host := ctx.Param("host")
	duration := ctx.Param("duration")
	message := dbapi.GetMetrics(host, duration)
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
	router.GET("/nala/trigger/:host/:duration", triggerDetection)
	router.Run("localhost:8088")
}
