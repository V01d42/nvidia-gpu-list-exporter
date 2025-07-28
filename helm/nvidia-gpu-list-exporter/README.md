# NVIDIA GPU List Exporter Helm Chart

A Helm chart for deploying the NVIDIA GPU List Prometheus exporter as a DaemonSet on Kubernetes clusters with NVIDIA GPUs.

## Overview

This chart deploys the NVIDIA GPU List Exporter to collect comprehensive GPU metrics from NVIDIA GPUs using `nvidia-smi`. The exporter provides Prometheus-compatible metrics for GPU temperature, utilization, memory usage, and process information.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- NVIDIA GPU nodes with drivers installed
- `nvidia-smi` command available on GPU nodes
- Prometheus Operator (optional, for ServiceMonitor support)

## Installation

### Install from Helm Repository (Recommended)

```bash
# 1. Add the official Helm repository
helm repo add nvidia-gpu-exporter https://V01d42.github.io/nvidia-gpu-list-exporter
helm repo update

# 2. Install with default values
helm install nvidia-gpu-exporter nvidia-gpu-exporter/nvidia-gpu-list-exporter

# 3. Install in specific namespace
helm install nvidia-gpu-exporter nvidia-gpu-exporter/nvidia-gpu-list-exporter \
  --namespace monitoring --create-namespace

# 4. Install with specific version
helm install nvidia-gpu-exporter nvidia-gpu-exporter/nvidia-gpu-list-exporter \
  --version 1.0.0

# 5. Verify installation
helm list
kubectl get pods -l app.kubernetes.io/name=nvidia-gpu-list-exporter
```

### Install from Local Chart (Development)

```bash
# From the project root directory
helm install nvidia-gpu-exporter ./helm/nvidia-gpu-list-exporter

# Install with custom namespace
helm install nvidia-gpu-exporter ./helm/nvidia-gpu-list-exporter \
  --namespace monitoring --create-namespace

# Install with custom values
helm install nvidia-gpu-exporter ./helm/nvidia-gpu-list-exporter \
  --set image.tag=latest \
  --set exporter.interval=30 \
  --set monitoring.serviceMonitor.enabled=true
```

### Install with Values File

```bash
# Create custom values file
cat > custom-values.yaml << EOF
# Override resource names
nameOverride: "gpu-exporter"
namespaceOverride: "monitoring"

# Image configuration
image:
  repository: your-registry.com/nvidia-gpu-list-exporter
  tag: "1.0.0"

# Server configuration
host: "0.0.0.0"
port: 9090

# Exporter configuration
exporter:
  logLevel: debug
  interval: 30
  timeout: 15s

# Monitoring configuration
monitoring:
  serviceMonitor:
    enabled: true
    interval: 60s

# Resource limits
resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 200m
    memory: 256Mi
EOF

helm install nvidia-gpu-exporter ./helm/nvidia-gpu-list-exporter -f custom-values.yaml
```

## Configuration

### Key Parameters

