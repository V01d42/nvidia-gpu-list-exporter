// Package metrics provides Prometheus metrics collection and management for GPU monitoring.
package metrics

import (
	"strconv"

	"github.com/nvidia-gpu-list-exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics represents a collection of Prometheus metrics for GPU monitoring.
type Metrics struct {
	gpuTemperature    *prometheus.GaugeVec
	gpuFreeMemory     *prometheus.GaugeVec
	gpuUsedMemory     *prometheus.GaugeVec
	gpuTotalMemory    *prometheus.GaugeVec
	gpuUtilization    *prometheus.GaugeVec
	memoryUtilization *prometheus.GaugeVec
	processGPUMemory  *prometheus.GaugeVec
	processCPU        *prometheus.GaugeVec
	processMemory     *prometheus.GaugeVec
}

// New creates a new Prometheus metrics collection.
func New() *Metrics {
	gpuLabels := []string{"hostname", "gpu_id", "gpu_name"}
	processLabels := []string{"hostname", "gpu_id", "pid", "process_name", "user", "command"}

	return &Metrics{
		gpuTemperature: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nvidia_gpu_temperature_celsius",
				Help: "GPU temperature in Celsius",
			},
			gpuLabels,
		),

		gpuFreeMemory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nvidia_gpu_free_memory_bytes",
				Help: "GPU free memory in bytes",
			},
			gpuLabels,
		),

		gpuUsedMemory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nvidia_gpu_used_memory_bytes",
				Help: "GPU used memory in bytes",
			},
			gpuLabels,
		),

		gpuTotalMemory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nvidia_gpu_total_memory_bytes",
				Help: "GPU total memory in bytes",
			},
			gpuLabels,
		),

		gpuUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nvidia_gpu_utilization_percent",
				Help: "GPU utilization percentage",
			},
			gpuLabels,
		),

		memoryUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nvidia_gpu_memory_utilization_percent",
				Help: "GPU memory utilization percentage",
			},
			gpuLabels,
		),

		processGPUMemory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nvidia_gpu_process_gpu_memory_bytes",
				Help: "GPU process memory usage in bytes",
			},
			processLabels,
		),

		processCPU: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nvidia_gpu_process_cpu_percent",
				Help: "GPU process CPU usage percentage",
			},
			processLabels,
		),

		processMemory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nvidia_gpu_process_memory_percent",
				Help: "GPU process memory usage percentage",
			},
			processLabels,
		),
	}
}

// Register registers all metrics with the Prometheus registry.
func (m *Metrics) Register(registry prometheus.Registerer) error {
	collectors := []prometheus.Collector{
		m.gpuTemperature,
		m.gpuFreeMemory,
		m.gpuUsedMemory,
		m.gpuTotalMemory,
		m.gpuUtilization,
		m.memoryUtilization,
		m.processGPUMemory,
		m.processCPU,
		m.processMemory,
	}

	for _, collector := range collectors {
		if err := registry.Register(collector); err != nil {
			return err
		}
	}

	return nil
}

// UpdateGPU updates GPU metrics with the provided data.
func (m *Metrics) UpdateGPU(gpuMetrics []types.GPUMetrics) {
	for _, metric := range gpuMetrics {
		labels := prometheus.Labels{
			"hostname": metric.Hostname,
			"gpu_id":   strconv.Itoa(metric.GPUID),
			"gpu_name": metric.GPUName,
		}

		m.gpuTemperature.With(labels).Set(metric.Temperature)
		m.gpuFreeMemory.With(labels).Set(float64(metric.FreeMemory * 1024 * 1024))
		m.gpuUsedMemory.With(labels).Set(float64(metric.UsedMemory * 1024 * 1024))
		m.gpuTotalMemory.With(labels).Set(float64(metric.TotalMemory * 1024 * 1024))
		m.gpuUtilization.With(labels).Set(metric.GPUUtilization)
		m.memoryUtilization.With(labels).Set(metric.MemoryUtilization)
	}
}

// UpdateProcesses updates GPU process metrics.
func (m *Metrics) UpdateProcesses(processes []types.GPUProcess) {
	for _, process := range processes {
		labels := prometheus.Labels{
			"hostname":     process.Hostname,
			"gpu_id":       strconv.Itoa(process.GPUID),
			"pid":          strconv.Itoa(process.PID),
			"process_name": process.ProcessName,
			"user":         process.User,
			"command":      process.Command,
		}

		m.processGPUMemory.With(labels).Set(float64(process.UsedGPUMemory * 1024 * 1024))
		m.processCPU.With(labels).Set(process.UsedCPU)
		m.processMemory.With(labels).Set(process.UsedMemory)
	}
}
