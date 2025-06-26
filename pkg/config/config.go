// Package config handles application configuration loading and validation.
package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/nvidia-gpu-list-exporter/pkg/types"
)

// ServerConfig represents HTTP server configuration.
type ServerConfig struct {
	Host                  string `json:"host"`
	Port                  int    `json:"port"`
	MetricsUpdateInterval int    `json:"metrics_update_interval"`
}

// Config represents the application configuration.
type Config struct {
	Server    ServerConfig          `json:"server"`
	Collector types.CollectorConfig `json:"collector"`
}

// Load loads the application configuration.
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host:                  "0.0.0.0",
			Port:                  8080,
			MetricsUpdateInterval: 15,
		},
		Collector: types.CollectorConfig{
			Timeout:          10 * time.Second,
			NvidiaSmiPath:    "nvidia-smi",
			HostnameOverride: "",
		},
	}

	flag.StringVar(&cfg.Server.Host, "host", cfg.Server.Host, "HTTP server host")
	flag.IntVar(&cfg.Server.Port, "port", cfg.Server.Port, "HTTP server port")
	flag.IntVar(&cfg.Server.MetricsUpdateInterval, "interval", cfg.Server.MetricsUpdateInterval, "Metrics update interval (seconds)")
	flag.DurationVar(&cfg.Collector.Timeout, "timeout", cfg.Collector.Timeout, "nvidia-smi command timeout")
	flag.StringVar(&cfg.Collector.NvidiaSmiPath, "nvidia-smi-path", cfg.Collector.NvidiaSmiPath, "Path to nvidia-smi command")
	flag.StringVar(&cfg.Collector.HostnameOverride, "hostname", cfg.Collector.HostnameOverride, "Hostname override")

	if host := os.Getenv("EXPORTER_HOST"); host != "" {
		cfg.Server.Host = host
	}
	if port := os.Getenv("EXPORTER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Server.Port = p
		}
	}
	if interval := os.Getenv("EXPORTER_INTERVAL"); interval != "" {
		if i, err := strconv.Atoi(interval); err == nil {
			cfg.Server.MetricsUpdateInterval = i
		}
	}
	if timeout := os.Getenv("EXPORTER_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			cfg.Collector.Timeout = d
		}
	}
	if path := os.Getenv("NVIDIA_SMI_PATH"); path != "" {
		cfg.Collector.NvidiaSmiPath = path
	}
	if hostname := os.Getenv("HOSTNAME_OVERRIDE"); hostname != "" {
		cfg.Collector.HostnameOverride = hostname
	}

	flag.Parse()

	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return nil, fmt.Errorf("invalid port number: %d", cfg.Server.Port)
	}
	if cfg.Server.MetricsUpdateInterval <= 0 {
		return nil, fmt.Errorf("invalid update interval: %d", cfg.Server.MetricsUpdateInterval)
	}
	if cfg.Collector.Timeout <= 0 {
		return nil, fmt.Errorf("invalid timeout: %v", cfg.Collector.Timeout)
	}

	return cfg, nil
}
