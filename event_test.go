package orchestrator_test

import (
	"testing"

	"github.com/dapper-data/dapper-orchestrator"
)

func TestEvent_JSON(t *testing.T) {
	expect := `{"location":"the ether","operation":"delete","id":"an-id","trigger":"tests"}`
	received, err := orchestrator.Event{
		Location:  "the ether",
		Operation: orchestrator.OperationDelete,
		ID:        "an-id",
		Trigger:   "tests",
	}.JSON()

	if err != nil {
		t.Fatal(err)
	}

	if expect != received {
		t.Errorf("expected:\n%s\nreceived:\n%s", expect, received)
	}
}
