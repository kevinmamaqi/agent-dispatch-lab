package sim

import "testing"

func TestDefaultBatch(t *testing.T) {
	rows, err := Batch(Config{Runs: 1000, Seed: 42, Workers: 16, Scenario: "all", Strategy: "all"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 21 {
		t.Fatalf("cells=%d", len(rows))
	}
	for _, r := range rows {
		if r.Strategy == "exactly-once" && (r.Completed != 1000 || r.Lost+r.DuplicateEffects+r.DoubleBilled+r.DrainAbandoned != 0) {
			t.Fatalf("%s: %+v", r.Scenario, r)
		}
	}
	for _, tc := range []Config{{Runs: 0, Workers: 1, Scenario: "all", Strategy: "all"}, {Runs: 1, Workers: 0, Scenario: "all", Strategy: "all"}, {Runs: 1, Workers: 1, Scenario: "missing", Strategy: "all"}, {Runs: 1, Workers: 1, Scenario: "all", Strategy: "missing"}} {
		if _, err := Batch(tc); err == nil {
			t.Fatalf("accepted invalid config: %+v", tc)
		}
	}
}
