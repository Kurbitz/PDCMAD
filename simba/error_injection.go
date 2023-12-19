package main

import (
	"fmt"
	system_metrics "internal/system_metrics"
	"math"
	"time"
)

var anomalyMap = map[string]func(m *system_metrics.SystemMetric) *system_metrics.SystemMetric{
	"a1": anomaly1,
	"a2": anomaly2,
}

func Pipeline(metrics *system_metrics.SystemMetric, anomalyFlag string, aStart time.Duration, aEnd time.Duration) {
	var startIndex int
	var metricsAux system_metrics.SystemMetric

	if anomalyFlag == "" {
		return
	}

	// Find the first metric that is after the startAt time
	for i, m := range metrics.Metrics {
		if time.Second*time.Duration(m.Timestamp) >= aStart {
			startIndex = i
			break
		}
	}

	fmt.Printf("StartIndex = %d\n", startIndex)

	metricsAux.Metrics = make([]*system_metrics.Metric, len(metrics.Metrics))
	copy(metricsAux.Metrics, metrics.Metrics)
	metricsAux.SliceBetween(aStart, aEnd)

	fmt.Printf("MetricsAux len = %d\n", len(metricsAux.Metrics))

	metrics.Swap(anomalyMap[anomalyFlag](&metricsAux), startIndex)
}

// Basic example anomaly. Sets Cpu_User to 1
func anomaly1(metrics *system_metrics.SystemMetric) *system_metrics.SystemMetric {
	for _, m := range metrics.Metrics {
		m.Cpu_User = 1
	}

	return metrics
}

// Changes Cpu_User to a timestamp based sine
func anomaly2(metrics *system_metrics.SystemMetric) *system_metrics.SystemMetric {
	for _, m := range metrics.Metrics {
		m.Cpu_User = math.Abs(math.Sin(float64(m.Timestamp / 10)))
	}

	return metrics
}
