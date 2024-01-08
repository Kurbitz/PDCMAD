package main

import (
	"internal/influxdbapi"
	"internal/system_metrics"
	"log"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
)

// Fill the database with metrics from the specified files.
// The files are read in parallel and the metrics are written to the database in parallel making this function reasonably fast.
// The relative timestamps of the metrics will be translated to absolute timestamps based on the time parameters (gap and duration) but their relative order and time difference will be preserved.
// If the anomaly flag is set, an anomaly transformation will be applied to the metrics before they are written to the database.
// FIXME: The goroutines might return an error but the function will not return it, potentially causing silent errors.
func Fill(flags FillArgs) error {
	// Initialize the influxdb api
	var influxDBApi = influxdbapi.NewInfluxDBApi(flags.DBArgs.Token, flags.DBArgs.Host, flags.DBArgs.Port, flags.DBArgs.Org, flags.DBArgs.Bucket, flags.DBArgs.Measurement)
	defer influxDBApi.Close()

	log.Printf("Filling database with metrics from %v files\n", len(flags.Files))

	// Initialize the progress bar
	bar := progressbar.Default(int64(len(flags.Files)), "Processing files")

	// The wait group is used to wait for all goroutines to finish
	var wg sync.WaitGroup

	// For each file we create a goroutine that reads and parses the file, then writes the metrics to the database
	for _, file := range flags.Files {
		wg.Add(1)

		go func(filePath string, bar *progressbar.ProgressBar) error {
			defer wg.Done()

			// Get the id from the file name
			id := GetIdFromFileName(filePath)

			bar.Describe("Reading file " + filePath)

			// Read and parse the file
			// FIXME: Handle this error
			metric, _ := system_metrics.ReadFromFile(filePath, id)

			bar.Describe("Slicing metrics")

			// Modify the metrics slice based on the startat and duration parameters
			// If the parameters are 0, it will return all metrics, so we don't need to check for that
			metric.SliceBetween(flags.StartAt, flags.Duration)

			// Create a channel to send progress updates to the progress bar, this allows us to update the progress bar
			// when the metrics are being written to the database
			progressChan := make(chan int)
			defer close(progressChan)

			// Update the progress bar max value
			bar.ChangeMax(bar.GetMax() + len(metric.Metrics))
			// Add one to the progress bar to account for being done with the parsing the file
			bar.Add(1)

			// If the anomaly flag is set, inject an anomaly into the metrics
			if len(flags.Anomaly) > 0 {
				bar.Describe("Injecting anomaly")
				if err := InjectAnomaly(metric, flags.Anomaly); err != nil {
					return err
				}
			}

			// Start a goroutine that will update the progress bar when the metrics are being written to the database
			go func() {
				for range progressChan {
					bar.Add(1)
				}
			}()

			bar.Describe("Writing metrics to database")

			// Write the metrics to the database
			// The WriteMetrics function will call the callback function to update the progress bar every time a metric is written
			// FIXME: Handle this returning an error
			influxDBApi.WriteMetrics(*metric, flags.Gap, func() {
				progressChan <- 1
			})
			return nil

		}(file, bar)
	}
	// Wait for all goroutines to finish
	wg.Wait()
	bar.Finish()
	log.Println("Finished filling database")
	return nil
}

