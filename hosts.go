package main

import (
	"database/sql"
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
