package main

import (
	"fmt"
	system_metrics "internal/system_metrics"
	"math"
)

// AnomalyMap is a map that maps anomaly names to transformation functions that will be applied to the metrics.
// The anomaly names are the same as the anomaly flags that can be passed to the fill and stream commands.
// To add a new anomaly, add a new entry to this map with the anomaly name as the key and the transformation function as the value.
var AnomalyMap = map[string]func(m *system_metrics.SystemMetric) error{
	"cpu-user-high": cpuUserHigh,
	"cpu-user-sin":  cpuUserSin,
}

// InjectAnomaly injects an anomaly into the metrics based on the anomalyFlag.
// If the anomalyFlag is empty, no anomaly will be injected.
// If the anomalyFlag is not empty, but does not exist in the AnomalyMap, an error will be returned.
// If the anomalyFlag exists in the AnomalyMap, the transformation function will be called with the metrics as the argument.
// Any errors that the transformation function returns will be returned.
func InjectAnomaly(metrics *system_metrics.SystemMetric, anomalyFlag string) error {
	if anomalyFlag == "" {
		return nil
	}

	// Check if the anomalyFlag string exists in the AnomalyMap
	if _, ok := AnomalyMap[anomalyFlag]; !ok {
		return fmt.Errorf("anomaly flag %s does not exist", anomalyFlag)
	}

	// Call the transformation function found in the AnomalyMap
	if err := AnomalyMap[anomalyFlag](metrics); err != nil {
		return err
	}

	return nil
}

// Basic example anomaly. Sets Cpu_User to 1 for all metrics
func cpuUserHigh(metrics *system_metrics.SystemMetric) error {
	for _, m := range metrics.Metrics {
		m.Cpu_User = 1
	}

	return nil
}

// Changes Cpu_User to a timestamp based sin function (absolut value of sin)
func cpuUserSin(metrics *system_metrics.SystemMetric) error {
	for _, m := range metrics.Metrics {
		m.Cpu_User = math.Abs(math.Sin(float64(m.Timestamp / 10)))
	}

	return nil
}
