package simba

import (
	"context"
	"fmt"
	"log"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type InfluxDBApi struct {
	Token string
	Url   string
}

func NewInfluxDBApi(token, host, port string) InfluxDBApi {
	return InfluxDBApi{
		Token: token,
		Url:   "http://" + host + ":" + port,
	}
}

func (i InfluxDBApi) WriteMetrics(m SystemMetric, gap time.Duration) error {
	client := influxdb2.NewClient(i.Url, i.Token)
	defer client.Close()
	writeAPI := client.WriteAPI("test", "metrics")

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

// Deletes all the metrics contained in bucket b in the time interval
// defined by the current time and the range specified by t
func (i InfluxDBApi) DeleteBucket(b string, t time.Duration) error {
	var err error = nil

	client := influxdb2.NewClient(i.Url, i.Token)
	defer client.Close()

	//TODO: allow org selection
	org, err := client.OrganizationsAPI().FindOrganizationByName(context.Background(), "test")
	if err != nil {
		fmt.Printf("Error retrieving organization: %s\n", err)
		return err
	}

	bucket, err := client.BucketsAPI().FindBucketByName(context.Background(), b)
	if err != nil {
		fmt.Printf("Error retrieving bucket '%s': %s\n", b, err)
		return err
	}

	err = client.DeleteAPI().Delete(context.Background(), org, bucket, time.Now().Local().Add(-t), time.Now().Local(), "")
	if err != nil {
		fmt.Printf("Error deleting contents of bucket '%s': %s\n", b, err)
		return err
	}

	fmt.Printf("Data from bucket '%s' deleted succesfully\n", b)

	return nil
}

// Deletes all the metrics from host/system h contained in bucket b in
// the time interval defined by the current time and the range specified by t
func (i InfluxDBApi) DeleteHost(b string, h string, t time.Duration) error {
	var err error = nil

	client := influxdb2.NewClient(i.Url, i.Token)
	defer client.Close()

	//TODO: allow org selection
	org, err := client.OrganizationsAPI().FindOrganizationByName(context.Background(), "test")
	if err != nil {
		fmt.Printf("Error retrieving organization: %s\n", err)
		return err
	}

	bucket, err := client.BucketsAPI().FindBucketByName(context.Background(), b)
	if err != nil {
		fmt.Printf("Error retrieving bucket '%s': %s\n", b, err)
		return err
	}

	predicate := fmt.Sprintf(`host="%s"`, h)

	err = client.DeleteAPI().Delete(context.Background(), org, bucket, time.Now().Local().Add(-t), time.Now().Local(), predicate)
	if err != nil {
		fmt.Printf("Error deleting host '%s': %s\n", h, err)
		return err
	}

	fmt.Printf("Data from host '%s' in bucket '%s' deleted succesfully\n", h, b)

	return nil
}
