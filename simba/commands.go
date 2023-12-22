package main

import (
	"internal/influxdbapi"
	"internal/system_metrics"
	"log"
	"sync"
	"time"
)

func Fill(flags FillArgs) error {

	var influxDBApi = influxdbapi.NewInfluxDBApi(flags.DBToken, flags.DBIp, flags.DBPort, flags.DBOrg, flags.DBBucket, flags.DBMeasurement)
	defer influxDBApi.Close()

	var wg sync.WaitGroup
	for _, file := range flags.Files {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()

			id := GetIdFromFileName(filePath)
			metric, _ := system_metrics.ReadFromFile(filePath, id)

			// Slice the metric between startAt and duration
			// If the parameters are 0, it will return all metrics, so we don't need to check for that
			metric.SliceBetween(flags.StartAt, flags.Duration)
			influxDBApi.WriteMetrics(*metric, flags.Gap)
		}(file)

	}
	wg.Wait()

	return nil
}

func Stream(flags StreamArgs) error {
	var influxDBApi = influxdbapi.NewInfluxDBApi(flags.DBToken, flags.DBIp, flags.DBPort, flags.DBOrg, flags.DBBucket, flags.DBMeasurement)
	id := GetIdFromFileName(flags.File)

	insertTime := time.Now()

	// If append is set we need to get the last metric and start from there
	// else we start from now
	// If we start from now make sure the timemultiplier is set to 1
	if flags.Append {
		lastMetric, err := influxDBApi.GetLastMetric(id)
		if err != nil {
			return err
		}

		insertTime = time.Unix(lastMetric.Timestamp, 0)
	} else if flags.TimeMultiplier > 1 {
		log.Fatal("Timemultiplier can only be set while appending")
	}

	metrics, err := system_metrics.ReadFromFile(flags.File, id)
	if err != nil {
		return err
	}
	metrics.SliceBetween(flags.Startat, flags.Duration)

	// If we are appending we need to calculate the time delta between the first two metrics
	var timeDelta int64 = 0
	if flags.Append {
		timeDelta = (metrics.Metrics[1].Timestamp - metrics.Metrics[0].Timestamp)
		insertTime = insertTime.Add(time.Duration(timeDelta) * time.Second)
	}
	// Insert all metrics except the last one
	for i, metric := range metrics.Metrics[:len(metrics.Metrics)-1] {
		if insertTime.After(time.Now()) {
			log.Println("You have exceeded the current time. The time multiplier might be too high, exiting...")
			return nil
		}
		influxDBApi.WriteMetric(*metric, id, insertTime)
		log.Println("Inserted metric at", insertTime)

		timeDelta = (metrics.Metrics[i+1].Timestamp - metric.Timestamp)
		insertTime = insertTime.Add(time.Duration(timeDelta) * time.Second)

		time.Sleep((time.Second * time.Duration(timeDelta)) / time.Duration(flags.TimeMultiplier))
	}
	// Handle the last metric
	influxDBApi.WriteMetric(*metrics.Metrics[len(metrics.Metrics)-1], id, insertTime)
	log.Println("Inserted metrics at", insertTime)

	return nil
}

func Clean(flags CleanArgs) error {
	var influxDBApi = influxdbapi.NewInfluxDBApi(flags.DBToken, flags.DBIp, flags.DBPort, flags.DBOrg, flags.DBBucket, flags.DBMeasurement)
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
