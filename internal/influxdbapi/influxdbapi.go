package influxdbapi

import (
	"context"
	"encoding/json"
	"fmt"
	"internal/system_metrics"
	"log"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

// InfluxDBApi is a struct containing the InfluxDB client and the
// information needed to write to the database
// A zero value of this struct is not usable. Use NewInfluxDBApi() to create a new instance.
type InfluxDBApi struct {
	influxdb2.Client        // The influxdb2 client from the influxdb-client-go library
	Org              string // The name of the organization in InfluxDB
	Bucket           string // The name of the bucket in InfluxDB
	Measurement      string // The name of the measurement in InfluxDB
}

// Creates a new InfluxDBApi struct
func NewInfluxDBApi(token, host, port, org, bucket, measurement string) InfluxDBApi {
	return InfluxDBApi{
		influxdb2.NewClient("http://"+host+":"+port, token),
		org,
		bucket,
		measurement,
	}
}

// GetLastMetric gets the last metric for the given host from InfluxDB.
// Returns the metric as a Metric struct. If any error occurs, it returns nil and the error.
func (api InfluxDBApi) GetLastMetric(host string) (*system_metrics.Metric, error) {
	// Set up the query
	q := api.QueryAPI(api.Org)
	query := fmt.Sprintf("from(bucket:\"%v\") |> range(start: -30d) |> filter(fn: (r) => r._measurement == \"%v\") |> filter(fn: (r) => r.host == \"%v\")|> last()", api.Bucket, api.Measurement, host)

	// Execute the query
	result, err := q.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}

	// Create a map to store the results
	results := make(map[string]interface{}, 0)

	// Iterate over the results and store them in the map
	// Next() returns false when there are no more results
	for result.Next() {
		results[result.Record().Field()] = result.Record().Value()
	}

	// Convert the map to JSON so it can be unmarshalled into a struct
	j, err := json.Marshal(results)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON into a struct
	metric := system_metrics.Metric{}
	if err := json.Unmarshal(j, &metric); err != nil {
		return nil, err
	}

	return &metric, nil
}

// WriteMetrics writes the given system metrics to InfluxDB asynchronously.
// It takes the metrics to be written, the time gap between the newest metric and the end time,
// and a callback function to be executed after each metric is written.
// Returns an error if any error occurs during the writing process.
// This function will mutate the timestamps of the metrics to match the time they were written.
func (api InfluxDBApi) WriteMetrics(metrics system_metrics.SystemMetric, gap time.Duration, onWrite func()) error {
	// Create a non-blocking write client
	writeAPI := api.WriteAPI(api.Org, api.Bucket)

	// To get the correct timestamp, we need to calculate the time of the first metric to be written.
	// This needs to take the time gap into account to leave a gap between the last metric and now.
	// To do this, we take the current time and subtract gap from it leaving us with the time of the last metric to be written.
	// We then subtract the timestamp of the last metric in the slice from the time of the last metric to be written.
	// This leaves us with the time of the first metric to be written.
	now := time.Now()
	end := now.Add(-gap)
	then := end.Add(time.Second * time.Duration(-metrics.Metrics[len(metrics.Metrics)-1].Timestamp))

	// Iterate over the metrics and write them to InfluxDB
	for _, x := range metrics.Metrics {
		// Calculate the timestamp of the metricTime metric
		metricTime := then.Add(time.Second * time.Duration(x.Timestamp))
		// Convert the timestamp to unix time
		x.Timestamp = metricTime.Unix()

		// Create a new point and write it to InfluxDB
		// The host is stored as a tag instead of a field to make it easier to filter the data
		p := influxdb2.NewPoint(api.Measurement, map[string]string{"host": metrics.Id}, x.ToMap(), metricTime)
		writeAPI.WritePoint(p)

		// Execute the callback function (usually used to update the progress bar)
		onWrite()
	}

	// Write any remaining points
	writeAPI.Flush()
	// FIXME: Handle errors that might occur when writing asynchronously
	return nil
}

