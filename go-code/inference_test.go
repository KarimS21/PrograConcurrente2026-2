package gocode

import (
	"math"
	"testing"
)

func TestWorkerPoolMatchesSequentialInference(t *testing.T) {
	network := NewNeuralNetwork(7)
	samples := make([]Sample, 32)
	for index := range samples {
		for feature := 0; feature < InputSize; feature++ {
			samples[index].Features[feature] = float64(index+feature) / 10
		}
		samples[index].Target = float64(index) / 2
	}

	sequential := PredictSequential(network, samples)
	concurrent, err := PredictWorkerPool(network, samples, 4)
	if err != nil {
		t.Fatal(err)
	}
	for index := range sequential {
		if sequential[index] != concurrent[index] {
			t.Fatalf("prediction %d differs: sequential=%f concurrent=%f", index, sequential[index], concurrent[index])
		}
	}
}

func TestEvaluate(t *testing.T) {
	samples := []Sample{{Target: 10}, {Target: 20}, {Target: 30}}
	predictions := []float64{8, 23, 29}
	metrics, err := Evaluate(samples, predictions)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(metrics.MAE-2) > 1e-12 || math.Abs(metrics.MSE-14/3.0) > 1e-12 {
		t.Fatalf("unexpected metrics: %+v", metrics)
	}
}

func TestWorkerPoolRejectsInvalidWorkerCount(t *testing.T) {
	_, err := PredictWorkerPool(NewNeuralNetwork(1), []Sample{{}}, 0)
	if err == nil {
		t.Fatal("expected an error for zero workers")
	}
}
