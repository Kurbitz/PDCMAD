package simba

import (
	"log"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type InfluxDBApi struct {
	Token string
	Url   string
}

func (i InfluxDBApi) WriteMetrics(m SystemMetric, gap time.Duration) error {
	client := influxdb2.NewClient(i.Url, i.Token)

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
	client.Close()
	return nil
}
