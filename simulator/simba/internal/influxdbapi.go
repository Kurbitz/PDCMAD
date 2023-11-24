package simba

import (
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func WriteMetric(m SystemMetric) error {
	// FIXME: Hide API key, maybe .env file?
	client := influxdb2.NewClient("http://localhost:8086", "secret")
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
