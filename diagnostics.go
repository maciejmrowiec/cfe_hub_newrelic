package main

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
)

const DELTA = "delta"
const REBASE = "rebase"

type LocalAverageDiagnostics struct {
	sample_name       string
	query_type        string
	table_diagnostics *Diagnostics
	interval          int
	subset_path       string
}

type LocalCountDiagnostics struct {
	sample_name       string
	query_type        string
	table_diagnostics *Diagnostics
	interval          int
}

func NewLocalCountDiagnostics(conn *sql.DB, name string, query_type string, interval int) *LocalCountDiagnostics {
	cp := &LocalCountDiagnostics{
		query_type:        query_type,
		table_diagnostics: NewDiagnostics(conn),
		interval:          interval,
		sample_name:       name,
	}

	return cp
}

func NewLocalAverageDiagnostics(conn *sql.DB, name string, query_type string, interval int, subsetpath string) *LocalAverageDiagnostics {
	cp := &LocalAverageDiagnostics{
		query_type:        query_type,
		table_diagnostics: NewDiagnostics(conn),
		interval:          interval,
		sample_name:       name,
		subset_path:       subsetpath,
	}

	return cp
}

func (l *LocalCountDiagnostics) GetName() string {
	if l.query_type != "" {
		return "count/" + l.sample_name + "/" + l.query_type
	}

	return l.sample_name
}

func (l *LocalCountDiagnostics) GetUnits() string {
	return "Count"
}

func (l *LocalCountDiagnostics) GetValue() (float64, error) {
	value, err := l.table_diagnostics.GetSampleLocalCount(l.sample_name, l.interval, l.query_type)
	if err != nil {
		fmt.Println(err.Error())
		return 0, err
	}

	return float64(value), nil
}

func (l *LocalAverageDiagnostics) GetName() string {

	name := "average/"

	if l.subset_path != "" {
		name += l.subset_path + "/"
	}

	name += l.sample_name

	if l.query_type != "" {
		name += "/" + l.query_type
	}

	return name
}

func (l *LocalAverageDiagnostics) GetUnits() string {
	return l.table_diagnostics.GetSampleUnits(l.sample_name)
}

func (l *LocalAverageDiagnostics) GetValue() (float64, error) {
	value, err := l.table_diagnostics.GetSampleLocalAverage(l.sample_name, l.interval, l.query_type)
	if err != nil {
		fmt.Println(err.Error())
		return 0, err
	}

	return value, nil
}

type Diagnostics struct {
	db *sql.DB
}

func NewDiagnostics(db *sql.DB) *Diagnostics {
	return &Diagnostics{
		db: db,
	}
}

func (d *Diagnostics) GetSampleUnits(observable string) string {
	query := `
		SELECT distinct(units)
		FROM diagnostics
		WHERE name = $1
		LIMIT 1`

	var value sql.NullString
	if err := d.db.QueryRow(query, observable).Scan(&value); err != nil {
		fmt.Println(err.Error())
		return "unknown"
	}

	return value.String
}

func (d *Diagnostics) GetSampleLocalCount(observable string, interval int, query_type string) (int64, error) {

	query := fmt.Sprintf(`
		SELECT count(value)
		FROM diagnostics
		WHERE timestamp > NOW() - INTERVAL '%d Seconds'
		AND name = '%s'`, interval, observable)

	if query_type != "" {
		query += " AND details = '" + query_type + "'"
	}

	var value sql.NullInt64
	if err := d.db.QueryRow(query).Scan(&value); err != nil {
		return 0, err
	}

	return value.Int64, nil
}

func (d *Diagnostics) GetSampleLocalAverage(observable string, interval int, query_type string) (float64, error) {

	query := fmt.Sprintf(`
		SELECT avg(value)
		FROM diagnostics
		WHERE timestamp > NOW() - INTERVAL '%d Seconds'
		AND name = '%s'`, interval, observable)

	if query_type != "" {
		query += " AND details = '" + query_type + "'"
	}

	var value sql.NullFloat64
	if err := d.db.QueryRow(query).Scan(&value); err != nil {
		return 0, err
	}

	return value.Float64, nil
}

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
