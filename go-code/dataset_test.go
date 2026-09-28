package gocode

import (
	"path/filepath"
	"testing"
)

func TestLoadSamplesFromProcessedCSV(t *testing.T) {
	path := filepath.Join("..", "processed", "yellow_tripdata_2026-01.csv")
	samples, err := LoadSamplesFromCSV(path, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(samples) != 3 {
		t.Fatalf("expected 3 samples, got %d", len(samples))
	}
	if samples[0].Target != 17 {
		t.Fatalf("unexpected first target: %f", samples[0].Target)
	}
	if samples[0].Features[2] != 10.866666666666667 {
		t.Fatalf("unexpected trip time: %f", samples[0].Features[2])
	}
}

func TestProcessedCSVInferenceFlow(t *testing.T) {
	path := filepath.Join("..", "processed", "yellow_tripdata_2026-01.csv")
	samples, err := LoadSamplesFromCSV(path, 128)
	if err != nil {
		t.Fatal(err)
	}
	standardizer, err := FitStandardizer(samples)
	if err != nil {
		t.Fatal(err)
	}
	normalized := standardizer.Transform(samples)
	network := NewNeuralNetwork(42)

	sequentialPredictions := PredictSequential(network, normalized)
	concurrentPredictions, err := PredictWorkerPool(network, normalized, 4)
	if err != nil {
		t.Fatal(err)
	}
	sequentialMetrics, err := Evaluate(samples, sequentialPredictions)
	if err != nil {
		t.Fatal(err)
	}
	concurrentMetrics, err := Evaluate(samples, concurrentPredictions)
	if err != nil {
		t.Fatal(err)
	}
	if sequentialMetrics != concurrentMetrics {
		t.Fatalf("metrics differ: sequential=%+v concurrent=%+v", sequentialMetrics, concurrentMetrics)
	}
}