// Stream metrics one by one from the specified file to the database.
// The metrics are streamed in order and the time difference between them is preserved.
// The relative timestamps of the metrics will be translated to absolute timestamps based on the time parameters (gap and duration if set).
// If the append flag is set, the metrics will be appended to the existing metrics in the database, otherwise the metric will be inserted at the current time.
// The time multiplier flag can be used to speed up the streaming process.
// If the anomaly flag is set, an anomaly transformation will be applied to the metrics before they are written to the database.
func Stream(flags StreamArgs) error {
	// Initialize the influxdb api
	var influxDBApi = influxdbapi.NewInfluxDBApi(flags.DBArgs.Token, flags.DBArgs.Host, flags.DBArgs.Port, flags.DBArgs.Org, flags.DBArgs.Bucket, flags.DBArgs.Measurement)
	defer influxDBApi.Close()

	id := GetIdFromFileName(flags.File)

	// The time at which the first metric will be inserted defaults to the current time
	insertTime := time.Now()

	// If we are appending we need to get the last metric from the database and start from there
	// If we start from now make sure the timemultiplier is set to 1 so we don't exceed the current time
	if flags.Append {
		lastMetric, err := influxDBApi.GetLastMetric(id)
		if err != nil {
			return err
		}

		insertTime = time.Unix(lastMetric.Timestamp, 0)
	} else if flags.TimeMultiplier > 1 {
		log.Fatal("Timemultiplier can only be set while appending")
	}

	// Read and parse the file
	metrics, err := system_metrics.ReadFromFile(flags.File, id)
	if err != nil {
		return err
	}

	// Modify the metrics slice based on the startat and duration parameters
	metrics.SliceBetween(flags.Startat, flags.Duration)

	if len(flags.Anomaly) > 0 {
		if err := InjectAnomaly(metrics, flags.Anomaly); err != nil {
			return err
		}
	}

	// If we are appending we need to calculate the time delta between the first two metrics to know where to insert
	// the first metric.
	var timeDelta int64 = 0
	if flags.Append {
		if len(metrics.Metrics) < 2 {
			log.Println("Not enough metrics to calculate time delta, exiting...")
			return nil
		}
		timeDelta = (metrics.Metrics[1].Timestamp - metrics.Metrics[0].Timestamp)
		insertTime = insertTime.Add(time.Duration(timeDelta) * time.Second)
	}

	// Insert all metrics except the last one since we need to handle that one separately to avoid an out of bounds error
	for i, metric := range metrics.Metrics[:len(metrics.Metrics)-1] {
		// If the time multiplier is set, we might exceed the current wall time, so we need to check for that, otherwise
		// we might try to insert metrics with timestamps in the future which will cause an error
		if insertTime.After(time.Now()) {
			log.Println("You have exceeded the current time. The time multiplier might be too high, exiting...")
			return nil
		}

		// Write the metric to the database
		err := influxDBApi.WriteMetric(*metric, id, insertTime)
		if err != nil {
			return err
		}
		log.Printf("%v: metric written at %v\n", id, insertTime.Format(time.RFC3339))

		// Calculate the time delta between the current metric and the next one to get the next insert time
		timeDelta = (metrics.Metrics[i+1].Timestamp - metric.Timestamp)
		insertTime = insertTime.Add(time.Duration(timeDelta) * time.Second)

		// Sleep until the next metric should be inserted
		// The time multiplier can be used to speed up the streaming process
		// FIXME: Using time.Sleep is not very accurate and might cause drift over time, should not be a huge problem though
		// since the inser time should be completely accurate, but it might be worth looking into a better solution (maybe time.Ticker?)
		time.Sleep((time.Second * time.Duration(timeDelta)) / time.Duration(flags.TimeMultiplier))
	}
	// Handle the last metric
	influxDBApi.WriteMetric(*metrics.Metrics[len(metrics.Metrics)-1], id, insertTime)
	log.Printf("%v: metric written at %v\n", id, insertTime.Format(time.RFC3339))

	return nil
}

// Clean the database by deleting either all data in the bucket or all data for the specified hosts.
// The duration flag can be used to specify how far back to delete data.
// Returns an error if something goes wrong.
func Clean(flags CleanArgs) error {
	// Initialize the influxdb api
	var influxDBApi = influxdbapi.NewInfluxDBApi(flags.DBArgs.Token, flags.DBArgs.Host, flags.DBArgs.Port, flags.DBArgs.Org, flags.DBArgs.Bucket, flags.DBArgs.Measurement)
	defer influxDBApi.Close()

	// Clean the entire bucket if the all flag is set
	if flags.All {
		return influxDBApi.DeleteBucket(flags.Duration)
	}

	// Delete the data for each host in parallel
	// The wait group is used to wait for all goroutines to finish
	var wg sync.WaitGroup
	for _, host := range flags.Hosts {
		wg.Add(1)

		// Start a goroutine for each host
		go func(hostName string) {
			defer wg.Done()
			// FIXME: Handle this error
			influxDBApi.DeleteHost(hostName, flags.Duration)
		}(host)
	}
	wg.Wait()

	return nil
}
