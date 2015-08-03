package diagnostics

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
)

const DELTA = "delta"
const REBASE = "rebase"

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

	if !value.Valid {
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

	if !value.Valid {
		return 0, errors.New("NullReponse")
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

	if !value.Valid {
		return 0, errors.New("NullReponse")
	}

	return value.Float64, nil
}
