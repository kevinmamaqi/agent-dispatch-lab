package sim

import (
	"encoding/json"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/fault"
	"testing"
	"time"
)

func TestStrategyFaultMatrix(t *testing.T) {
	cases := []struct {
		name  string
		fs    []fault.Fault
		drain time.Duration
	}{
		{"baseline", nil, 0},
		{"crash-5", []fault.Fault{fault.Crash{Rate: .05}}, 0},
		{"dup-5", []fault.Fault{fault.Duplicate{Rate: .05}}, 0},
		{"hang-5", []fault.Fault{fault.Hang{Rate: .05}}, 0},
		{"error-10", []fault.Fault{fault.Error{Rate: .1}}, 0},
		{"drain", []fault.Fault{fault.Drain{Rate: 1}}, 85 * time.Second},
		{"storm", []fault.Fault{fault.Crash{Rate: .03}, fault.Duplicate{Rate: .03}, fault.Hang{Rate: .03}, fault.Error{Rate: .05}, fault.Drain{Rate: 1}}, 85 * time.Second},
	}
	for _, tc := range cases {
		for _, name := range []string{"naive", "at-least-once", "exactly-once"} {
			t.Run(tc.name+"/"+name, func(t *testing.T) {
				cfg := Config{Runs: 1000, Seed: 42, Workers: 16, Strategy: name, Scenario: tc.name, Faults: tc.fs, DrainAt: tc.drain}
				r, err := Run(cfg)
				if err != nil {
					t.Fatal(err)
				}
				if name == "exactly-once" && (r.Completed != 1000 || r.Lost != 0 || r.DuplicateEffects != 0 || r.DoubleBilled != 0 || r.DrainAbandoned != 0) {
					t.Fatalf("invariant failure: %+v", r)
				}
				if name == "naive" && (tc.name == "crash-5" || tc.name == "drain") && r.Lost == 0 {
					t.Fatal("naive did not lose work")
				}
				if name == "at-least-once" && (tc.name == "dup-5" || tc.name == "hang-5" || tc.name == "storm") && r.DoubleBilled == 0 {
					t.Fatal("retry did not duplicate billing")
				}
				again, err := Run(cfg)
				if err != nil {
					t.Fatal(err)
				}
				a, _ := json.Marshal(r)
				b, _ := json.Marshal(again)
				if string(a) != string(b) {
					t.Fatal("nondeterministic result")
				}
			})
		}
	}
}
