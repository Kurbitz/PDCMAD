package main

import (
	"pdc-mad/metrics"
)

func main() {
	fileName := "../../../dataset/system-1.csv"
	metrics.ReadFromFile(fileName)

}
