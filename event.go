package orchestrator

import (
	"encoding/json"
)

// Event represents basic metadata that each Input provides
type Event struct {
	// Location could be a table name, a topic, an event from some store,
	// or anything really- it is up to both the Input and the Process to
	// agree on what this means
	Location  string    `json:"location"`
	Operation Operation `json:"operation"`
	ID        string    `json:"id"`

	// Trigger is the name or ID of the Input which triggers this
	// process, which can be useful for routing/ flow control in
	// triggers
	Trigger string `json:"trigger"`
}

// JSON returns the json representation for an event, in a way that our
// processes can later work with
func (e Event) JSON() (string, error) {
	b, err := json.Marshal(e)

	return string(b), err
}
