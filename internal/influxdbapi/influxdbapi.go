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
type InfluxDBApi struct {
	influxdb2.Client
	Org         string
	Bucket      string
	Measurement string
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

// Returns the last metric from the host h
func (api InfluxDBApi) GetLastMetric(host string) (*system_metrics.Metric, error) {
	q := api.QueryAPI(api.Org)
	query := fmt.Sprintf("from(bucket:\"%v\") |> range(start: -30d) |> filter(fn: (r) => r._measurement == \"%v\") |> filter(fn: (r) => r.host == \"%v\")|> last()", api.Bucket, api.Measurement, host)
	result, err := q.Query(context.Background(), query)

	results := make(map[string]interface{}, 0)

	if err != nil {
		return nil, err
	}
	for result.Next() {
		results[result.Record().Field()] = result.Record().Value()
	}

	j, err := json.Marshal(results)
	if err != nil {
		return nil, err
	}

	metric := system_metrics.Metric{}
	if err := json.Unmarshal(j, &metric); err != nil {
		return nil, err
	}

	return &metric, nil
}

// Writes all the metrics contained in the SystemMetric m
func (api InfluxDBApi) WriteMetrics(m system_metrics.SystemMetric, gap time.Duration, onWrite func()) error {
	writeAPI := api.WriteAPI(api.Org, api.Bucket)

	// Find the newest timestamp and go that many seconds back in time
	now := time.Now().Local()
	end := now.Add(-gap)
	then := end.Add(time.Second * time.Duration(-m.Metrics[len(m.Metrics)-1].Timestamp))

	// Send all metrics to InfluxDB asynchronously
	for _, x := range m.Metrics {
		current := then.Add(time.Second * time.Duration(x.Timestamp))
		// Set the timestamp to the current Unix timestamp
		x.Timestamp = current.Unix()
		p := influxdb2.NewPoint(api.Measurement, map[string]string{"host": m.Id}, x.ToMap(), current)
		writeAPI.WritePoint(p)
		onWrite()
	}

	// Write any remaining points
	writeAPI.Flush()
	// FIXME: Handle errors
	return nil
}

// TODO duration need to be checked before use
func (api InfluxDBApi) GetMetrics(host, duration string) (system_metrics.SystemMetric, error) {
	queryAPI := api.QueryAPI(api.Org)
	query := fmt.Sprintf(`from(bucket: "%v") |> range(start: -%v) |> filter(fn: (r) => r._measurement == "%v") |> filter(fn: (r) => r["host"] == "%v") |> pivot(rowKey: ["_time"],columnKey: ["_field"], valueColumn: "_value")`, api.Bucket, duration, api.Measurement, host)
	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		log.Printf("Error performing query on bucket '%s': %s\n", api.Bucket, err)
		return system_metrics.SystemMetric{}, err
	}
	if result == nil {
		log.Printf("Duration '%v' on bucket '%v' is empty", duration, api.Bucket)
		return system_metrics.SystemMetric{}, fmt.Errorf("Empty query result")
	}
	var metrics []map[string]interface{}
	// Use Next() to iterate over query result lines
	for result.Next() {
		currentValue := result.Record().Values()
		delete(currentValue, "_start")
		delete(currentValue, "_stop")
		delete(currentValue, "_time")
		metrics = append(metrics, currentValue)
	}

	if result.Err() != nil {
		log.Printf("Query error: %s\n", result.Err().Error())
		return system_metrics.SystemMetric{}, result.Err()
	}

	parsed, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("Error when encoding data to json: %v\n", err)
		return system_metrics.SystemMetric{}, err
	}
	var parsedMetrics []*system_metrics.Metric
	err = json.Unmarshal(parsed, &parsedMetrics)
	if err != nil {
		log.Printf("Error when decoding json to struct: %v\n", err)
		return system_metrics.SystemMetric{}, err
	}
	return system_metrics.SystemMetric{Id: host, Metrics: parsedMetrics}, nil
}

// Deletes all the metrics contained in the bucket in the time interval
// defined by the current time and the range specified by t
func (api InfluxDBApi) DeleteBucket(t time.Duration) error {
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
	predicate := fmt.Sprintf(`_measurement="%s"`, api.Measurement)

	err = api.DeleteAPI().Delete(context.Background(), org, bucket, time.Now().Local().Add(-t), time.Now().Local(), predicate)
	if err != nil {
		fmt.Printf("Error deleting contents of bucket '%s': %s\n", api.Bucket, err)
		return err
	}

	fmt.Printf("Data from bucket '%s' deleted succesfully\n", api.Bucket)

	return nil
}

// Deletes all the metrics from host/system h contained in the bucket in
// the time interval defined by the current time and the range specified by t
func (api InfluxDBApi) DeleteHost(h string, t time.Duration) error {
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

	predicate := fmt.Sprintf(`host="%s" and _measurement="%s"`, h, api.Measurement)

	err = api.DeleteAPI().Delete(context.Background(), org, bucket, time.Now().Local().Add(-t), time.Now().Local(), predicate)
	if err != nil {
		fmt.Printf("Error deleting host '%s': %s\n", h, err)
		return err
	}

	fmt.Printf("Data from host '%s' in bucket '%s' deleted succesfully\n", h, api.Bucket)

	return nil
}

func (api InfluxDBApi) WriteMetric(m system_metrics.Metric, id string, timeStamp time.Time) error {
	writeAPI := api.WriteAPIBlocking(api.Org, api.Bucket)
	m.Timestamp = timeStamp.Unix()
	p := influxdb2.NewPoint(api.Measurement, map[string]string{"host": id}, m.ToMap(), timeStamp)
	if err := writeAPI.WritePoint(context.Background(), p); err != nil {
		return err
	}

	return nil
}

func (api InfluxDBApi) WriteAnomalies(anomalies []system_metrics.Anomaly) error {
	writeAPI := api.WriteAPI(api.Org, api.Bucket)
	for _, a := range anomalies {
		p := influxdb2.NewPoint("anomalies", map[string]string{"host": a.Host}, a.ToMap(), time.Unix(a.Timestamp, 0))
		writeAPI.WritePoint(p)
	}
	writeAPI.Flush()
	return nil
}
