package main

import (
	"pdc-mad/metrics"
)

func main() {
	fileName := "system-1.csv"
	metrics.ReadFromFile(fileName)

}
