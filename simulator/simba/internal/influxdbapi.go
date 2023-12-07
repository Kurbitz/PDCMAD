package simba

import (
	"fmt"
	"os"

	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func WriteAllMetrics(m SystemMetric) error {
	// FIXME: Hide API key, maybe .env file?
	token := os.Getenv("INFLUXDB_TOKEN")
	serverUrl := fmt.Sprintf("http://%v:%v", os.Getenv("INFLUXDB_IP"), os.Getenv("INFLUXDB_PORT"))
	client := influxdb2.NewClient(serverUrl, token)
	writeAPI := client.WriteAPI("test", "metrics")

	// Find the newest timestamp and go that many seconds back in time
	// FIXME: Maybe add time as parameter
	now := time.Now().Local()
	then := now.Add(time.Second * time.Duration(-m.Metrics[len(m.Metrics)-1].Timestamp))

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
func WriteMetrics(m SystemMetric, gap time.Duration) error {
	token := os.Getenv("INFLUXDB_TOKEN")
	serverUrl := fmt.Sprintf("http://%v:%v", os.Getenv("INFLUXDB_IP"), os.Getenv("INFLUXDB_PORT"))
	client := influxdb2.NewClient(serverUrl, token)
	writeAPI := client.WriteAPI("test", "metrics")

	// Find the newest timestamp and go that many seconds back in time
	// FIXME: Maybe add time as parameter
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
