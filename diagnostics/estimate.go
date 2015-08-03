package diagnostics

import (
	"database/sql"
	"errors"
	_ "github.com/lib/pq"
)

type EstimatedCapacity struct {
	path                             string
	table_diagnostics                *Diagnostics
	name                             string
	interval                         int
	max_time_per_collection_round_ms int
}

func NewEstimatedCapacity(path string, conn *sql.DB, name string, interval int) *EstimatedCapacity {
	return &EstimatedCapacity{
		path:              path,
		table_diagnostics: NewDiagnostics(conn),
		name:              name,
		interval:          interval,
		max_time_per_collection_round_ms: 300000,
	}
}

func (e *EstimatedCapacity) GetName() string {
	return e.path
}

func (e *EstimatedCapacity) GetUnits() string {
	return "Count"
}

func (e *EstimatedCapacity) getHubEstimation() (float64, error) {
	var hub_avg float64
	var err error

	if hub_avg, err = e.table_diagnostics.GetSampleLocalAverage("hub_processing_time_per_host", e.interval, ""); err != nil {
		return 0, err
	}

	hub_processes := 50.0 // hub uses up to 50 threads

	return float64(e.max_time_per_collection_round_ms) / (hub_avg / hub_processes), nil
}

func (e *EstimatedCapacity) getConsumerEstimation() (float64, error) {
	var consumer_avg float64
	var err error

	if consumer_avg, err = e.table_diagnostics.GetSampleLocalAverage("consumer_processing_time_per_host", e.interval, ""); err != nil {
		return 0, err
	}

	consumer_processes := 25.0 // hub uses up to 50 threads

	return float64(e.max_time_per_collection_round_ms) / (consumer_avg / consumer_processes), nil
}

func (e *EstimatedCapacity) GetValue() (float64, error) {

	var value float64
	var err error = errors.New("Unkown metric")

	switch e.name {
	case "hub":
		value, err = e.getHubEstimation()
	case "consumer":
		value, err = e.getConsumerEstimation()
	}

	return value, err
}
