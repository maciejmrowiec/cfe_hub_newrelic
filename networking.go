package main

import (
	"database/sql"
	"errors"
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

	if !value.Valid {
		return 0, errors.New("NullReponse")
	}

	return float64(value.Int64), nil
}

type ConnectionEstablished struct {
	db       *sql.DB
	path     string
	name     string
	interval int
}

func NewConnectionEstablished(path string, db *sql.DB, name string, interval int) *ConnectionEstablished {
	return &ConnectionEstablished{
		db:       db,
		path:     path,
		name:     name,
		interval: interval,
	}
}

func (c *ConnectionEstablished) GetUnits() string {
	return "Count"
}

func (c *ConnectionEstablished) GetName() string {
	return c.path
}

func (c *ConnectionEstablished) GetValue() (float64, error) {
	query := fmt.Sprintf(`
		SELECT count(*) 
		FROM __lastseenhosts 
		WHERE hostkey = (select hostkey from __contexts where contextname = 'policy_server') 
		AND lastseendirection = '%s' 
		AND lastseentimestamp > NOW() - INTERVAL '%d Seconds'
		`, c.name, c.interval+300) // +300 as it looks at 5min old data upfront

	var value sql.NullInt64
	if err := c.db.QueryRow(query).Scan(&value); err != nil {
		return 0, err
	}

	if !value.Valid {
		return 0, errors.New("NullReponse")
	}

	return float64(value.Int64), nil
}
