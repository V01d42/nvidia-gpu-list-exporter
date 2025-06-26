# NVIDIA GPU List Exporter

A Prometheus exporter for NVIDIA GPU metrics that provides comprehensive monitoring data for GPU utilization, memory usage, temperature, and process information using `nvidia-smi`.

## Features

- **GPU Metrics**: Temperature, utilization, memory usage (free/used/total)
- **Process Monitoring**: GPU process information including memory usage and user details
- **System Information**: Boot image version, OS version, and kernel version
- **Prometheus Integration**: Native Prometheus metrics format
- **Health Endpoints**: Built-in health check endpoint
- **Docker Support**: Ready-to-use Docker images
- **Configurable**: Environment variables and command-line options
- **Lightweight**: Minimal resource footprint

## Prerequisites

- NVIDIA GPU with driver installed
- `nvidia-smi` command available in PATH
- Go 1.22+ (for building from source)
- Docker (optional, for containerized deployment)

## Installation

### Binary Release

Download the latest binary from the [releases page](https://github.com/your-repo/nvidia-gpu-list-exporter/releases).

### Building from Source

```bash
git clone https://github.com/V01d42/nvidia-gpu-list-exporter.git
cd nvidia-gpu-list-exporter
go build -o exporter ./cmd/exporter
```

### Docker

```bash
docker pull your-registry/nvidia-gpu-list-exporter:latest
```

## Usage

### Command Line

```bash
# Basic usage
./exporter

# Custom configuration
./exporter --port 8080 --interval 15 --host 0.0.0.0

# With custom nvidia-smi path
./exporter --nvidia-smi-path /usr/local/cuda/bin/nvidia-smi
```

### Docker

```bash
# Basic Docker run
docker run -d \
  --name gpu-exporter \
  --gpus all \
  -p 8080:8080 \
  your-registry/nvidia-gpu-list-exporter:latest

# With custom configuration
docker run -d \
  --name gpu-exporter \
  --gpus all \
  -p 8080:8080 \
  -e EXPORTER_PORT=8080 \
  -e EXPORTER_INTERVAL=30 \
  your-registry/nvidia-gpu-list-exporter:latest
```

## Configuration

### Command Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `--host` | HTTP server host | `0.0.0.0` |
| `--port` | HTTP server port | `8080` |
| `--interval` | Metrics update interval (seconds) | `15` |
| `--timeout` | nvidia-smi command timeout | `10s` |
| `--nvidia-smi-path` | Path to nvidia-smi command | `nvidia-smi` |
| `--hostname` | Hostname override | (system hostname) |

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `EXPORTER_HOST` | HTTP server host | `0.0.0.0` |
| `EXPORTER_PORT` | HTTP server port | `8080` |
| `EXPORTER_INTERVAL` | Metrics update interval (seconds) | `15` |
| `EXPORTER_TIMEOUT` | nvidia-smi command timeout | `10s` |
| `NVIDIA_SMI_PATH` | Path to nvidia-smi command | `nvidia-smi` |
| `HOSTNAME_OVERRIDE` | Hostname override | (system hostname) |

## Metrics

The exporter provides the following Prometheus metrics:

### GPU Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `nvidia_gpu_temperature_celsius` | Gauge | GPU temperature in Celsius |
| `nvidia_gpu_utilization_percent` | Gauge | GPU utilization percentage |
| `nvidia_gpu_memory_utilization_percent` | Gauge | GPU memory utilization percentage |
| `nvidia_gpu_memory_free_bytes` | Gauge | GPU free memory in bytes |
| `nvidia_gpu_memory_used_bytes` | Gauge | GPU used memory in bytes |
| `nvidia_gpu_memory_total_bytes` | Gauge | GPU total memory in bytes |

### Process Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `nvidia_gpu_process_memory_bytes` | Gauge | GPU process memory usage in bytes |
| `nvidia_gpu_process_count` | Gauge | Number of GPU processes |

### System Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `nvidia_system_image_info` | Gauge | System image information (value is always 1) |

### Labels

All metrics include the following labels:
- `hostname`: System hostname
- `gpu_id`: GPU identifier
- `gpu_uuid`: GPU UUID
- `gpu_name`: GPU model name

Process metrics include additional labels:
- `pid`: Process ID
- `user`: Process owner
- `command`: Process command

## Endpoints

| Endpoint | Description |
|----------|-------------|
| `/metrics` | Prometheus metrics |
| `/health` | Health check endpoint |

### Health Check Response

```json
{
  "status": "ok",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## Monitoring Setup

### Prometheus Configuration

Add the following to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'nvidia-gpu-exporter'
    static_configs:
      - targets: ['localhost:8080']
    scrape_interval: 15s
    metrics_path: /metrics
```

### Grafana Dashboard

Import the provided Grafana dashboard or create custom visualizations using the available metrics.

## Example Queries

### PromQL Examples

```promql
# Average GPU temperature
avg(nvidia_gpu_temperature_celsius)

# GPU memory utilization percentage
(nvidia_gpu_memory_used_bytes / nvidia_gpu_memory_total_bytes) * 100

# Top GPU processes by memory usage
topk(10, nvidia_gpu_process_memory_bytes)

# GPU utilization over time
rate(nvidia_gpu_utilization_percent[5m])
```

## Troubleshooting

### Common Issues

1. **nvidia-smi not found**
   ```bash
   # Ensure NVIDIA drivers are installed
   nvidia-smi --version
   
   # Or specify custom path
   ./exporter --nvidia-smi-path /usr/local/cuda/bin/nvidia-smi
   ```

2. **Permission denied**
   ```bash
   # Run with appropriate permissions or as root
   sudo ./exporter
   ```

3. **Docker GPU access**
   ```bash
   # Ensure Docker has GPU support
   docker run --rm --gpus all nvidia/cuda:11.0-base nvidia-smi
   ```

### Logging

Enable verbose logging by checking the application logs:

```bash
# View logs
./exporter 2>&1 | tee exporter.log

# Docker logs
docker logs nvidia-gpu-exporter
```

## Development

### Requirements

- Go 1.22+
- NVIDIA GPU with drivers
- Docker (optional)

### Building

```bash
# Build binary
go build -o exporter ./cmd/exporter

# Build Docker image
docker build -t nvidia-gpu-list-exporter:latest .

# Run tests
go test ./...

# Format code
go fmt ./...

# Lint code
go vet ./...
```

### Project Structure

```
.
├── cmd/exporter/          # Main application entry point
├── internal/
│   ├── collector/         # GPU metrics collection logic
│   └── metrics/           # Prometheus metrics management
├── pkg/
│   ├── config/           # Configuration handling
│   └── types/            # Data structures
├── Dockerfile            # Docker build configuration
├── go.mod               # Go module definition
└── README.md            # This file
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Style

- Follow Go conventions and best practices
- Use `go fmt` for formatting
- Add tests for new functionality
- Update documentation as needed

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- NVIDIA for the `nvidia-smi` tool
- Prometheus community for metrics standards
- Go community for excellent tooling

## Support

- Create an issue for bug reports or feature requests
- Check existing issues before creating new ones
- Provide detailed information for faster resolution

---

**Note**: This exporter requires NVIDIA GPUs and drivers to function properly. Ensure your system meets the prerequisites before deployment.
