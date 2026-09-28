package gocode

import (
	"errors"
	"sync"
)

type Metrics struct {
	MAE float64
	MSE float64
}

func PredictSequential(network *NeuralNetwork, samples []Sample) []float64 {
	predictions := make([]float64, len(samples))
	for index, sample := range samples {
		predictions[index] = network.Predict(sample.Features)
	}
	return predictions
}

func PredictWorkerPool(network *NeuralNetwork, samples []Sample, workerCount int) ([]float64, error) {
	if network == nil {
		return nil, errors.New("network cannot be nil")
	}
	if workerCount <= 0 {
		return nil, errors.New("worker count must be positive")
	}
	if len(samples) == 0 {
		return []float64{}, nil
	}
	if workerCount > len(samples) {
		workerCount = len(samples)
	}

	type job struct {
		index  int
		sample Sample
	}
	type result struct {
		index      int
		prediction float64
	}

	jobs := make(chan job)
	results := make(chan result, len(samples))
	var workers sync.WaitGroup
	workers.Add(workerCount)
	for worker := 0; worker < workerCount; worker++ {
		go func() {
			defer workers.Done()
			for task := range jobs {
				results <- result{index: task.index, prediction: network.Predict(task.sample.Features)}
			}
		}()
	}

	go func() {
		for index, sample := range samples {
			jobs <- job{index: index, sample: sample}
		}
		close(jobs)
	}()
	go func() {
		workers.Wait()
		close(results)
	}()

	predictions := make([]float64, len(samples))
	for output := range results {
		predictions[output.index] = output.prediction
	}
	return predictions, nil
}

func Evaluate(samples []Sample, predictions []float64) (Metrics, error) {
	if len(samples) != len(predictions) {
		return Metrics{}, errors.New("samples and predictions must have the same length")
	}
	if len(samples) == 0 {
		return Metrics{}, nil
	}

	var metrics Metrics
	for index, sample := range samples {
		difference := predictions[index] - sample.Target
		if difference < 0 {
			metrics.MAE -= difference
		} else {
			metrics.MAE += difference
		}
		metrics.MSE += difference * difference
	}
	metrics.MAE /= float64(len(samples))
	metrics.MSE /= float64(len(samples))
	return metrics, nil
}
