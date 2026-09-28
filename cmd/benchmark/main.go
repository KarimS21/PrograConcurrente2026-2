package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
	"time"

	gocode "tp-concurrente/go-code"
)

func main() {
	csvPath := flag.String("csv", "processed/yellow_tripdata_2026-01.csv", "path to the processed CSV dataset")
	maxRows := flag.Int("max-rows", 100000, "maximum rows to load; use 0 for the complete dataset")
	trainRows := flag.Int("train-rows", 10000, "number of initial rows used for training")
	epochs := flag.Int("epochs", 20, "training epochs")
	learningRate := flag.Float64("learning-rate", 0.001, "training learning rate")
	workers := flag.Int("workers", runtime.NumCPU(), "number of inference workers")
	flag.Parse()

	loadStart := time.Now()
	samples, err := gocode.LoadSamplesFromCSV(*csvPath, *maxRows)
	if err != nil {
		log.Fatal(err)
	}
	loadDuration := time.Since(loadStart)
	if len(samples) < 2 {
		log.Fatal("dataset must contain at least two rows")
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
	trainingStart := time.Now()
	if _, err := network.Train(normalized[:*trainRows], *epochs, *learningRate); err != nil {
		log.Fatal(err)
	}
	trainingDuration := time.Since(trainingStart)

	evaluation := samples[*trainRows:]
	normalizedEvaluation := normalized[*trainRows:]
	sequentialStart := time.Now()
	sequentialPredictions := gocode.PredictSequential(network, normalizedEvaluation)
	sequentialDuration := time.Since(sequentialStart)
	sequentialMetrics, err := gocode.Evaluate(evaluation, sequentialPredictions)
	if err != nil {
		log.Fatal(err)
	}

	concurrentStart := time.Now()
	concurrentPredictions, err := gocode.PredictWorkerPool(network, normalizedEvaluation, *workers)
	concurrentDuration := time.Since(concurrentStart)
	if err != nil {
		log.Fatal(err)
	}
	concurrentMetrics, err := gocode.Evaluate(evaluation, concurrentPredictions)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("rows loaded: %d\n", len(samples))
	fmt.Printf("training rows: %d\n", *trainRows)
	fmt.Printf("evaluation rows: %d\n", len(evaluation))
	fmt.Printf("workers: %d\n", *workers)
	fmt.Printf("load: %s\n", loadDuration)
	fmt.Printf("training: %s\n", trainingDuration)
	fmt.Printf("sequential inference: %s (MAE %.4f, MSE %.4f)\n", sequentialDuration, sequentialMetrics.MAE, sequentialMetrics.MSE)
	fmt.Printf("worker pool inference: %s (MAE %.4f, MSE %.4f)\n", concurrentDuration, concurrentMetrics.MAE, concurrentMetrics.MSE)
	if concurrentDuration > 0 {
		fmt.Printf("speedup: %.2fx\n", float64(sequentialDuration)/float64(concurrentDuration))
	}
}
