package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type HostCount struct {
	db   *sql.DB
	path string
}

func NewHostCount(path string, db *sql.DB) *HostCount {
	return &HostCount{
		db:   db,
		path: path,
	}
}

func (h *HostCount) GetUnits() string {
	return "Count"
}

func (h *HostCount) GetName() string {
	return h.path
}

func (h *HostCount) GetValue() (float64, error) {
	query := `SELECT count(*) FROM __hosts`

	var value sql.NullInt64
	if err := h.db.QueryRow(query).Scan(&value); err != nil {
		return 0, err
	}

	return float64(value.Int64), nil
}

type AverageBenchmark struct {
	db       *sql.DB
	path     string
	interval int
	name     string
}

func NewAverageBenchmark(path string, interval int, db *sql.DB, name string) *AverageBenchmark {
	return &AverageBenchmark{
		db:       db,
		path:     path,
		interval: interval,
		name:     name,
	}
}

func (a *AverageBenchmark) GetUnits() string {
	return "ms"
}

func (a *AverageBenchmark) GetName() string {
	return a.path
}

func (a *AverageBenchmark) GetValue() (float64, error) {
	query := fmt.Sprintf(`
		SELECT avg(averagevalue)
		FROM __benchmarkslog
		WHERE checktimestamp > NOW() - INTERVAL '%d Seconds'
		AND eventname  = '%s'`, a.interval, a.name)

	var value sql.NullInt64
	if err := a.db.QueryRow(query).Scan(&value); err != nil {
		return 0, err
	}

	return float64(value.Int64), nil
}
