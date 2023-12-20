package main

import (
	system_metrics "internal/system_metrics"
	"math"
)

var AnomalyMap = map[string]func(m *system_metrics.SystemMetric) error{
	"a1": anomaly1,
	"a2": anomaly2,
}

func injectAnomaly(metrics *system_metrics.SystemMetric, anomalyFlag string) error {
	if anomalyFlag == "" {
		return nil
	}
	// Find the first metric that is after the startAt time
	if err := AnomalyMap[anomalyFlag](metrics); err != nil {
		return err
	}

	return nil
}

// Basic example anomaly. Sets Cpu_User to 1
func anomaly1(metrics *system_metrics.SystemMetric) error {
	for _, m := range metrics.Metrics {
		m.Cpu_User = 1
	}

	return nil
}

// Changes Cpu_User to a timestamp based sine
func anomaly2(metrics *system_metrics.SystemMetric) error {
	for _, m := range metrics.Metrics {
		m.Cpu_User = math.Abs(math.Sin(float64(m.Timestamp / 10)))
	}

	return nil
}
