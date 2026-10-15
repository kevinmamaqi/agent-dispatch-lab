package sim

import (
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/fault"
	"testing"
	"time"
)

func TestNaiveDrain(t *testing.T) {
	r, err := Run(Config{Runs: 1000, Seed: 42, Workers: 16, Strategy: "naive", Scenario: "drain", Faults: []fault.Fault{fault.Drain{Rate: 1}}, DrainAt: 50 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if !(r.Lost > 0 && r.DrainAbandoned > 0) {
		t.Fatalf("fault did not demonstrate failure: %+v", r)
	}
}
