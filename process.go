package orchestrator

import (
	"context"
)

// ProcessExitStatus represents the final status of a Process
type ProcessExitStatus uint8

const (
	ProcessUnknown ProcessExitStatus = iota
	ProcessUnstarted
	ProcessSuccess
	ProcessFail
)

// ProcessStatus contains various bits and pieces a process might return,
// such as logs and statuscodes and so on
type ProcessStatus struct {
	Name   string
	Logs   []string
	Status ProcessExitStatus
}

// Process is an interface which processes must implement
//
// The inteface is pretty simple: given a
type Process interface {
	Run(context.Context, Event) (ProcessStatus, error)
	ID() string
}
