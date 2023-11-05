package orchestrator

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Supported set of operations
const (
	OperationUnknown Operation = iota
	OperationCreate
	OperationRead
	OperationUpdate
	OperationDelete
)

// Operation represents one of the basic CRUD operations
// on a piece of data and can be used in Inputs to do clever
// things around ignoring certain events
type Operation uint8

// UnmarshalText implements the encoding.TextUnmarshaler interface
// allowing for a byte slice containing certain crud operations to be
// cast to Operations
func (o *Operation) UnmarshalText(b []byte) error {
	switch strings.ToLower(string(b)) {
	case "create", "insert":
		*o = OperationCreate
	case "read":
		*o = OperationRead
	case "update":
		*o = OperationUpdate
	case "delete", "remove":
		*o = OperationDelete

	default:
		return fmt.Errorf("Unknown operation %q", string(b))
	}

	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface, allowing
// for the operation type to be represented in json properly
func (o *Operation) UnmarshalJSON(b []byte) (err error) {
	var s string

	err = json.Unmarshal(b, &s)
	if err != nil {
		return
	}

	return o.UnmarshalText([]byte(s))
}

// MarshalText implements the encoding.TextMarshaler interface in order
// to get a textual representation of an Operation
func (o Operation) MarshalText() (b []byte, err error) {
	return []byte(o.String()), nil
}

// MarshalJSON implements the json.Marshaler interface which allows an
// Operation to be represented in json (which is really a json string)
func (o Operation) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.String())
}

// String returns the string representation of an Operation, or
// "unknown" for any Operation value it doesn't know about
func (o Operation) String() string {
	switch o {
	case OperationCreate:
		return "create"
	case OperationRead:
		return "read"
	case OperationUpdate:
		return "update"
	case OperationDelete:
		return "delete"
	}

	return "unknown"
}
