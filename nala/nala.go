package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
)

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
func pythonSmokeTest() {

	log.Println("Running python smoke test...")
	cmd := exec.Command("python", "./testpy.py", "Python is working!")
	//executes command, listends to stdout, puts w/e into "out" var unless error
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err) // Only gives exit 1 if error, use "cmd.Stderr = os.Stderr" (import os)
	}
	//Print, Need explicit typing or it prints an array with unicode numbers
	log.Print(string(out))
	log.Println("Python smoke test complete!")
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

	// Test-write a log message to /var/log/nala.log

	router := gin.Default()

	router.GET("/nala/trigger", triggerDetection)
	router.Run("0.0.0.0:8088")
}
