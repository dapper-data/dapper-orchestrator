package orchestrator

import (
	"context"
)

// ProcessExitStatus represents the final status of a Process
type ProcessExitStatus uint8

// Provided set of ExitStatuses
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

// NewProcessFunc is the suggested function that an Process should be instantiated with
// and, as such, can be used when creating a registry of Processs an orchestrator
// supports when creating Processs dynamically say from a config file, or from an API.
//
// For instance:
//
//	var processs = map[string]orchestrator.NewProcessFunc{
//	   "docker": orchestrator.NewContainerProcess,
//	   "webhook": webhooks.NewWebhookProcess,
//	}
//	func createProcess(cfg orchestrator.ProcessConfig) (orchestrator.Process, error) {
//	   return processs[cfg.Type](cfg)
//	}
type NewProcessFunc func(ProcessConfig) (Process, error)
