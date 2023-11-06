package orchestrator_test

import (
	"testing"

	"github.com/dapper-data/dapper-orchestrator"
)

func TestInputConfig_ID(t *testing.T) {
	for _, test := range []struct {
		ic     orchestrator.InputConfig
		expect string
	}{
		{orchestrator.InputConfig{}, ""},
		{orchestrator.InputConfig{Name: "test"}, "test"},
	} {
		t.Run(test.expect, func(t *testing.T) {
			received := test.ic.ID()
			if test.expect != received {
				t.Errorf("expected %q, received %q", test.expect, received)
			}
		})
	}
}

func TestProcessConfig_ID(t *testing.T) {
	for _, test := range []struct {
		ic     orchestrator.ProcessConfig
		expect string
	}{
		{orchestrator.ProcessConfig{}, ""},
		{orchestrator.ProcessConfig{Name: "test"}, "test"},
	} {
		t.Run(test.expect, func(t *testing.T) {
			received := test.ic.ID()
			if test.expect != received {
				t.Errorf("expected %q, received %q", test.expect, received)
			}
		})
	}
}
