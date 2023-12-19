package main

import (
	"fmt"
	"log"
	"net/http"
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
	router.Run("0.0.0.0:8088")
}
