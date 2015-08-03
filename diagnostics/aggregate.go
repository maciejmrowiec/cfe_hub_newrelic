package diagnostics

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type LocalAverageDiagnostics struct {
	sample_name       string
	query_type        string
	table_diagnostics *Diagnostics
	interval          int
	subset_path       string
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
