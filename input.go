package orchestrator

import (
	"context"
)

// Input is a simple interface, and exposes a long running process called Process
// which is expected to stream Events.
//
// It is the job of the Orchestrator to understand which channel is assigned to
// which input and to route messages accordingly
type Input interface {
	Process(context.Context, chan Event) error
	ID() string
}
