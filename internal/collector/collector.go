// Package collector provides GPU metrics collection using nvidia-smi.
package collector

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/nvidia-gpu-list-exporter/pkg/types"
)

// Collector collects GPU metrics using nvidia-smi.
type Collector struct {
	config   types.CollectorConfig
	hostname string
}

// New creates a new Collector instance.
func New(config types.CollectorConfig) (*Collector, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}

	// Check for NODE_NAME environment variable (Kubernetes)
	if nodeName := os.Getenv("NODE_NAME"); nodeName != "" {
		hostname = nodeName
	}

	// Check for explicit hostname override
	if config.HostnameOverride != "" {
		hostname = config.HostnameOverride
	}

	if config.NvidiaSmiPath == "" {
		config.NvidiaSmiPath = "nvidia-smi"
	}

	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}

	c := &Collector{
		config:   config,
		hostname: hostname,
	}

	if err := c.checkAvailability(); err != nil {
		return nil, fmt.Errorf("nvidia-smi availability check failed: %w", err)
	}

	return c, nil
}

func (c *Collector) checkAvailability() error {
	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.config.NvidiaSmiPath, "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("nvidia-smi not found or cannot be executed: %w", err)
	}
	return nil
}

// CollectGPUMetrics collects current GPU metrics.
func (c *Collector) CollectGPUMetrics() ([]types.GPUMetrics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	query := "timestamp,index,gpu_name,memory.free,memory.used,memory.total,utilization.gpu,utilization.memory,temperature.gpu"
	cmd := exec.CommandContext(ctx, c.config.NvidiaSmiPath,
		"--query-gpu="+query,
		"--format=csv,noheader")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get GPU metrics: %w", err)
	}

	return c.parseGPUMetrics(string(output))
}

func (c *Collector) parseGPUMetrics(output string) ([]types.GPUMetrics, error) {
	reader := csv.NewReader(strings.NewReader(output))
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	metrics := make([]types.GPUMetrics, 0, len(records))

	for _, record := range records {
		if len(record) != 9 {
			continue
		}

		timestampStr := strings.TrimSpace(record[0])
		timestamp, err := time.Parse("2006/01/02 15:04:05.000", timestampStr)
		if err != nil {
			timestamp = time.Now()
		}

		gpuIndexStr := strings.TrimSpace(record[1])
		gpuIndex, err := strconv.Atoi(gpuIndexStr)
		if err != nil {
			continue
		}

		gpuName := strings.TrimSpace(record[2])
		if gpuName == "" {
			gpuName = "unknown"
		}

		freeMemory, err := c.parseUint64(record[3])
		if err != nil {
			continue
		}

		usedMemory, err := c.parseUint64(record[4])
		if err != nil {
			continue
		}

		totalMemory, err := c.parseUint64(record[5])
		if err != nil {
			continue
		}

		gpuUtil, err := c.parseFloat(record[6])
		if err != nil {
			continue
		}

		memUtil, err := c.parseFloat(record[7])
		if err != nil {
			continue
		}

		temperature, err := c.parseFloat(record[8])
		if err != nil {
			continue
		}

		metric := types.GPUMetrics{
			Hostname:          c.hostname,
			GPUID:             gpuIndex,
			Timestamp:         timestamp,
			GPUName:           gpuName,
			Temperature:       temperature,
			FreeMemory:        freeMemory,
			UsedMemory:        usedMemory,
			TotalMemory:       totalMemory,
			GPUUtilization:    gpuUtil,
			MemoryUtilization: memUtil,
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (c *Collector) parseFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "N/A" || s == "[Not Supported]" {
		return 0.0, nil
	}

	s = strings.ReplaceAll(s, "%", "")
	s = strings.ReplaceAll(s, "℃", "")
	s = strings.ReplaceAll(s, "W", "")
	s = strings.TrimSpace(s)

	return strconv.ParseFloat(s, 64)
}

func (c *Collector) parseUint64(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "N/A" || s == "[Not Supported]" {
		return 0, nil
	}

	s = strings.ReplaceAll(s, "MiB", "")
	s = strings.ReplaceAll(s, "GiB", "")
	s = strings.ReplaceAll(s, "KiB", "")
	s = strings.ReplaceAll(s, "MB", "")
	s = strings.ReplaceAll(s, "GB", "")
	s = strings.ReplaceAll(s, "KB", "")
	s = strings.TrimSpace(s)

	return strconv.ParseUint(s, 10, 64)
}

// CollectProcesses collects GPU process information.
func (c *Collector) CollectProcesses() ([]types.GPUProcess, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout*2)
	defer cancel()

	script := `nvidia-smi --query-compute-apps=timestamp,index,pid,process_name,used_gpu_memory --format=csv,noheader | grep -v 'Not Found' | while read s; do echo $s | sed -z 's/\n//'; echo $(ps --noheader -o 'user,%mem,%cpu,command' -p $(echo $s | awk 'BEGIN{FS=", "}{print $3}') | sed -e 's/,/./g' | awk '{printf(",%s,%s,%s, ",$1,$2,$3);for(i=4;i<NF;i++){printf("%s ",$i)}print $NF}'); done`

	cmd := exec.CommandContext(ctx, "bash", "-c", script)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get GPU process information: %w", err)
	}

	return c.parseProcesses(string(output))
}

func (c *Collector) parseProcesses(output string) ([]types.GPUProcess, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return []types.GPUProcess{}, nil
	}

	processes := make([]types.GPUProcess, 0)

	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Split(line, ",")
		if len(fields) < 4 {
			continue
		}

		timestampStr := strings.TrimSpace(fields[0])
		timestamp, err := time.Parse("2006/01/02 15:04:05.000", timestampStr)
		if err != nil {
			timestamp = time.Now()
		}

		gpuIDStr := strings.TrimSpace(fields[1])
		gpuID, err := strconv.Atoi(gpuIDStr)
		if err != nil {
			continue
		}

		pid, err := strconv.Atoi(fields[2])
		if err != nil {
			continue
		}

		processName := strings.TrimSpace(fields[3])
		if processName == "" {
			processName = "unknown"
		}

		usedGPUMemory, err := c.parseUint64(fields[4])
		if err != nil {
			continue
		}

		user := strings.TrimSpace(fields[5])
		if user == "" {
			user = "unknown"
		}

		usedMemory, err := c.parseFloat(fields[6])
		if err != nil {
			continue
		}

		usedCPU, err := c.parseFloat(fields[7])
		if err != nil {
			continue
		}

		process := types.GPUProcess{
			Hostname:      c.hostname,
			GPUID:         gpuID,
			Timestamp:     timestamp,
			User:          user,
			PID:           pid,
			ProcessName:   processName,
			UsedGPUMemory: usedGPUMemory,
			UsedCPU:       usedCPU,
			UsedMemory:    usedMemory,
			Command:       processName,
		}
		processes = append(processes, process)
	}

	return processes, nil
}