// GetMetrics gets the metrics for the given host from InfluxDB within the given time.
// The duration must be in the InfluxDB format (e.g. 1h, 30m, 1d).
// The duration specifies how far back in time from now the metrics should be fetched.
// Returns the metrics as a SystemMetric struct. If any error occurs, it returns an empty SystemMetric struct and the error.
func (api InfluxDBApi) GetMetrics(host, duration string) (system_metrics.SystemMetric, error) {
	// Set up the query
	queryAPI := api.QueryAPI(api.Org)

	// The query does a pivot to get the data in a format that is easier to work with
	// Instead of having a row for each metric, we have a row for each timestamp
	// From: _time, _measurement, _field, _value
	// To: _time, cpu, memory, disk, network etc.
	// If we don't do this, we would have to iterate over the result and group the metrics by timestamp ourselves
	// To visualize this you can open the InfluxDB UI and run the query with and without the pivot
	query := fmt.Sprintf(`from(bucket: "%v") |> range(start: -%v) |> filter(fn: (r) => r._measurement == "%v") |> filter(fn: (r) => r["host"] == "%v") |> pivot(rowKey: ["_time"],columnKey: ["_field"], valueColumn: "_value")`, api.Bucket, duration, api.Measurement, host)

	// Execute the query
	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		log.Printf("Error performing query on bucket '%s': %s\n", api.Bucket, err)
		return system_metrics.SystemMetric{}, err
	}
	if result == nil {
		log.Printf("Duration '%v' on bucket '%v' is empty", duration, api.Bucket)
		return system_metrics.SystemMetric{}, fmt.Errorf("Empty query result")
	}

	// Create a slice to store the metrics
	var metrics []map[string]interface{}

	// Iterate over the results and store them in the slice
	// Next() returns false when there are no more results
	for result.Next() {
		// Store the current result in a map and remove the metadata
		currentValue := result.Record().Values()
		delete(currentValue, "_start")
		delete(currentValue, "_stop")
		delete(currentValue, "_time")
		metrics = append(metrics, currentValue)
	}

	// Check if there was an error during iteration
	if result.Err() != nil {
		log.Printf("Query error: %s\n", result.Err().Error())
		return system_metrics.SystemMetric{}, result.Err()
	}

	// Convert the slice to JSON so it can be unmarshalled into a struct
	parsed, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("Error when encoding data to json: %v\n", err)
		return system_metrics.SystemMetric{}, err
	}

	// Unmarshal the JSON into a struct
	var parsedMetrics []*system_metrics.Metric
	err = json.Unmarshal(parsed, &parsedMetrics)
	if err != nil {
		log.Printf("Error when decoding json to struct: %v\n", err)
		return system_metrics.SystemMetric{}, err
	}

	// Check if any metrics were found
	// If the database doesn't contain any metrics for the given duration, the query will return an empty result without throwing an error
	if len(parsedMetrics) == 0 {
		log.Printf("No metrics found for host '%s' in bucket '%s'\n", host, api.Bucket)
		return system_metrics.SystemMetric{}, fmt.Errorf("No metrics found for host '%s' in bucket '%s'", host, api.Bucket)
	}

	return system_metrics.SystemMetric{Id: host, Metrics: parsedMetrics}, nil
}

