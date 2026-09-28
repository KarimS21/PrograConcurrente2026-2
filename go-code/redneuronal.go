package gocode

import (
	"errors"
	"math"
	"math/rand"
	"time"
)

const (
	InputSize   = 9
	Hidden1Size = 16
	Hidden2Size = 8
)

// TaxiTrip contains the fields used to build the model input.
type TaxiTrip struct {
	PickupTime     time.Time
	DropoffTime    time.Time
	PassengerCount float64
	TripDistance   float64
	RatecodeID     float64
	PULocationID   float64
	DOLocationID   float64
	PaymentType    float64
	FareAmount     float64
}

// Features returns the nine numeric features agreed for the regression model.
func (trip TaxiTrip) Features() [InputSize]float64 {
	tripTimeMin := trip.DropoffTime.Sub(trip.PickupTime).Minutes()
	return [InputSize]float64{
		trip.PassengerCount,
		trip.TripDistance,
		tripTimeMin,
		trip.RatecodeID,
		trip.PULocationID,
		trip.DOLocationID,
		trip.PaymentType,
		float64(trip.PickupTime.Hour()),
		float64(trip.PickupTime.Weekday()),
	}
}

type Sample struct {
	Features [InputSize]float64
	Target   float64
}

type Standardizer struct {
	Mean  [InputSize]float64
	Scale [InputSize]float64
}

func FitStandardizer(samples []Sample) (Standardizer, error) {
	if len(samples) == 0 {
		return Standardizer{}, errors.New("cannot fit a standardizer with no samples")
	}

	var standardizer Standardizer
	for _, sample := range samples {
		for feature := 0; feature < InputSize; feature++ {
			standardizer.Mean[feature] += sample.Features[feature]
		}
	}
	for feature := 0; feature < InputSize; feature++ {
		standardizer.Mean[feature] /= float64(len(samples))
	}
	for _, sample := range samples {
		for feature := 0; feature < InputSize; feature++ {
			difference := sample.Features[feature] - standardizer.Mean[feature]
			standardizer.Scale[feature] += difference * difference
		}
	}
	for feature := 0; feature < InputSize; feature++ {
		standardizer.Scale[feature] = math.Sqrt(standardizer.Scale[feature] / float64(len(samples)))
		if standardizer.Scale[feature] == 0 {
			standardizer.Scale[feature] = 1
		}
	}
	return standardizer, nil
}

func (standardizer Standardizer) Transform(samples []Sample) []Sample {
	transformed := make([]Sample, len(samples))
	for index, sample := range samples {
		transformed[index].Target = sample.Target
		for feature := 0; feature < InputSize; feature++ {
			transformed[index].Features[feature] = (sample.Features[feature] - standardizer.Mean[feature]) / standardizer.Scale[feature]
		}
	}
	return transformed
}

type NeuralNetwork struct {
	weights1 [Hidden1Size][InputSize]float64
	biases1  [Hidden1Size]float64
	weights2 [Hidden2Size][Hidden1Size]float64
	biases2  [Hidden2Size]float64
	weights3 [Hidden2Size]float64
	bias3    float64
}

func NewNeuralNetwork(seed int64) *NeuralNetwork {
	network := &NeuralNetwork{}
	random := rand.New(rand.NewSource(seed))
	for neuron := 0; neuron < Hidden1Size; neuron++ {
		for feature := 0; feature < InputSize; feature++ {
			network.weights1[neuron][feature] = random.NormFloat64() * math.Sqrt(2/float64(InputSize))
		}
	}
	for neuron := 0; neuron < Hidden2Size; neuron++ {
		for previous := 0; previous < Hidden1Size; previous++ {
			network.weights2[neuron][previous] = random.NormFloat64() * math.Sqrt(2/float64(Hidden1Size))
		}
		network.weights3[neuron] = random.NormFloat64() * math.Sqrt(2/float64(Hidden2Size))
	}
	return network
}

func relu(value float64) float64 {
	if value > 0 {
		return value
	}
	return 0
}

