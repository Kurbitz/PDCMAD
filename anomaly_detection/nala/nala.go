package main

import (
	// "fmt"
	"net/http"

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

func main() {
	router := gin.Default()
	router.GET("/metrics", getMetrics)
	router.Run("localhost:8088")
}
