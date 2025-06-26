// Package types defines data structures for GPU metrics collection.
package types

import "time"

// GPUMetrics represents current metrics for a single GPU.
type GPUMetrics struct {
	Hostname          string    `json:"hostname"`
	GPUID             string    `json:"gpu_id"`
	Timestamp         time.Time `json:"timestamp"`
	GPUName           string    `json:"gpu_name"`
	Temperature       float64   `json:"temperature"`
	MemoryFree        uint64    `json:"memory_free"`
	MemoryUsed        uint64    `json:"memory_used"`
	MemoryTotal       uint64    `json:"memory_total"`
	GPUUtilization    float64   `json:"gpu_utilization"`
	MemoryUtilization float64   `json:"memory_utilization"`
}

// GPUProcess represents information about a process running on GPU.
type GPUProcess struct {
	Hostname      string    `json:"hostname"`
	Timestamp     time.Time `json:"timestamp"`
	GPUID         string    `json:"gpu_id"`
	PID           uint32    `json:"pid"`
	ProcessName   string    `json:"process_name"`
	UsedGPUMemory uint64    `json:"used_gpu_memory"` // MiB
	User          string    `json:"user"`
	Command       string    `json:"command"`
}

// SystemImageInfo represents system image information.
type SystemImageInfo struct {
	Hostname         string    `json:"hostname"`
	Timestamp        time.Time `json:"timestamp"`
	BootImageVersion string    `json:"boot_image_version"`
	OSVersion        string    `json:"os_version"`
	KernelVersion    string    `json:"kernel_version"`
}

// CollectorConfig represents GPU metrics collection configuration.
type CollectorConfig struct {
	Timeout          time.Duration `json:"timeout"`
	NvidiaSmiPath    string        `json:"nvidia_smi_path"`
	HostnameOverride string        `json:"hostname_override"`
}
