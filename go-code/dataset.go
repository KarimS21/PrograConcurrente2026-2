package gocode

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

const csvTimeLayout = "2006-01-02 15:04:05"

var requiredCSVColumns = []string{
	"tpep_pickup_datetime",
	"tpep_dropoff_datetime",
	"passenger_count",
	"trip_distance",
	"trip_time_min",
	"RatecodeID",
	"PULocationID",
	"DOLocationID",
	"payment_type",
	"fare_amount",
}

func LoadSamplesFromCSV(path string, maxRows int) ([]Sample, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open dataset: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read CSV header: %w", err)
	}
	columns := make(map[string]int, len(header))
	for index, column := range header {
		columns[column] = index
	}
	for _, column := range requiredCSVColumns {
		if _, ok := columns[column]; !ok {
			return nil, fmt.Errorf("CSV is missing required column %q", column)
		}
	}

	samples := make([]Sample, 0)
	for rowNumber := 2; maxRows <= 0 || len(samples) < maxRows; rowNumber++ {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read CSV row %d: %w", rowNumber, err)
		}
		if len(record) != len(header) {
			return nil, fmt.Errorf("CSV row %d has %d fields, expected %d", rowNumber, len(record), len(header))
		}

		pickupTime, err := parseCSVTime(record, columns["tpep_pickup_datetime"], rowNumber)
		if err != nil {
			return nil, err
		}
		dropoffTime, err := parseCSVTime(record, columns["tpep_dropoff_datetime"], rowNumber)
		if err != nil {
			return nil, err
		}
		passengerCount, err := parseCSVFloat(record, columns["passenger_count"], "passenger_count", rowNumber)
		if err != nil {
			return nil, err
		}
		tripDistance, err := parseCSVFloat(record, columns["trip_distance"], "trip_distance", rowNumber)
		if err != nil {
			return nil, err
		}
		ratecodeID, err := parseCSVFloat(record, columns["RatecodeID"], "RatecodeID", rowNumber)
		if err != nil {
			return nil, err
		}
		pickupLocation, err := parseCSVFloat(record, columns["PULocationID"], "PULocationID", rowNumber)
		if err != nil {
			return nil, err
		}
		dropoffLocation, err := parseCSVFloat(record, columns["DOLocationID"], "DOLocationID", rowNumber)
		if err != nil {
			return nil, err
		}
		paymentType, err := parseCSVFloat(record, columns["payment_type"], "payment_type", rowNumber)
		if err != nil {
			return nil, err
		}
		fareAmount, err := parseCSVFloat(record, columns["fare_amount"], "fare_amount", rowNumber)
		if err != nil {
			return nil, err
		}

		trip := TaxiTrip{
			PickupTime:     pickupTime,
			DropoffTime:    dropoffTime,
			PassengerCount: passengerCount,
			TripDistance:   tripDistance,
			RatecodeID:     ratecodeID,
			PULocationID:   pickupLocation,
			DOLocationID:   dropoffLocation,
			PaymentType:    paymentType,
			FareAmount:     fareAmount,
		}
		samples = append(samples, Sample{Features: trip.Features(), Target: trip.FareAmount})
	}
	return samples, nil
}

func parseCSVTime(record []string, index, rowNumber int) (time.Time, error) {
	value, err := time.Parse(csvTimeLayout, record[index])
	if err != nil {
		return time.Time{}, fmt.Errorf("row %d: invalid datetime %q: %w", rowNumber, record[index], err)
	}
	return value, nil
}

func parseCSVFloat(record []string, index int, column string, rowNumber int) (float64, error) {
	value, err := strconv.ParseFloat(record[index], 64)
	if err != nil {
		return 0, fmt.Errorf("row %d: invalid %s %q: %w", rowNumber, column, record[index], err)
	}
	return value, nil
}