| Parameter | Description | Default | Type |
|-----------|-------------|---------|------|
| `nameOverride` | Override chart name | `""` | string |
| `fullnameOverride` | Override full resource names | `""` | string |
| `namespaceOverride` | Override namespace | `""` | string |
| `image.repository` | Container image repository | `nvidia-gpu-list-exporter` | string |
| `image.tag` | Container image tag | `1.0.0` | string |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` | string |
| `host` | HTTP server bind address | `0.0.0.0` | string |
| `port` | HTTP server port | `8080` | int |
| `exporter.logLevel` | Log level | `info` | string |
| `exporter.interval` | Metrics update interval (seconds) | `15` | int |
| `exporter.timeout` | nvidia-smi command timeout | `10s` | duration |
| `exporter.nvidiaSmiPath` | Custom nvidia-smi path | `""` | string |
| `exporter.hostnameOverride` | Override hostname | `""` | string |
| `nodeSelector` | Node selector for GPU nodes | `{}` | object |
| `tolerations` | Pod tolerations | `{}` | object |
| `affinity` | Pod affinity rules | `{}` | object |
| `resources.requests.cpu` | CPU request | `100m` | string |
| `resources.requests.memory` | Memory request | `128Mi` | string |
| `resources.limits.cpu` | CPU limit | `200m` | string |
| `resources.limits.memory` | Memory limit | `256Mi` | string |
| `serviceAccount.create` | Create ServiceAccount | `true` | bool |
| `serviceAccount.name` | ServiceAccount name | `""` | string |
| `serviceAccount.annotations` | ServiceAccount annotations | `{}` | object |
| `monitoring.serviceMonitor.enabled` | Create ServiceMonitor | `true` | bool |
| `monitoring.serviceMonitor.interval` | Scrape interval | `30s` | duration |
| `monitoring.serviceMonitor.scrapeTimeout` | Scrape timeout | `10s` | duration |
| `updateStrategy.type` | DaemonSet update strategy | `RollingUpdate` | string |
| `hostNetwork` | Use host network | `false` | bool |
| `hostPID` | Use host PID namespace | `false` | bool |

## GPU Node Targeting

Since this is deployed as a DaemonSet, it will run on nodes that match the selection criteria.

### Basic Node Selection

```yaml
# Deploy to all nodes (not recommended for mixed clusters)
nodeSelector: {}

# Deploy to nodes with specific labels
nodeSelector:
  nvidia.com/gpu.present: "true"
```

### Tolerations for GPU Nodes

Many GPU clusters taint GPU nodes to prevent non-GPU workloads from scheduling there:

```yaml
tolerations:
  - key: nvidia.com/gpu
    operator: Exists
    effect: NoSchedule
  - key: dedicated
    operator: Equal
    value: gpu
    effect: NoSchedule
```

### Advanced Node Selection with Affinity

```yaml
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
      - matchExpressions:
        - key: accelerator
          operator: In
          values:
          - nvidia-tesla-v100
          - nvidia-tesla-a100
          - nvidia-rtx-a6000
```

## Prometheus Integration

### ServiceMonitor (Prometheus Operator)

When using Prometheus Operator, enable ServiceMonitor creation:

```yaml
monitoring:
  serviceMonitor:
    enabled: true
    interval: 30s
    scrapeTimeout: 10s
    path: /metrics
    honorLabels: true
    additionalLabels:
      release: prometheus-operator
```

### Manual Prometheus Configuration

For manual Prometheus setup, add this job to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'nvidia-gpu-exporter'
    kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
          - default  # Adjust to your namespace
    relabel_configs:
      - source_labels: [__meta_kubernetes_service_name]
        action: keep
        regex: nvidia-gpu-list-exporter
      - source_labels: [__meta_kubernetes_endpoint_port_name]
        action: keep
        regex: metrics
    scrape_interval: 30s
    metrics_path: /metrics
```

## Available Metrics

The exporter provides these Prometheus metrics:

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
| `nvidia_gpu_process_count` | Gauge | Number of GPU processes per GPU |

### System Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `nvidia_system_image_info` | Gauge | System information (always 1) |

### Common Labels

All metrics include these labels:
- `hostname`: Node hostname
- `gpu_id`: GPU identifier
- `gpu_uuid`: GPU UUID
- `gpu_name`: GPU model name

Process metrics include additional labels:
- `pid`: Process ID
- `user`: Process owner
- `command`: Process command

## Example Queries

### PromQL Examples

```promql
# Average GPU temperature across cluster
avg(nvidia_gpu_temperature_celsius)

# GPU memory utilization percentage
(nvidia_gpu_memory_used_bytes / nvidia_gpu_memory_total_bytes) * 100

# GPU utilization by node
avg(nvidia_gpu_utilization_percent) by (hostname)

# Top GPU memory consumers
topk(10, nvidia_gpu_process_memory_bytes)

# Count of GPUs per node
count(nvidia_gpu_temperature_celsius) by (hostname)
```

