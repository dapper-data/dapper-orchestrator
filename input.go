package orchestrator

import (
	"context"
)

// Input is a simple interface, and exposes a long running process called Handle
// which is expected to stream Events.
//
// It is the job of the Orchestrator to understand which channel is assigned to
// which input and to route messages accordingly
type Input interface {
	// Handle inputs from this input source, creating Events and
	// streaming down the Event channel
	Handle(context.Context, chan Event) error
	ID() string
}
