package orchestrator_test

import (
	"encoding/json"
	"testing"

	"github.com/dapper-data/dapper-orchestrator"
)

func TestOperation_UnmarshalText(t *testing.T) {
	for _, test := range []struct {
		input       string
		expect      orchestrator.Operation
		expectError bool
	}{
		{"create", orchestrator.OperationCreate, false},
		{"CREATE", orchestrator.OperationCreate, false},
		{"cReAtE", orchestrator.OperationCreate, false},
		{"insert", orchestrator.OperationCreate, false},
		{"read", orchestrator.OperationRead, false},
		{"update", orchestrator.OperationUpdate, false},
		{"delete", orchestrator.OperationDelete, false},
		{"remove", orchestrator.OperationDelete, false},

		// Error cases
		{"new", orchestrator.OperationUnknown, true},
	} {
		t.Run(test.input, func(t *testing.T) {
			o := new(orchestrator.Operation)

			err := o.UnmarshalText([]byte(test.input))
			if err == nil && test.expectError {
				t.Errorf("expected error, received none")
			} else if err != nil && !test.expectError {
				t.Errorf("unexpected error %#v", err)
			}

			if test.expect != *o {
				t.Errorf("expected %#v, received %#v", test.expect, *o)
			}
		})
	}
}

func TestOperation_UnmarshalJSON(t *testing.T) {
	for _, test := range []struct {
		input       string
		expect      orchestrator.Operation
		expectError bool
	}{
		{"create", orchestrator.OperationCreate, false},
		{"CREATE", orchestrator.OperationCreate, false},
		{"cReAtE", orchestrator.OperationCreate, false},
		{"insert", orchestrator.OperationCreate, false},
		{"read", orchestrator.OperationRead, false},
		{"update", orchestrator.OperationUpdate, false},
		{"delete", orchestrator.OperationDelete, false},
		{"remove", orchestrator.OperationDelete, false},

		// Error cases
		{"new", orchestrator.OperationUnknown, true},
	} {
		t.Run(test.input, func(t *testing.T) {
			input, err := json.Marshal(test.input)
			if err != nil {
				t.Fatal(err)
			}

			o := new(orchestrator.Operation)

			err = o.UnmarshalJSON(input)
			if err == nil && test.expectError {
				t.Errorf("expected error, received none")
			} else if err != nil && !test.expectError {
				t.Errorf("unexpected error %#v", err)
			}

			if test.expect != *o {
				t.Errorf("expected %#v, received %#v", test.expect, *o)
			}
		})
	}
}

func TestOperation_MarshalText(t *testing.T) {
	for _, test := range []struct {
		o           orchestrator.Operation
		expect      string
		expectError bool
	}{
		{orchestrator.OperationCreate, "create", false},
		{orchestrator.OperationRead, "read", false},
		{orchestrator.OperationUpdate, "update", false},
		{orchestrator.OperationDelete, "delete", false},
		{orchestrator.OperationUnknown, "unknown", false},

		// Cover off extra operations that may be added later; these tests
		// ought to fail and, thus, remind us to write tests
		{orchestrator.Operation(5), "unknown", false},
		{orchestrator.Operation(6), "unknown", false},
		{orchestrator.Operation(7), "unknown", false},
		{orchestrator.Operation(8), "unknown", false},
		{orchestrator.Operation(9), "unknown", false},
		{orchestrator.Operation(10), "unknown", false},
		{orchestrator.Operation(11), "unknown", false},
		{orchestrator.Operation(12), "unknown", false},
	} {
		t.Run(test.expect, func(t *testing.T) {
			received, err := test.o.MarshalText()
			if err == nil && test.expectError {
				t.Errorf("expected error, received none")
			} else if err != nil && !test.expectError {
				t.Errorf("unexpected error %#v", err)
			}

			if test.expect != string(received) {
				t.Errorf("expected %q, received %q", test.expect, string(received))
			}

		})
	}
}
