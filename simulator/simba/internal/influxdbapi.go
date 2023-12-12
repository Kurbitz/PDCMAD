package simba

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

var org = "PDC-MAD"
var bucket = "metrics"
var measurement = "test"

type InfluxDBApi struct {
	influxdb2.Client
}

func NewInfluxDBApi(token, host, port string) InfluxDBApi {
	return InfluxDBApi{
		influxdb2.NewClient("http://"+host+":"+port, token),
	}
}
func (i InfluxDBApi) Close() {
	i.Close()
}

func (i InfluxDBApi) GetLastMetric(id string) (*Metric, error) {
	q := i.QueryAPI(org)
	query := fmt.Sprintf("from(bucket:\"%v\") |> range(start: -30d) |> filter(fn: (r) => r._measurement == \"%v\") |> filter(fn: (r) => r.host == \"%v\")|> last()", bucket, measurement, id)
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

	fmt.Println(string(j))

	metric := Metric{}
	if err := json.Unmarshal(j, &metric); err != nil {
		return nil, err
	}

	return &metric, nil
}

func (i InfluxDBApi) WriteMetrics(m SystemMetric, gap time.Duration) error {

	writeAPI := i.WriteAPI(org, bucket)

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
		p := influxdb2.NewPoint(measurement, map[string]string{"host": m.Id}, x.ToMap(), current)
		writeAPI.WritePoint(p)
	}

	// Write any remaining points
	writeAPI.Flush()
	// FIXME: Handle errors
	return nil
}
