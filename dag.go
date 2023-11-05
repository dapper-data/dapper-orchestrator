package orchestrator

import (
	"context"
	"fmt"
	"sync"

	"github.com/heimdalr/dag"
	"golang.org/x/sync/semaphore"
)

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

// DAG (or 'Directed Acyclic Graph') runs inputs and processes
type DAG struct {
	*dag.DAG
	inputs    *sync.Map
	processes *sync.Map
	wg        *semaphore.Weighted

	ErrorChan chan error
}

// New returns a DAG ready for use
func New() *DAG {
	return &DAG{
		DAG:       dag.NewDAG(),
		inputs:    new(sync.Map),
		processes: new(sync.Map),
		wg:        semaphore.NewWeighted(ConcurrentProcessors),
		ErrorChan: make(chan error),
	}
}

// AddInput takes an Input, adds it to the DAG, and runs it
// ready for events to flow through
func (d *DAG) AddInput(ctx context.Context, i Input) (err error) {
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

// AddProcess takes a Process and adds it to the DAG, so that Inputs
// can then pass Events to them
func (d DAG) AddProcess(p Process) error {
	id := p.ID()
	d.processes.Store(id, p)

	return d.AddVertexByID(id, id)
}

// AddLink accepts an Input and a Process, and links them so that when the
// input takes a
func (d DAG) AddLink(input Input, process Process) (err error) {
	return d.AddEdge(input.ID(), process.ID())
}

func (d DAG) runInput(id string, c chan Event) {
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

func (d DAG) runChild(inputID string, child string, event Event) error {
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
