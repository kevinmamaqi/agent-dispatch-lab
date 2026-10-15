package main

import (
	"flag"
	"fmt"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/report"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/sim"
	"os"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "lab:", err)
		os.Exit(1)
	}
}
func run(args []string) error {
	reportMode := len(args) > 0 && args[0] == "report"
	if reportMode {
		args = args[1:]
	}
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
	if reportMode {
		if err := report.Write("report", rows); err != nil {
			return err
		}
		fmt.Println("Wrote report/results.{md,json,svg} and results-dark.svg; refreshed README measured blocks.")
	}
	report.Terminal(os.Stdout, rows)
	if *verbose {
		for _, r := range rows {
			fmt.Printf("\n%s / %s evidence:\n", r.Scenario, r.Strategy)
			for _, line := range r.Evidence {
				fmt.Println(line)
			}
		}
	}
	return nil
}
