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

// NewInputFunc is the suggested function that an Input should be instantiated with
// and, as such, can be used when creating a registry of Inputs an orchestrator
// supports when creating Inputs dynamically say from a config file, or from an API.
//
// For instance:
//
//	var inputs = map[string]orchestrator.NewInputFunc{
//	   "postgres": orchestrator.NewPostgresInput,
//	   "webhook": webhooks.NewWebhookInput,
//	}
//	func createInput(cfg orchestrator.InputConfig) (orchestrator.Input, error) {
//	   return inputs[cfg.Type](cfg)
//	}
type NewInputFunc func(InputConfig) (Input, error)
