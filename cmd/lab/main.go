package main

import (
	"fmt"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/sim"
	"os"
)

func main() {
	r, err := sim.Run(sim.Config{Runs: 1000, Seed: 42, Workers: 16, Strategy: "naive", Scenario: "baseline"})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("scenario: baseline runs=%d seed=%d\nstrategy completed lost dup-effects double-billed wasted-calls p50 p99\n", r.Runs, r.Seed)
	fmt.Printf("%-16s %6.1f%% %5d %5d %5d %5d %.2fs %.2fs\n", r.Strategy, 100*float64(r.Completed)/float64(r.Runs), r.Lost, r.DuplicateEffects, r.DoubleBilled, r.WastedCalls, float64(r.P50MS)/1000, float64(r.P99MS)/1000)
}
