package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type ConnectionErrorCount struct {
	db       *sql.DB
	path     string
	name     string
	interval int
}

func NewConnectionErrorCount(path string, db *sql.DB, error_name string, interval int) *ConnectionErrorCount {
	return &ConnectionErrorCount{
		db:       db,
		path:     path,
		name:     error_name,
		interval: interval,
	}
}

func (c *ConnectionErrorCount) GetUnits() string {
	return "Count"
}

func (c *ConnectionErrorCount) GetName() string {
	return c.path
}

func (c *ConnectionErrorCount) GetValue() (float64, error) {
	query := fmt.Sprintf(`
		SELECT count(*)
		FROM __hubconnectionerrors
		WHERE checktimestamp > NOW() - INTERVAL '%d Seconds'
		AND message = '%s'`, c.interval, c.name)

	var value sql.NullInt64
	if err := c.db.QueryRow(query).Scan(&value); err != nil {
		return 0, err
	}

	return float64(value.Int64), nil
}
