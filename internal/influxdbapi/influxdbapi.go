package influxdbapi

import (
	"context"
	"fmt"
	"internal/system_metrics"
	"log"
	"math"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

// FIXME: Move to config file or something
const (
	ORG         = "pdc-mad"
	BUCKET      = "simba"
	MEASUREMENT = "test"
)

// TODO: move it to a better place
var anomalyMap = map[string]func(m *system_metrics.Metric) *system_metrics.Metric{
	"a0": anomaly0,
	"a1": anomaly1,
	"a2": anomaly2,
}

type InfluxDBApi struct {
	influxdb2.Client
}

func NewInfluxDBApi(token, host, port string) InfluxDBApi {
	return InfluxDBApi{
		influxdb2.NewClient("http://"+host+":"+port, token),
	}
}

func (api InfluxDBApi) WriteMetrics(m system_metrics.SystemMetric, gap time.Duration, anomaly string) error {
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
		transformed := anomalyMap[anomaly](x) //Technically the pipeline. If no transformation is to be applied the map should return the same metric
		p := influxdb2.NewPoint("test", map[string]string{"host": m.Id}, transformed.ToMap(), current)
		writeAPI.WritePoint(p)
	}

	// Write any remaining points
	writeAPI.Flush()
	// FIXME: Handle errors
	return nil
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

// TODO: should move to own file?
// Doesn't change
func anomaly0(m *system_metrics.Metric) *system_metrics.Metric {
	return m
}

// Basic example anomaly. Sets Cpu_User to 1
func anomaly1(m *system_metrics.Metric) *system_metrics.Metric {
	m.Cpu_User = 1

	return m
}

// Changes Cpu_User to a timestamp based sine
func anomaly2(m *system_metrics.Metric) *system_metrics.Metric {
	m.Cpu_User = math.Abs(math.Sin(float64(m.Timestamp)))

	return m
}
