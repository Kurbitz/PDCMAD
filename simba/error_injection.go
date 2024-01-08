package main

import (
	system_metrics "internal/system_metrics"
	"math"
)

var AnomalyMap = map[string]func(m *system_metrics.SystemMetric) error{
	"cpu-user-high": cpuUserHigh,
	"cpu-user-sin":  cpuUserSin,
}

func InjectAnomaly(metrics *system_metrics.SystemMetric, anomalyFlag string) error {
	if anomalyFlag == "" {
		return nil
	}
	//Call the error injection function that is related to the anomalyFlag in the AnomalyMap
	if err := AnomalyMap[anomalyFlag](metrics); err != nil {
		return err
	}

	return nil
}

// Basic example anomaly. Sets Cpu_User to 1
func cpuUserHigh(metrics *system_metrics.SystemMetric) error {
	for _, m := range metrics.Metrics {
		m.Cpu_User = 1
	}

	return nil
}

// Changes Cpu_User to a timestamp based sine
func cpuUserSin(metrics *system_metrics.SystemMetric) error {
	for _, m := range metrics.Metrics {
		m.Cpu_User = math.Abs(math.Sin(float64(m.Timestamp / 10)))
	}

	return nil
}
