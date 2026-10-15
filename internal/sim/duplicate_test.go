package sim

import (
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/fault"
	"testing"
	"time"
)

func TestNaiveDuplicate(t *testing.T) {
	r, err := Run(Config{Runs: 1000, Seed: 42, Workers: 16, Strategy: "naive", Scenario: "duplicate", Faults: []fault.Fault{fault.Duplicate{Rate: 0.05}}, DrainAt: 0 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if !(r.DuplicateEffects > 0 && r.DoubleBilled > 0) {
		t.Fatalf("fault did not demonstrate failure: %+v", r)
	}
}