// DeleteBucket deletes the contents of the bucket within the given time d.
// The duration d specifies how far back in time from now the metrics should be deleted.
// Returns an error if any error occurs during the deletion process.
func (api InfluxDBApi) DeleteBucket(d time.Duration) error {
	// Get the organization
	org, err := api.OrganizationsAPI().FindOrganizationByName(context.Background(), api.Org)
	if err != nil {
		fmt.Printf("Error retrieving organization: %s\n", err)
		return err
	}

	// Get the bucket
	bucket, err := api.BucketsAPI().FindBucketByName(context.Background(), api.Bucket)
	if err != nil {
		fmt.Printf("Error retrieving bucket '%s': %s\n", api.Bucket, err)
		return err
	}

	// The predicate is used to filter the data to be deleted.
	// In this case we only want to delete the data with the given measurement.
	predicate := fmt.Sprintf(`_measurement="%s"`, api.Measurement)

	// Delete the data
	err = api.DeleteAPI().Delete(context.Background(), org, bucket, time.Now().Local().Add(-d), time.Now().Local(), predicate)
	if err != nil {
		fmt.Printf("Error deleting contents of bucket '%s': %s\n", api.Bucket, err)
		return err
	}

	fmt.Printf("Deleted contents of bucket '%s' with measurement '%s'\n", api.Bucket, api.Measurement)
	return nil
}

// DeleteHost deletes the metrics for the given host within the given time d.
// The duration d specifies how far back in time from now the metrics should be deleted.
// Returns an error if any error occurs during the deletion process.
func (api InfluxDBApi) DeleteHost(host string, d time.Duration) error {
	//TODO: allow org selection
	org, err := api.OrganizationsAPI().FindOrganizationByName(context.Background(), api.Org)
	if err != nil {
		fmt.Printf("Error retrieving organization: %s\n", err)
		return err
	}

	bucket, err := api.BucketsAPI().FindBucketByName(context.Background(), api.Bucket)
	if err != nil {
		fmt.Printf("Error retrieving bucket '%s': %s\n", api.Bucket, err)
		return err
	}

	// The predicate is used to filter the data to be deleted.
	// In this case we only want to delete the data with the given measurement and host.
	predicate := fmt.Sprintf(`host="%s" and _measurement="%s"`, host, api.Measurement)

	// Delete the data
	err = api.DeleteAPI().Delete(context.Background(), org, bucket, time.Now().Local().Add(-d), time.Now().Local(), predicate)
	if err != nil {
		fmt.Printf("Error deleting host '%s': %s\n", host, err)
		return err
	}

	fmt.Printf("Deleted host '%s' from bucket '%s' with measurement '%s'\n", host, api.Bucket, api.Measurement)
	return nil
}

// WriteMetric writes the given metric to InfluxDB synchronously.
// It takes the metric to be written m, the id of the host the metric belongs to, and the timestamp that should be used.
// Returns an error if any error occurs during the writing process.
func (api InfluxDBApi) WriteMetric(m system_metrics.Metric, id string, timestamp time.Time) error {
	// Create a blocking write client
	writeAPI := api.WriteAPIBlocking(api.Org, api.Bucket)

	// Convert the timestamp field to Unix time based on the given timestamp
	m.Timestamp = timestamp.Unix()

	// Create a new point and write it to InfluxDB
	p := influxdb2.NewPoint(api.Measurement, map[string]string{"host": id}, m.ToMap(), timestamp)
	if err := writeAPI.WritePoint(context.Background(), p); err != nil {
		return err
	}

	return nil
}

// WriteAnomalies writes the given anomalies to InfluxDB asynchronously.
// It takes the anomalies to be written, the host the anomalies belong to, and the algorithm used to detect the anomalies.
// Returns an error if any error occurs during the writing process.
func (api InfluxDBApi) WriteAnomalies(anomalies []system_metrics.AnomalyDetectionOutput, host string, algorithm string) error {
	// Create a non-blocking write client
	writeAPI := api.WriteAPI(api.Org, api.Bucket)

	// Iterate over the anomalies and write them to InfluxDB
	for _, a := range anomalies {
		// Create a new point and write it to InfluxDB
		p := influxdb2.NewPoint(api.Measurement, map[string]string{"host": host, "algorithm": algorithm}, a.ToMap(), time.Unix(a.Timestamp, 0))
		writeAPI.WritePoint(p)
	}
	// Write any remaining points
	writeAPI.Flush()
	// FIXME: Handle errors that might occur when writing asynchronously
	return nil
}
