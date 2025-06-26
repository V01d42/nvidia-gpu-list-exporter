// Package metrics provides Prometheus metrics collection and management for GPU monitoring.
package metrics

import (
	"fmt"
	"strings"

	"github.com/nvidia-gpu-list-exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics represents a collection of Prometheus metrics for GPU monitoring.
type Metrics struct {
	gpuTemperature    *prometheus.GaugeVec
	gpuMemoryFree     *prometheus.GaugeVec
	gpuMemoryUsed     *prometheus.GaugeVec
	gpuMemoryTotal    *prometheus.GaugeVec
	gpuUtilization    *prometheus.GaugeVec
	memoryUtilization *prometheus.GaugeVec
	gpuProcessMemory  *prometheus.GaugeVec
	gpuProcessCount   *prometheus.GaugeVec
	systemImageInfo   *prometheus.GaugeVec
}

// New creates a new Prometheus metrics collection.
func New() *Metrics {
	commonLabels := []string{"hostname", "gpu_id", "gpu_uuid", "gpu_name"}

	return &Metrics{
		gpuTemperature: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nvidia_gpu_temperature_celsius",
				Help: "GPU temperature in Celsius (DCGM_FI_DEV_GPU_TEMP)",
			},
			commonLabels,
		),

		gpuMemoryFree: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nvidia_gpu_memory_free_bytes",
				Help: "GPU free memory in bytes (DCGM_FI_DEV_FB_FREE)",
			},
			commonLabels,
		),

		gpuMemoryUsed: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nvidia_gpu_memory_used_bytes",
				Help: "GPU used memory in bytes (DCGM_FI_DEV_FB_USED)",
			},
			commonLabels,
		),

		gpuMemoryTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nvidia_gpu_memory_total_bytes",
				Help: "GPU total memory in bytes",
			},
			commonLabels,
		),

		gpuUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nvidia_gpu_utilization_percent",
				Help: "GPU utilization percentage (DCGM_FI_DEV_GPU_UTIL)",
			},
			commonLabels,
		),

		memoryUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nvidia_gpu_memory_utilization_percent",
				Help: "GPU memory utilization percentage (DCGM_FI_DEV_MEM_COPY_UTIL)",
			},
			commonLabels,
		),

		gpuProcessMemory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nvidia_gpu_process_memory_bytes",
				Help: "GPU process memory usage in bytes",
			},
			append(commonLabels, "pid", "user", "command"),
		),

		gpuProcessCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nvidia_gpu_process_count",
				Help: "Number of GPU processes",
			},
			commonLabels,
		),

		systemImageInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nvidia_system_image_info",
				Help: "System image information (Image Info) - value is always 1",
			},
			[]string{"hostname", "boot_image_version", "os_version", "kernel_version"},
		),
	}
}

// Register registers all metrics with the Prometheus registry.
func (m *Metrics) Register(registry prometheus.Registerer) error {
	collectors := []prometheus.Collector{
		m.gpuTemperature,
		m.gpuMemoryFree,
		m.gpuMemoryUsed,
		m.gpuMemoryTotal,
		m.gpuUtilization,
		m.memoryUtilization,
		m.gpuProcessMemory,
		m.gpuProcessCount,
		m.systemImageInfo,
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
		gpuName := metric.GPUName
		if gpuName == "" {
			gpuName = "unknown"
		}

		gpuUUID := fmt.Sprintf("GPU-%s", metric.GPUID)

		labels := prometheus.Labels{
			"hostname": metric.Hostname,
			"gpu_id":   metric.GPUID,
			"gpu_uuid": gpuUUID,
			"gpu_name": gpuName,
		}

		m.gpuTemperature.With(labels).Set(metric.Temperature)
		m.gpuMemoryFree.With(labels).Set(float64(metric.MemoryFree * 1024 * 1024))
		m.gpuMemoryUsed.With(labels).Set(float64(metric.MemoryUsed * 1024 * 1024))
		m.gpuMemoryTotal.With(labels).Set(float64(metric.MemoryTotal * 1024 * 1024))
		m.gpuUtilization.With(labels).Set(metric.GPUUtilization)
		m.memoryUtilization.With(labels).Set(metric.MemoryUtilization)
	}
}

// UpdateProcesses updates GPU process metrics.
func (m *Metrics) UpdateProcesses(processes []types.GPUProcess) {
	processCountByGPU := make(map[string]int)

	for _, process := range processes {
		gpuUUID := fmt.Sprintf("GPU-%s", process.GPUID)
		gpuName := "unknown"

		processCountByGPU[process.GPUID]++
		processMemoryLabels := prometheus.Labels{
			"hostname": process.Hostname,
			"gpu_id":   process.GPUID,
			"gpu_uuid": gpuUUID,
			"gpu_name": gpuName,
			"pid":      fmt.Sprintf("%d", process.PID),
			"user":     process.User,
			"command":  process.ProcessName,
		}
		m.gpuProcessMemory.With(processMemoryLabels).Set(float64(process.UsedGPUMemory * 1024 * 1024))
	}

	for gpuID, count := range processCountByGPU {
		hostname := "unknown"
		if parts := strings.Split(gpuID, "-"); len(parts) >= 2 {
			hostname = strings.Join(parts[:len(parts)-1], "-")
		}

		processCountLabels := prometheus.Labels{
			"hostname": hostname,
			"gpu_id":   gpuID,
			"gpu_uuid": fmt.Sprintf("GPU-%s", gpuID),
			"gpu_name": "unknown",
		}
		m.gpuProcessCount.With(processCountLabels).Set(float64(count))
	}
}

// UpdateSystemInfo updates system information metrics.
func (m *Metrics) UpdateSystemInfo(systemInfo types.SystemImageInfo) {
	labels := prometheus.Labels{
		"hostname":           systemInfo.Hostname,
		"boot_image_version": systemInfo.BootImageVersion,
		"os_version":         systemInfo.OSVersion,
		"kernel_version":     systemInfo.KernelVersion,
	}

	m.systemImageInfo.With(labels).Set(1)
}
