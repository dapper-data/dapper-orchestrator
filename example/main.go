package main

import (
	"context"
	"fmt"

	"github.com/dapper-data/dapper-orchestrator"
)

var (
	raw       = "postgresql://postgres:postgres@localhost:5432/raw?sslmode=disable"
	cleansed  = "postgresql://postgres:postgres@localhost:5432/cleansed?sslmode=disable"
	reporting = "postgresql://postgres:postgres@localhost:5432/reporting?sslmode=disable"
)

func main() {
	orchestrator.ConcurrentProcessors = 4

	rawInput, err := orchestrator.NewPostgresInput(orchestrator.InputConfig{
		Name:             "raw_writes",
		ConnectionString: raw,
	})
	if err != nil {
		panic(err)
	}

	cleansedInput, err := orchestrator.NewPostgresInput(orchestrator.InputConfig{
		Name:             "cleansed_writes",
		ConnectionString: cleansed,
	})
	if err != nil {
		panic(err)
	}

	rawToCleansed, err := NewWriterProcess("raw_to_cleansed", raw, cleansed)
	if err != nil {
		panic(err)
	}

	cleansedToReporting, err := NewWriterProcess("cleansed_to_reporting", cleansed, reporting)
	if err != nil {
		panic(err)
	}

	d := orchestrator.New()

	err = d.AddInput(context.Background(), rawInput)
	if err != nil {
		panic(err)
	}

	err = d.AddInput(context.Background(), cleansedInput)
	if err != nil {
		panic(err)
	}

	err = d.AddProcess(rawToCleansed)
	if err != nil {
		panic(err)
	}

	err = d.AddProcess(cleansedToReporting)
	if err != nil {
		panic(err)
	}

	err = d.AddLink(rawInput, rawToCleansed)
	if err != nil {
		panic(err)
	}

	err = d.AddLink(cleansedInput, cleansedToReporting)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case err = <-d.ErrorChan:
			fmt.Println(err)
		}
	}
}
