package sim

import (
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/fault"
	"testing"
)

func TestRetryDoesNotDeduplicateBilling(t *testing.T) {
	for _, tc := range []struct {
		name string
		fs   []fault.Fault
	}{
		{"dup-5", []fault.Fault{fault.Duplicate{Rate: .05}}},
		{"hang-5", []fault.Fault{fault.Hang{Rate: .05}}},
		{"storm", []fault.Fault{fault.Crash{Rate: .03}, fault.Duplicate{Rate: .03}, fault.Hang{Rate: .03}, fault.Error{Rate: .05}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := Run(Config{Runs: 1000, Seed: 42, Workers: 16, Strategy: "at-least-once", Scenario: tc.name, Faults: tc.fs})
			if err != nil {
				t.Fatal(err)
			}
			if r.DoubleBilled == 0 {
				t.Fatalf("retry demo became trivial: %+v", r)
			}
		})
	}
}
