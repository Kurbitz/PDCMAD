package simba

import (
	"log"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

// FIXME: Move to config file or something
const (
	org         = "pdc-mad"
	bucket      = "simba"
	measurement = "test"
)

type InfluxDBApi struct {
	influxdb2.Client
}

func NewInfluxDBApi(token, host, port string) InfluxDBApi {
	return InfluxDBApi{
		influxdb2.NewClient("http://"+host+":"+port, token),
	}
}

func (api InfluxDBApi) WriteMetrics(m SystemMetric, gap time.Duration) error {
	writeAPI := api.WriteAPI(org, bucket)

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
		// Set the timestamp to the current Unix timestamp
		x.Timestamp = current.Unix()
		p := influxdb2.NewPoint(measurement, map[string]string{"host": m.Id}, x.ToMap(), current)
		writeAPI.WritePoint(p)
	}

	// Write any remaining points
	writeAPI.Flush()
	// FIXME: Handle errors
	return nil
}

func (api InfluxDBApi) WriteMetric(m Metric, id string, timeStamp time.Time) error {
	writeAPI := api.WriteAPI(org, bucket)

	m.Timestamp = timeStamp.Unix()
	p := influxdb2.NewPoint(measurement, map[string]string{"host": id}, m.ToMap(), timeStamp)

	writeAPI.WritePoint(p)

	return nil
}
