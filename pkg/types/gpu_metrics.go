// Package types defines data structures for GPU metrics collection.
package types

import "time"

// GPUMetrics represents current metrics for a single GPU.
type GPUMetrics struct {
	Hostname          string    `json:"hostname"`
	GPUID             int       `json:"gpu_id"`
	Timestamp         time.Time `json:"timestamp"`
	GPUName           string    `json:"gpu_name"`
	Temperature       float64   `json:"temperature"`
	FreeMemory        uint64    `json:"free_memory"`
	UsedMemory        uint64    `json:"used_memory"`
	TotalMemory       uint64    `json:"total_memory"`
	GPUUtilization    float64   `json:"gpu_utilization"`
	MemoryUtilization float64   `json:"memory_utilization"`
}

// GPUProcess represents information about a process running on GPU.
type GPUProcess struct {
	Hostname      string    `json:"hostname"`
	GPUID         int       `json:"gpu_id"`
	Timestamp     time.Time `json:"timestamp"`
	User          string    `json:"user"`
	PID           int       `json:"pid"`
	ProcessName   string    `json:"process_name"`
	UsedGPUMemory uint64    `json:"used_gpu_memory"` // MiB
	UsedCPU       float64   `json:"used_cpu"`
	UsedMemory    float64   `json:"used_memory"`
	Command       string    `json:"command"`
}

// CollectorConfig represents GPU metrics collection configuration.
type CollectorConfig struct {
	Timeout          time.Duration `json:"timeout"`
	NvidiaSmiPath    string        `json:"nvidia_smi_path"`
	HostnameOverride string        `json:"hostname_override"`
}
