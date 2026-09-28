package gocode

import (
	"testing"
	"time"
)

func TestTrainingReducesMSE(t *testing.T) {
	baseTime := time.Date(2026, time.January, 5, 8, 0, 0, 0, time.UTC)
	samples := make([]Sample, 0, 12)
	for index := 0; index < 12; index++ {
		trip := TaxiTrip{
			PickupTime:     baseTime.Add(time.Duration(index) * time.Hour),
			DropoffTime:    baseTime.Add(time.Duration(index)*time.Hour + 20*time.Minute),
			PassengerCount: float64(1 + index%3),
			TripDistance:   float64(index + 1),
			RatecodeID:     1,
			PULocationID:   float64(10 + index),
			DOLocationID:   float64(20 + index),
			PaymentType:    1,
		}
		features := trip.Features()
		samples = append(samples, Sample{Features: features, Target: 5 + 2*features[1] + features[2]})
	}

	standardizer, err := FitStandardizer(samples)
	if err != nil {
		t.Fatal(err)
	}
	normalized := standardizer.Transform(samples)
	network := NewNeuralNetwork(42)
	initialMSE := network.MSE(normalized)
	if _, err := network.Train(normalized, 500, 0.01); err != nil {
		t.Fatal(err)
	}
	if network.MSE(normalized) >= initialMSE {
		t.Fatalf("training did not reduce MSE: initial=%f final=%f", initialMSE, network.MSE(normalized))
	}
}
