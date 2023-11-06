package orchestrator

import (
	"context"
	"fmt"
	"sync"

	"github.com/heimdalr/dag"
	"golang.org/x/sync/semaphore"
)

// ConcurrentProcessors limits the number of processes which can be kicked off
// at once
var ConcurrentProcessors int64 = 8

// ProcessInterfaceConversionError returns when trying to load a process from our
// internal process store returns completely unexpected data
//
// This error represents a huge failure somewhere and should cause a stop-the-world
// event
type ProcessInterfaceConversionError struct {
	input, process string
	iface          any
}

// Error returns a descriptive error message
func (e ProcessInterfaceConversionError) Error() string {
	return fmt.Sprintf("unable to run %q -> %q part of process, %[2]q cannot be converted from interface{} to Process (instead it looks like a %T)", e.input, e.process, e.iface)
}

// NewTestProcessInterfaceConversionError can be used to return a testable error (in tests)
func NewTestProcessInterfaceConversionError(input, process string, iface any) ProcessInterfaceConversionError {
	return ProcessInterfaceConversionError{
		input:   input,
		process: process,
		iface:   iface,
	}
}

// UnknownProcessError returns when an input tries to trigger a process whch doesn't
// exist
type UnknownProcessError struct {
	input, process string
}

// Error returns a descriptive error message
func (e UnknownProcessError) Error() string {
	return fmt.Sprintf("unable to run %q -> %q part of process, process %[2]q is unknown", e.input, e.process)
}

// NewTestUnknownProcessError can be used to return a testable error (in tests)
func NewTestUnknownProcessError(input, process string) UnknownProcessError {
	return UnknownProcessError{
		input:   input,
		process: process,
	}
}

// Orchestrator is the workhorse of this package. It:
//
//  1. Supervises inputs
//  2. Manages the lifecycle of processes, which run on events
//  3. Syncs events from inputs across multiple processes in a DAG
//
// Multiple Orchestrators _can_ be run, like in a cluster, but out of the box
// doesn't contain any logic to synchronise inputs and/or processes that wont cluster
// natively (such as the postgres sample input)
type Orchestrator struct {
	*dag.DAG
	inputs    *sync.Map
	processes *sync.Map
	wg        *semaphore.Weighted

	ErrorChan chan error
}

// New returns an Orchestrator ready for use
func New() *Orchestrator {
	return &Orchestrator{
		DAG:       dag.NewDAG(),
		inputs:    new(sync.Map),
		processes: new(sync.Map),
		wg:        semaphore.NewWeighted(ConcurrentProcessors),
		ErrorChan: make(chan error),
	}
}

// AddInput takes an Input, adds it to the Orchestrator's DAG, and runs it
// ready for events to flow through
//
// AddInput will error when duplicate input IDs are specified. Any other error
// from the running of an Input comes via the Orchestrator's ErrorChan - this is
// because Inputs are run in separate goroutines
func (d *Orchestrator) AddInput(ctx context.Context, i Input) (err error) {
	id := i.ID()
	d.inputs.Store(id, i)

	err = d.AddVertexByID(id, id)
	if err != nil {
		return
	}

	c := make(chan Event)
	go func() {
		panic(i.Process(ctx, c))
	}()

	go d.runInput(id, c)

	return
}

// AddProcess adds a Process to the Orchestrator's DAG, ready to be triggered
// by Inputs.
//
// Processes are not run until an Event is generated by an Input, and that Input is
// linked to the specified Process.
//
// This means long running processes with state should either be re-architected to use
// some kind of persistence level, or should be a separate service which exposes (say)
// a webhook or similar trigger
func (d Orchestrator) AddProcess(p Process) error {
	id := p.ID()
	d.processes.Store(id, p)

	return d.AddVertexByID(id, id)
}

// AddLink accepts an Input and a Process, and links them so that when the
// input triggers an event, the specified process is called
func (d Orchestrator) AddLink(input Input, process Process) (err error) {
	return d.AddEdge(input.ID(), process.ID())
}

func (d Orchestrator) runInput(id string, c chan Event) {
	for event := range c {
		children, err := d.GetChildren(id)
		if err != nil {
			continue
		}

		for k := range children {
			go func() {
				err = d.runChild(id, k, event)
				if err != nil {
					d.ErrorChan <- err
				}
			}()
		}
	}
}

func (d Orchestrator) runChild(inputID string, child string, event Event) error {
	d.wg.Acquire(context.Background(), 1)
	defer d.wg.Release(1)

	process, ok := d.processes.Load(child)
	if !ok {
		return UnknownProcessError{
			input:   inputID,
			process: child,
		}
	}

	pp, ok := process.(Process)
	if !ok {
		return ProcessInterfaceConversionError{
			input:   inputID,
			process: child,
			iface:   process,
		}
	}

	status, err := pp.Run(context.Background(), event)
	if err != nil {
		return err
	}

	for _, l := range status.Logs {
		fmt.Printf("%s -> %s\n", status.Name, l)
	}

	return nil
}
