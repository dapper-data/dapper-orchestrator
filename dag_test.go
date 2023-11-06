package orchestrator_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/dapper-data/dapper-orchestrator"
)

type dummyInput struct{}

func (dummyInput) Process(_ context.Context, c chan orchestrator.Event) error {
	for {
		time.Sleep(time.Millisecond * 100)
		c <- orchestrator.Event{
			Location:  "dag_test.go",
			Operation: orchestrator.OperationCreate,
			ID:        "1",
			Trigger:   "dummy-input",
		}
	}
}

func (dummyInput) ID() string {
	return "dummy-input"
}

type dummyProcess struct {
	ev orchestrator.Event
}

func (d *dummyProcess) Run(_ context.Context, ev orchestrator.Event) (orchestrator.ProcessStatus, error) {
	d.ev = ev

	return orchestrator.ProcessStatus{
		Name:   "dummy-process",
		Logs:   []string{ev.Operation.String(), "Hello, tests!"},
		Status: orchestrator.ProcessSuccess,
	}, nil
}

func (dummyProcess) ID() string {
	return "dummy-process"
}

func TestNew(t *testing.T) {
	defer func() {
		err := recover()
		if err != nil {
			t.Fatal(err)
		}
	}()

	orchestrator.New()
}

// TestOrchestrator is a really bad test- it has sleeps and all kinds of nonsense
// in it. It may be flakey, too.
func TestOrchestrator(t *testing.T) {
	d := orchestrator.New()

	di := dummyInput{}
	err := d.AddInput(context.Background(), di)
	if err != nil {
		t.Fatal(err)
	}

	dp := new(dummyProcess)
	err = d.AddProcess(dp)
	if err != nil {
		t.Fatal(err)
	}

	err = d.AddLink(di, dp)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Millisecond * 150)

	expect := orchestrator.Event{
		Location:  "dag_test.go",
		Operation: orchestrator.OperationCreate,
		ID:        "1",
		Trigger:   "dummy-input",
	}
	if !reflect.DeepEqual(expect, dp.ev) {
		t.Errorf("expected\n%#v\nreceived\n%#v", expect, dp.ev)
	}
}

func TestProcessInterfaceConversionError_Error(t *testing.T) {
	expect := `unable to run "dummy-input" -> "dummy-process" part of process, "dummy-process" cannot be converted from interface{} to Process (instead it looks like a orchestrator.Event)`
	err := orchestrator.NewTestProcessInterfaceConversionError("dummy-input", "dummy-process", orchestrator.Event{})

	if expect != err.Error() {
		t.Errorf("expected\n%s\nreceived\n%s", expect, err.Error())
	}
}

func TestUnknownProcessError_Error(t *testing.T) {
	expect := `unable to run "dummy-input" -> "dummy-process" part of process, process "dummy-process" is unknown`
	err := orchestrator.NewTestUnknownProcessError("dummy-input", "dummy-process")

	if expect != err.Error() {
		t.Errorf("expected\n%s\nreceived\n%s", expect, err.Error())
	}

}
