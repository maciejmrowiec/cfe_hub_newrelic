package main

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
)

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
	return "s"
}

func (a *AverageBenchmark) GetName() string {
	return a.path
}

func (a *AverageBenchmark) GetValue() (float64, error) {
	query := fmt.Sprintf(`
        SELECT avg(averagevalue)
        FROM __benchmarkslog
        WHERE checktimestamp > NOW() - INTERVAL '%d Seconds'
        AND eventname  = $1`, a.interval)

	var value sql.NullFloat64
	if err := a.db.QueryRow(query, a.name).Scan(&value); err != nil {
		return 0, err
	}

	if !value.Valid {
		return 0, errors.New("NullValue")
	}

	return value.Float64, nil
}
