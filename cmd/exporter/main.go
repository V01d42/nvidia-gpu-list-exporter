package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/nvidia-gpu-list-exporter/internal/collector"
	"github.com/nvidia-gpu-list-exporter/internal/metrics"
	"github.com/nvidia-gpu-list-exporter/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	gpuCollector, err := collector.New(cfg.Collector)
	if err != nil {
		log.Fatalf("Failed to create collector: %v", err)
	}

	promMetrics := metrics.New()

	registry := prometheus.NewRegistry()
	err = promMetrics.Register(registry)
	if err != nil {
		log.Fatalf("Failed to register metrics: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Server.Port),
		Handler: mux,
	}

	go func() {
		ticker := time.NewTicker(time.Duration(cfg.Server.MetricsUpdateInterval) * time.Second)
		defer ticker.Stop()

		for {
			gpuMetrics, err := gpuCollector.CollectGPUMetrics()
			if err != nil {
				log.Printf("Failed to collect GPU metrics: %v", err)
			} else {
				promMetrics.UpdateGPU(gpuMetrics)
				log.Printf("GPU metrics updated: %d items", len(gpuMetrics))
			}

			processes, err := gpuCollector.CollectProcesses()
			if err != nil {
				log.Printf("Failed to collect process information: %v", err)
			} else {
				promMetrics.UpdateProcesses(processes)
				if len(processes) == 0 {
					log.Printf("Process information updated: no GPU processes running")
				} else {
					log.Printf("Process information updated: %d processes", len(processes))
				}
			}

			<-ticker.C
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Server starting on %s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutdown initiated...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}

	log.Println("Server stopped")
}

// handleHealth serves the health check endpoint.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"ok","timestamp":"`+time.Now().Format(time.RFC3339)+`"}`)
}
