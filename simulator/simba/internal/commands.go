package simba

import (
	"log"
	"sync"
	"time"
)

func Fill(flags FillFlags) error {

	var influxDBApi = NewInfluxDBApi(flags.DBToken, flags.DBIp, flags.DBPort)
	defer influxDBApi.Close()

	var wg sync.WaitGroup
	for _, file := range flags.Files {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()

			id := GetIdFromFileName(filePath)
			metric, _ := ReadFromFile(filePath, id)

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
	var influxDBApi = NewInfluxDBApi(flags.DBToken, flags.DBIp, flags.DBPort)
	id := GetIdFromFileName(flags.File)

	insertTime := time.Now()

	// If append is set we need to get the last metric and start from there
	// else we start from now
	if flags.Append {
		lastMetric, err := influxDBApi.GetLastMetric(id)
		if err != nil {
			return err
		}

		insertTime = time.Unix(lastMetric.Timestamp, 0)
	} else if flags.TimeMultiplier > 1 {
		log.Fatal("Timemultiplier can only be set while appending\n")
	}

	metrics, err := ReadFromFile(flags.File, id)
	if err != nil {
		return err
	}
	metrics.SliceBetween(flags.Startat, flags.Duration)

	// Insert all metrics except the last one
	for i, metric := range metrics.Metrics[:len(metrics.Metrics)-1] {
		if insertTime.After(time.Now()) {
			log.Println("You have exceeded the current time. The time multiplier might be too high, exiting...")
			return nil
		}
		influxDBApi.WriteMetric(*metric, id, insertTime)
		log.Println("Inserted metric at", insertTime)
		println(i)

		timeDelta := (metrics.Metrics[i+1].Timestamp - metric.Timestamp)
		insertTime = insertTime.Add(time.Duration(timeDelta) * time.Second)

		time.Sleep((time.Second * time.Duration(timeDelta)) / time.Duration(flags.TimeMultiplier))
	}
	// Handle the last metric
	influxDBApi.WriteMetric(*(metrics.Metrics[len(metrics.Metrics)-1]), id, insertTime)
	log.Println("Inserted metrics at", insertTime)
	println(metrics.Metrics[len(metrics.Metrics)-1].Cpu_Io_Wait)

	return nil
}
