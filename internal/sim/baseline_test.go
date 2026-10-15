package sim

import "testing"

func TestBaseline(t *testing.T) {
	r, err := Run(Config{Runs: 1000, Seed: 42, Workers: 16, Strategy: "naive", Scenario: "baseline"})
	if err != nil {
		t.Fatal(err)
	}
	if r.Completed != 1000 || r.Lost != 0 || r.DuplicateEffects != 0 || r.DoubleBilled != 0 {
		t.Fatalf("baseline: %+v", r)
	}
}
