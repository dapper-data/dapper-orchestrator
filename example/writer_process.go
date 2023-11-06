package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/jspc/pipelines-orchestrator"
)

type precipitation struct {
	Timestamp  time.Time `db:"timestamp"`
	LocName    string    `db:"location_name"`
	LocLat     float64   `db:"location_latitude"`
	LocLong    float64   `db:"location_longitude"`
	SensorName string    `db:"sensor"`
	ValueMM    float64   `db:"precipitation"`

	// Ignored, only here to stop `SELECT *` from borking
	ID string `db:"id"`
}

// WriterProcess is a custom process which will do some basic validations
// before writing to a destination database
type WriterProcess struct {
	name     string
	src, dst *sqlx.DB
}

// NewWriterProcess connects to both the src and dst databases and returns a WriterProcess
func NewWriterProcess(name, src, dst string) (wp WriterProcess, err error) {
	wp.name = name

	wp.src, err = sqlx.Connect("postgres", src)
	if err != nil {
		return
	}

	wp.dst, err = sqlx.Connect("postgres", dst)

	return
}

// ID returns the ID for this process
func (wp WriterProcess) ID() string {
	return wp.name
}

// Run the process
func (wp WriterProcess) Run(ctx context.Context, e orchestrator.Event) (ps orchestrator.ProcessStatus, err error) {
	ps.Logs = make([]string, 0)
	ps.Name = wp.name
	ps.Status = orchestrator.ProcessUnknown

	defer func() {
		if err != nil {
			ps.Status = orchestrator.ProcessFail
		} else {
			ps.Status = orchestrator.ProcessSuccess
		}
	}()

	in := new(precipitation)
	err = wp.src.Get(in, fmt.Sprintf("SELECT * FROM %s WHERE id=$1", e.Location), e.ID)
	if err != nil {
		return
	}

	switch e.Trigger {
	case "raw_writes":
		ps.Logs, err = wp.rawToCleansed(ctx, e, in)

	case "cleansed_writes":
		ps.Logs, err = wp.cleansedToReporting(ctx, e, in)
	}

	return
}

func (wp WriterProcess) rawToCleansed(ctx context.Context, e orchestrator.Event, in *precipitation) (logs []string, err error) {
	logs = make([]string, 0)

	// Validate precipitation levels, and that lat/long are within valid ranges
	if in.ValueMM < 0 {
		logs = append(logs, "precipitation cannot be negative")
	}

	if in.LocLat < -90 || in.LocLat > 90 {
		logs = append(logs, "latitude must be between -90 and 90")
	}

	if in.LocLong < -180 || in.LocLong > 180 {
		logs = append(logs, "latitude must be between -90 and 90")
	}

	if len(logs) > 0 {
		err = fmt.Errorf("Row failed validation")

		return
	}

	tx := wp.dst.MustBegin()
	tx.NamedExec("INSERT INTO precipitation (timestamp, location_name, location_latitude, location_longitude, sensor, precipitation) VALUES (:timestamp, :location_name, :location_latitude, :location_longitude, :sensor, :precipitation)", in)

	return []string{"valid data <3"}, tx.Commit()
}

func (wp WriterProcess) cleansedToReporting(ctx context.Context, e orchestrator.Event, in *precipitation) (logs []string, err error) {
	tx := wp.dst.MustBegin()
	tx.NamedExec("INSERT INTO sensor (sensor, location, latitude, longitude) VALUES (:sensor, :location_name, :location_latitude, :location_longitude)", in)
	tx.NamedExec("INSERT INTO precipitation (sensor, timestamp, value) VALUES (:sensor, :timestamp, :precipitation)", in)

	return []string{"created reporting data"}, tx.Commit()
}
