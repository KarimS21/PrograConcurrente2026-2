package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	gocode "tp-concurrente/go-code"
)

type measurement struct {
	Run           int
	Mode          string
	Workers       int
	Duration      time.Duration
	MAE           float64
	MSE           float64
	PeakHeap      uint64
	PeakSys       uint64
	MaxGoroutines int
}

type runtimeSnapshot struct {
	peakHeap      uint64
	peakSys       uint64
	maxGoroutines int
}

func main() {
	csvPath := flag.String("csv", "processed/yellow_tripdata_2026-01.csv", "path to the processed CSV dataset")
	maxRows := flag.Int("max-rows", 100000, "maximum rows to load; use 0 for the complete dataset")
	trainRows := flag.Int("train-rows", 10000, "number of initial rows used for training")
	epochs := flag.Int("epochs", 1, "training epochs")
	learningRate := flag.Float64("learning-rate", 0.001, "training learning rate")
	workerText := flag.String("workers", "1,2,4,8", "comma-separated worker counts")
	runs := flag.Int("runs", 11, "measured executions per mode")
	trimFraction := flag.Float64("trim", 0.2, "fraction removed from each side for the trimmed mean")
	output := flag.String("output", "results/inference_runs.csv", "CSV output path")
	flag.Parse()

	workerCounts, err := parseWorkers(*workerText)
	if err != nil {
		log.Fatal(err)
	}
	if *runs < 3 || *trimFraction < 0 || 2*int(float64(*runs)*(*trimFraction)) >= *runs {
		log.Fatal("runs must be at least 3 and trim must leave observations on both sides")
	}

	samples, err := gocode.LoadSamplesFromCSV(*csvPath, *maxRows)
	if err != nil {
		log.Fatal(err)
	}
	if *trainRows <= 0 || *trainRows >= len(samples) {
		log.Fatalf("train-rows must be between 1 and %d", len(samples)-1)
	}
	standardizer, err := gocode.FitStandardizer(samples[:*trainRows])
	if err != nil {
		log.Fatal(err)
	}
	normalized := standardizer.Transform(samples)
	network := gocode.NewNeuralNetwork(42)
	if _, err := network.Train(normalized[:*trainRows], *epochs, *learningRate); err != nil {
		log.Fatal(err)
	}
	evaluation := samples[*trainRows:]
	normalizedEvaluation := normalized[*trainRows:]

	measurements := make([]measurement, 0, (*runs)*(len(workerCounts)+1))
	for run := 1; run <= *runs; run++ {
		measurements = append(measurements, measure(run, "sequential", 0, func() ([]float64, error) {
			return gocode.PredictSequential(network, normalizedEvaluation), nil
		}, evaluation)...)
		for _, workerCount := range workerCounts {
			count := workerCount
			measurements = append(measurements, measure(run, "worker_pool", count, func() ([]float64, error) {
				return gocode.PredictWorkerPool(network, normalizedEvaluation, count)
			}, evaluation)...)
		}
	}

	if err := writeMeasurements(*output, measurements); err != nil {
		log.Fatal(err)
	}
	printSummary(measurements, *trimFraction)
	fmt.Printf("results written to %s\n", *output)
}

func measure(run int, mode string, workers int, predict func() ([]float64, error), samples []gocode.Sample) []measurement {
	_, _ = predict()
	stopMonitor := make(chan struct{})
	var monitor sync.WaitGroup
	var snapshot runtimeSnapshot
	monitor.Add(1)
	go func() {
		defer monitor.Done()
		ticker := time.NewTicker(time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				var memory runtime.MemStats
				runtime.ReadMemStats(&memory)
				if memory.HeapAlloc > snapshot.peakHeap {
					snapshot.peakHeap = memory.HeapAlloc
				}
				if memory.Sys > snapshot.peakSys {
					snapshot.peakSys = memory.Sys
				}
				if runtime.NumGoroutine() > snapshot.maxGoroutines {
					snapshot.maxGoroutines = runtime.NumGoroutine()
				}
			case <-stopMonitor:
				return
			}
		}
	}()
	start := time.Now()
	predictions, err := predict()
	close(stopMonitor)
	monitor.Wait()
	if err != nil {
		log.Fatal(err)
	}
	duration := time.Since(start)
	metrics, err := gocode.Evaluate(samples, predictions)
	if err != nil {
		log.Fatal(err)
	}
	return []measurement{{Run: run, Mode: mode, Workers: workers, Duration: duration, MAE: metrics.MAE, MSE: metrics.MSE, PeakHeap: snapshot.peakHeap, PeakSys: snapshot.peakSys, MaxGoroutines: snapshot.maxGoroutines}}
}

func parseWorkers(text string) ([]int, error) {
	parts := strings.Split(text, ",")
	workers := make([]int, 0, len(parts))
	for _, part := range parts {
		value, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || value <= 0 {
			return nil, fmt.Errorf("invalid worker count %q", part)
		}
		workers = append(workers, value)
	}
	if len(workers) == 0 {
		return nil, fmt.Errorf("at least one worker count is required")
	}
	return workers, nil
}

func writeMeasurements(path string, measurements []measurement) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	if err := writer.Write([]string{"run", "mode", "workers", "duration_ns", "mae", "mse", "peak_heap_bytes", "peak_sys_bytes", "max_goroutines"}); err != nil {
		return err
	}
	for _, item := range measurements {
		if err := writer.Write([]string{
			strconv.Itoa(item.Run), item.Mode, strconv.Itoa(item.Workers),
			strconv.FormatInt(item.Duration.Nanoseconds(), 10),
			strconv.FormatFloat(item.MAE, 'f', 8, 64),
			strconv.FormatFloat(item.MSE, 'f', 8, 64),
			strconv.FormatUint(item.PeakHeap, 10),
			strconv.FormatUint(item.PeakSys, 10),
			strconv.Itoa(item.MaxGoroutines),
		}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func printSummary(measurements []measurement, trimFraction float64) {
	groups := make(map[string][]measurement)
	for _, item := range measurements {
		key := fmt.Sprintf("%s/%d", item.Mode, item.Workers)
		groups[key] = append(groups[key], item)
	}
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		items := groups[key]
		durations := make([]float64, len(items))
		for index, item := range items {
			durations[index] = float64(item.Duration.Nanoseconds())
		}
		sort.Float64s(durations)
		trim := int(float64(len(durations)) * trimFraction)
		kept := durations[trim : len(durations)-trim]
		mean := 0.0
		for _, duration := range kept {
			mean += duration
		}
		mean /= float64(len(kept))
		variance := 0.0
		for _, duration := range kept {
			variance += (duration - mean) * (duration - mean)
		}
		stddev := math.Sqrt(variance / float64(len(kept)))
		fmt.Printf("%-18s trimmed_mean_ms=%10.3f stddev_ms=%10.3f n=%d\n", key, mean/1e6, stddev/1e6, len(kept))
	}
}
