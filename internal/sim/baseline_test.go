package sim

import (
	"fmt"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/workload"
	"testing"
)

func TestBaseline(t *testing.T) {
	r, err := Run(Config{Runs: 1000, Seed: 42, Workers: 16, Strategy: "naive", Scenario: "baseline"})
	if err != nil {
		t.Fatal(err)
	}
	expected := 2000
	for i := 0; i < 1000; i++ {
		expected += workload.Tools(42, fmt.Sprintf("r%04d", i))
	}
	if r.ProviderCalls != expected {
		t.Fatalf("workload omitted or repeated calls: %d != %d", r.ProviderCalls, expected)
	}
	if r.Completed != 1000 || r.Lost != 0 || r.DuplicateEffects != 0 || r.DoubleBilled != 0 {
		t.Fatalf("baseline: %+v", r)
	}
}