## Security Configuration

### Host Process Access

This exporter requires access to host processes to collect detailed GPU process information. The following security measures are implemented:

#### Required Settings
- `hostPID: true` - Required for accessing host process information
- **Minimal Linux Capabilities**: Only `SYS_PTRACE` for process information reading
- **Explicitly Dropped Capabilities**: `KILL`, `SYS_KILL`, `SYS_ADMIN` to prevent process manipulation

#### Security Features
- **Read-Only Root Filesystem**: `readOnlyRootFilesystem: true`
- **Non-Root User**: Runs as user 1000 with no privilege escalation
- **Seccomp Profile**: Uses `RuntimeDefault` profile
- **Environment Isolation**: Restricted environment variables
- **Command Length Limits**: Prevents buffer overflow attacks

#### Security Best Practices
```yaml
securityContext:
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: true
  runAsNonRoot: true
  runAsUser: 1000
  runAsGroup: 1000
  seccompProfile:
    type: RuntimeDefault
  capabilities:
    add:
      - SYS_PTRACE  # Required for process information only
    drop:
      - ALL
      - KILL        # Explicitly prevent process killing
      - SYS_KILL
      - SYS_ADMIN
```

### Service Account

A dedicated service account is created with minimal permissions:

```yaml
serviceAccount:
  create: true
  annotations: {}
  name: ""  # Auto-generated if empty
```

## Troubleshooting

### Common Issues

1. **No metrics collected**
   - Verify nvidia-smi is available: `kubectl exec -it <pod> -- nvidia-smi`
   - Check GPU drivers are installed on nodes
   - Verify tolerations allow scheduling on GPU nodes

2. **Pod scheduling issues**
   - Check node selectors and tolerations
   - Verify GPU nodes are properly labeled
   - Review DaemonSet status: `kubectl describe daemonset nvidia-gpu-list-exporter`

3. **Permission errors**
   - Ensure nvidia-smi is executable in container
   - Check security context settings
   - Verify no additional security policies block execution

### Debugging Commands

```bash
# Check DaemonSet status
kubectl get daemonset nvidia-gpu-list-exporter

# View pods across nodes
kubectl get pods -o wide -l app.kubernetes.io/name=nvidia-gpu-list-exporter

# Check logs
kubectl logs -l app.kubernetes.io/name=nvidia-gpu-list-exporter --tail=100

# Test nvidia-smi in pod
kubectl exec -it <pod-name> -- nvidia-smi

# Test metrics endpoint (using configured port)
export POD_NAME=$(kubectl get pods -l app.kubernetes.io/name=nvidia-gpu-list-exporter -o jsonpath="{.items[0].metadata.name}")
kubectl port-forward $POD_NAME 8080:8080
curl http://localhost:8080/metrics

# For custom port configurations
# kubectl port-forward $POD_NAME <local-port>:<configured-port>
```

## Upgrading

```bash
# Upgrade with new values
helm upgrade nvidia-gpu-exporter ./helm/nvidia-gpu-list-exporter \
  --set image.tag=1.1.0

# Upgrade with values file
helm upgrade nvidia-gpu-exporter ./helm/nvidia-gpu-list-exporter \
  -f custom-values.yaml
```

## Uninstalling

```bash
# Remove the release
helm uninstall nvidia-gpu-exporter

# Clean up any remaining resources
kubectl delete servicemonitor nvidia-gpu-list-exporter
```

## Development

### Local Testing

```bash
# Template rendering
helm template nvidia-gpu-exporter ./helm/nvidia-gpu-list-exporter

# Dry run installation
helm install nvidia-gpu-exporter ./helm/nvidia-gpu-list-exporter --dry-run

# Lint chart
helm lint ./helm/nvidia-gpu-list-exporter
```

## License

This chart is licensed under the Apache License 2.0. 