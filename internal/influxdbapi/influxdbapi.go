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

// FIXME: Move to config file or something
const (
	ORG         = "pdc-mad"
	BUCKET      = "simba"
	MEASUREMENT = "test"
)

type InfluxDBApi struct {
	influxdb2.Client
}

func NewInfluxDBApi(token, host, port string) InfluxDBApi {
	return InfluxDBApi{
		influxdb2.NewClient("http://"+host+":"+port, token),
	}
}

func (api InfluxDBApi) WriteMetrics(m system_metrics.SystemMetric, gap time.Duration) error {
	writeAPI := api.WriteAPI(ORG, BUCKET)

	// Find the newest timestamp and go that many seconds back in time
	// FIXME: Maybe add time as parameter
	if time.Duration(time.Duration.Seconds(gap)) > time.Duration(m.Metrics[len(m.Metrics)-1].Timestamp) {
		log.Fatal("Gap exceeds length of the metric file")
	}
	now := time.Now().Local()
	end := now.Add(-gap)
	then := end.Add(time.Second * time.Duration(-m.Metrics[len(m.Metrics)-1].Timestamp))

	// Send all metrics to InfluxDB asynchronously
	for _, x := range m.Metrics {
		current := then.Add(time.Second * time.Duration(x.Timestamp))
		p := influxdb2.NewPoint("test", map[string]string{"host": m.Id}, x.ToMap(), current)
		writeAPI.WritePoint(p)
	}

	// Write any remaining points
	writeAPI.Flush()
	// FIXME: Handle errors
	return nil
}

// TODO duration need to be checked before use
func (api InfluxDBApi) GetMetrics(host, duration string) system_metrics.SystemMetric {
	queryAPI := api.QueryAPI(ORG)
	fmt.Println(host, duration)
	query := fmt.Sprintf(`from(bucket: "%v") |> range(start: -%v) |> filter(fn: (r) => r._measurement == "%v") |> filter(fn: (r) => r["host"] == "%v") |> pivot(rowKey: ["_time"],columnKey: ["_field"], valueColumn: "_value")`, BUCKET, duration, MEASUREMENT, host)
	result, err := queryAPI.Query(context.Background(), query)
	var metrics []map[string]interface{}
	if err == nil {
		// Use Next() to iterate over query result lines
		for result.Next() {
			currentValue := result.Record().Values()
			delete(currentValue, "_start")
			delete(currentValue, "_stop")
			delete(currentValue, "_time")
			metrics = append(metrics, currentValue)
		}

		if result.Err() != nil {
			fmt.Printf("Query error: %s\n", result.Err().Error())
		}
		parsed, err := json.Marshal(metrics)
		if err != nil {
			panic(err)
		}
		var parsedMetrics []*system_metrics.Metric
		json.Unmarshal(parsed, &parsedMetrics)
		return system_metrics.SystemMetric{Id: host, Metrics: parsedMetrics}
	} else {
		panic(err)
	}
	// Ensures background processes finishes
}

// Deletes all the metrics contained in the bucket in the time interval
// defined by the current time and the range specified by t
func (api InfluxDBApi) DeleteBucket(t time.Duration) error {
	//TODO: allow org selection
	org, err := api.OrganizationsAPI().FindOrganizationByName(context.Background(), ORG)
	if err != nil {
		fmt.Printf("Error retrieving organization: %s\n", err)
		return err
	}

	bucket, err := api.BucketsAPI().FindBucketByName(context.Background(), BUCKET)
	if err != nil {
		fmt.Printf("Error retrieving bucket '%s': %s\n", BUCKET, err)
		return err
	}

	err = api.DeleteAPI().Delete(context.Background(), org, bucket, time.Now().Local().Add(-t), time.Now().Local(), "")
	if err != nil {
		fmt.Printf("Error deleting contents of bucket '%s': %s\n", BUCKET, err)
		return err
	}

	fmt.Printf("Data from bucket '%s' deleted succesfully\n", BUCKET)

	return nil
}

// Deletes all the metrics from host/system h contained in the bucket in
// the time interval defined by the current time and the range specified by t
func (api InfluxDBApi) DeleteHost(h string, t time.Duration) error {
	//TODO: allow org selection
	org, err := api.OrganizationsAPI().FindOrganizationByName(context.Background(), ORG)
	if err != nil {
		fmt.Printf("Error retrieving organization: %s\n", err)
		return err
	}

	bucket, err := api.BucketsAPI().FindBucketByName(context.Background(), BUCKET)
	if err != nil {
		fmt.Printf("Error retrieving bucket '%s': %s\n", BUCKET, err)
		return err
	}

	predicate := fmt.Sprintf(`host="%s"`, h)

	err = api.DeleteAPI().Delete(context.Background(), org, bucket, time.Now().Local().Add(-t), time.Now().Local(), predicate)
	if err != nil {
		fmt.Printf("Error deleting host '%s': %s\n", h, err)
		return err
	}

	fmt.Printf("Data from host '%s' in bucket '%s' deleted succesfully\n", h, BUCKET)

	return nil
}