func (network *NeuralNetwork) forward(features [InputSize]float64) ([Hidden1Size]float64, [Hidden2Size]float64, float64) {
	var hidden1 [Hidden1Size]float64
	for neuron := 0; neuron < Hidden1Size; neuron++ {
		value := network.biases1[neuron]
		for feature := 0; feature < InputSize; feature++ {
			value += network.weights1[neuron][feature] * features[feature]
		}
		hidden1[neuron] = relu(value)
	}

	var hidden2 [Hidden2Size]float64
	for neuron := 0; neuron < Hidden2Size; neuron++ {
		value := network.biases2[neuron]
		for previous := 0; previous < Hidden1Size; previous++ {
			value += network.weights2[neuron][previous] * hidden1[previous]
		}
		hidden2[neuron] = relu(value)
	}

	output := network.bias3
	for neuron := 0; neuron < Hidden2Size; neuron++ {
		output += network.weights3[neuron] * hidden2[neuron]
	}
	return hidden1, hidden2, output
}

func (network *NeuralNetwork) Predict(features [InputSize]float64) float64 {
	_, _, prediction := network.forward(features)
	return prediction
}

func (network *NeuralNetwork) MSE(samples []Sample) float64 {
	if len(samples) == 0 {
		return 0
	}
	var loss float64
	for _, sample := range samples {
		difference := network.Predict(sample.Features) - sample.Target
		loss += difference * difference
	}
	return loss / float64(len(samples))
}

// Train performs full-batch gradient descent and returns the MSE per epoch.
func (network *NeuralNetwork) Train(samples []Sample, epochs int, learningRate float64) ([]float64, error) {
	if len(samples) == 0 {
		return nil, errors.New("cannot train with no samples")
	}
	if epochs <= 0 || learningRate <= 0 {
		return nil, errors.New("epochs and learning rate must be positive")
	}

	lossHistory := make([]float64, 0, epochs)
	for epoch := 0; epoch < epochs; epoch++ {
		var gradientW1 [Hidden1Size][InputSize]float64
		var gradientB1 [Hidden1Size]float64
		var gradientW2 [Hidden2Size][Hidden1Size]float64
		var gradientB2 [Hidden2Size]float64
		var gradientW3 [Hidden2Size]float64
		var gradientB3 float64

		for _, sample := range samples {
			hidden1, hidden2, prediction := network.forward(sample.Features)
			errorGradient := 2 * (prediction - sample.Target)
			gradientB3 += errorGradient
			for neuron := 0; neuron < Hidden2Size; neuron++ {
				gradientW3[neuron] += errorGradient * hidden2[neuron]
			}

			var hidden2Gradient [Hidden2Size]float64
			for neuron := 0; neuron < Hidden2Size; neuron++ {
				if hidden2[neuron] > 0 {
					hidden2Gradient[neuron] = errorGradient * network.weights3[neuron]
				}
				gradientB2[neuron] += hidden2Gradient[neuron]
				for previous := 0; previous < Hidden1Size; previous++ {
					gradientW2[neuron][previous] += hidden2Gradient[neuron] * hidden1[previous]
				}
			}

			for previous := 0; previous < Hidden1Size; previous++ {
				var gradient float64
				for neuron := 0; neuron < Hidden2Size; neuron++ {
					gradient += hidden2Gradient[neuron] * network.weights2[neuron][previous]
				}
				if hidden1[previous] > 0 {
					gradientB1[previous] += gradient
					for feature := 0; feature < InputSize; feature++ {
						gradientW1[previous][feature] += gradient * sample.Features[feature]
					}
				}
			}
		}

		scale := learningRate / float64(len(samples))
		for neuron := 0; neuron < Hidden1Size; neuron++ {
			network.biases1[neuron] -= scale * gradientB1[neuron]
			for feature := 0; feature < InputSize; feature++ {
				network.weights1[neuron][feature] -= scale * gradientW1[neuron][feature]
			}
		}
		for neuron := 0; neuron < Hidden2Size; neuron++ {
			network.biases2[neuron] -= scale * gradientB2[neuron]
			network.weights3[neuron] -= scale * gradientW3[neuron]
			for previous := 0; previous < Hidden1Size; previous++ {
				network.weights2[neuron][previous] -= scale * gradientW2[neuron][previous]
			}
		}
		network.bias3 -= scale * gradientB3
		lossHistory = append(lossHistory, network.MSE(samples))
	}
	return lossHistory, nil
}
