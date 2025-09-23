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

	// GPU UUIDからGPU IDへのマッピングを取得
	gpuMapping, err := c.getGPUMapping(ctx)
	if err != nil {
		return []types.GPUProcess{}, fmt.Errorf("failed to get GPU mapping: %w", err)
	}

	// get detailed process information
	script := `
	set -euo pipefail
	
	# get GPU process information from nvidia-smi
	nvidia_output=$(nvidia-smi --query-compute-apps=timestamp,gpu_uuid,pid,process_name,used_gpu_memory --format=csv,noheader 2>/dev/null || echo "")
	
	if [ -z "$nvidia_output" ]; then
		exit 0  # if no processes, exit
	fi
	
	echo "$nvidia_output" | while IFS=',' read -r timestamp gpu_uuid pid process_name gpu_memory; do
		# triming values
		timestamp=$(echo "$timestamp" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
		gpu_uuid=$(echo "$gpu_uuid" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
		pid=$(echo "$pid" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
		process_name=$(echo "$process_name" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
		gpu_memory=$(echo "$gpu_memory" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
		
		# check if pid is a number
		if ! echo "$pid" | grep -q '^[0-9]\+$'; then
			continue
		fi
		
		# get process detailed information from ps command (ignore errors)
		if ps_info=$(ps --noheader -o 'user,%mem,%cpu,command' -p "$pid" 2>/dev/null); then
			# Parse ps output into 4 fields: user %mem %cpu command
			# Use awk to properly split the ps output
			user=$(echo "$ps_info" | awk '{print $1}')
			mem_percent=$(echo "$ps_info" | awk '{print $2}')
			cpu_percent=$(echo "$ps_info" | awk '{print $3}')
			# Command is everything from field 4 onwards, joined with spaces
			command=$(echo "$ps_info" | awk '{for(i=4;i<=NF;i++) printf "%s%s", $i, (i<NF?" ":""); print ""}')
			
			# CSV escape: replace commas with dots in command to avoid CSV issues
			command_escaped=$(echo "$command" | sed 's/,/./g')
			
			echo "$timestamp,$gpu_uuid,$pid,$process_name,$gpu_memory,$user,$mem_percent,$cpu_percent,$command_escaped"
		else
			# if process not found, use default value
			echo "$timestamp,$gpu_uuid,$pid,$process_name,$gpu_memory,unknown,0.0,0.0,$process_name"
		fi
	done`

	cmd := exec.CommandContext(ctx, "bash", "-c", script)

	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return []types.GPUProcess{}, fmt.Errorf("process collection script failed with exit code %d: %s", exitError.ExitCode(), string(exitError.Stderr))
		}
		return []types.GPUProcess{}, fmt.Errorf("failed to execute process collection script: %w", err)
	}

	return c.parseProcessesWithMapping(string(output), gpuMapping)
}

// getGPUMapping gets the mapping from GPU UUID to GPU index.
func (c *Collector) getGPUMapping(ctx context.Context) (map[string]int, error) {
	cmd := exec.CommandContext(ctx, c.config.NvidiaSmiPath,
		"--query-gpu=index,gpu_uuid",
		"--format=csv,noheader,nounits")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get GPU mapping: %w", err)
	}

	mapping := make(map[string]int)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) != 2 {
			continue
		}

		indexStr := strings.TrimSpace(parts[0])
		uuid := strings.TrimSpace(parts[1])

		index, err := strconv.Atoi(indexStr)
		if err != nil {
			continue
		}

		mapping[uuid] = index
	}

	return mapping, nil
}

// parseProcessesWithMapping parses process output with GPU UUID to ID mapping.
func (c *Collector) parseProcessesWithMapping(output string, gpuMapping map[string]int) ([]types.GPUProcess, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return []types.GPUProcess{}, nil
	}

	processes := make([]types.GPUProcess, 0)

	for lineNum, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Split(line, ",")
		if len(fields) != 9 {
			return nil, fmt.Errorf("line %d: invalid field count (%d), expected exactly 9 fields", lineNum+1, len(fields))
		}

		timestampStr := strings.TrimSpace(fields[0])
		timestamp, err := time.Parse("2006/01/02 15:04:05.000", timestampStr)
		if err != nil {
			return nil, fmt.Errorf("line %d: failed to parse timestamp '%s': %w", lineNum+1, timestampStr, err)
		}

		gpuUUID := strings.TrimSpace(fields[1])
		gpuID, exists := gpuMapping[gpuUUID]
		if !exists {
			// フォールバック：UUIDが見つからない場合は0を使用
			gpuID = 0
		}

		pidStr := strings.TrimSpace(fields[2])
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			return nil, fmt.Errorf("line %d: failed to parse PID '%s': %w", lineNum+1, pidStr, err)
		}

		processName := strings.TrimSpace(fields[3])
		if processName == "" {
			return nil, fmt.Errorf("line %d: empty process name", lineNum+1)
		}

		usedGPUMemory, err := c.parseUint64(fields[4])
		if err != nil {
			return nil, fmt.Errorf("line %d: failed to parse GPU memory '%s': %w", lineNum+1, fields[4], err)
		}

		uid := strings.TrimSpace(fields[5])
		if uid == "" {
			return nil, fmt.Errorf("line %d: empty user field", lineNum+1)
		}

		usedMemory, err := c.parseFloat(fields[6])
		if err != nil {
			return nil, fmt.Errorf("line %d: failed to parse memory usage '%s': %w", lineNum+1, fields[6], err)
		}

		usedCPU, err := c.parseFloat(fields[7])
		if err != nil {
			return nil, fmt.Errorf("line %d: failed to parse CPU usage '%s': %w", lineNum+1, fields[7], err)
		}

		// 9番目のフィールドが完全なコマンド（スペース含む）
		command := strings.TrimSpace(fields[8])
		if command == "" {
			command = processName // フォールバック
		}

		// コマンドが異常に長い場合は切り詰め（セキュリティ考慮）
		const maxCommandLength = 1024
		if len(command) > maxCommandLength {
			command = command[:maxCommandLength] + "..."
		}

		process := types.GPUProcess{
			Hostname:      c.hostname,
			GPUID:         gpuID,
			Timestamp:     timestamp,
			User:          uid,
			PID:           pid,
			ProcessName:   processName,
			UsedGPUMemory: usedGPUMemory,
			UsedCPU:       usedCPU,
			UsedMemory:    usedMemory,
			Command:       command,
		}
		processes = append(processes, process)
	}

	return processes, nil
}
