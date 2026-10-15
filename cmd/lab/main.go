package main

import (
	"flag"
	"fmt"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/sim"
	"os"
	"text/tabwriter"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "lab:", err)
		os.Exit(1)
	}
}
func run(args []string) error {
	fs := flag.NewFlagSet("lab", flag.ContinueOnError)
	runs := fs.Int("runs", 1000, "agent runs per cell")
	seed := fs.Int64("seed", 42, "deterministic seed")
	workers := fs.Int("workers", 16, "concurrent virtual workers")
	scenario := fs.String("scenario", "all", "baseline, crash-5, dup-5, hang-5, error-10, drain, storm, or all")
	name := fs.String("strategy", "all", "naive, at-least-once, exactly-once, or all")
	verbose := fs.Bool("v", false, "print fault and invariant-violation evidence")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %v", fs.Args())
	}
	rows, err := sim.Batch(sim.Config{Runs: *runs, Seed: *seed, Workers: *workers, Scenario: *scenario, Strategy: *name})
	if err != nil {
		return err
	}
	previous := ""
	for _, r := range rows {
		if r.Scenario != previous {
			description := ""
			for _, s := range sim.Scenarios() {
				if s.Name == r.Scenario {
					description = s.Description
				}
			}
			fmt.Printf("\nscenario: %s (%s) runs=%d seed=%d workers=%d\n", r.Scenario, description, r.Runs, r.Seed, r.Workers)
			previous = r.Scenario
		}
		printRow(r)
		if *verbose {
			for _, line := range r.Evidence {
				fmt.Println(line)
			}
		}
	}
	return nil
}
func printRow(r sim.Result) {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "strategy\tcompleted\tlost\tdup-effects\tdouble-billed\twasted-calls\tabandoned\tp50\tp99")
	fmt.Fprintf(w, "%s\t%.1f%%\t%d\t%d\t%d\t%d\t%d\t%.2fs\t%.2fs\n", r.Strategy, 100*float64(r.Completed)/float64(r.Runs), r.Lost, r.DuplicateEffects, r.DoubleBilled, r.WastedCalls, r.DrainAbandoned, float64(r.P50MS)/1000, float64(r.P99MS)/1000)
	w.Flush()
}
