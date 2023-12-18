package simba

import (
	"path/filepath"
	"sync"
)

func Fill(flags FillArgs) error {

	var influxDBApi = NewInfluxDBApi(flags.DBToken, flags.DBIp, flags.DBPort)
	defer influxDBApi.Close()

	var wg sync.WaitGroup
	for _, file := range flags.Files {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()
			// Remove the .csv from the file name
			// FIXME: Use better ID
			id := filepath.Base(filePath)[:len(filepath.Base(filePath))-len(filepath.Ext(filePath))]
			metric, _ := ReadFromFile(filePath, id)

			// Slice the metric between startAt and duration
			// If the parameters are 0, it will return all metrics, so we don't need to check for that
			metric.SliceBetween(flags.StartAt, flags.Duration)
			influxDBApi.WriteMetrics(*metric, flags.Gap, flags.Anomaly)
		}(file)

	}
	wg.Wait()

	return nil
}

func Stream(flags StreamArgs) error {

	return nil
}

func Clean(flags CleanArgs) error {
	var influxDBApi = NewInfluxDBApi(flags.DBToken, flags.DBIp, flags.DBPort)
	defer influxDBApi.Close()

	if flags.All { // Clean the bucket
		return influxDBApi.DeleteBucket(flags.Startat)
	}

	var wg sync.WaitGroup
	for _, host := range flags.Hosts {
		wg.Add(1)
		go func(hostName string) {
			defer wg.Done()
			influxDBApi.DeleteHost(hostName, flags.Startat)
		}(host)
	}
	wg.Wait()

	return nil
}
