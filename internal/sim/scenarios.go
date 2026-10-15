package sim

import (
	"fmt"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/fault"
	"time"
)

var Strategies = []string{"naive", "at-least-once", "exactly-once"}

type Scenario struct {
	Name, Description string
	Faults            []fault.Fault
	Drain             bool
}

func Scenarios() []Scenario {
	return []Scenario{
		{Name: "baseline", Description: "no injected faults"},
		{Name: "crash-5", Description: "worker-crash 5%", Faults: []fault.Fault{fault.Crash{Rate: .05}}},
		{Name: "dup-5", Description: "dup-delivery 5%", Faults: []fault.Fault{fault.Duplicate{Rate: .05}}},
		{Name: "hang-5", Description: "provider-hang 5%; timeout 2s", Faults: []fault.Fault{fault.Hang{Rate: .05}}},
		{Name: "error-10", Description: "provider-error 10%", Faults: []fault.Fault{fault.Error{Rate: .1}}},
		{Name: "drain", Description: "one drain at 50% of fault-free batch time", Faults: []fault.Fault{fault.Drain{Rate: 1}}, Drain: true},
		{Name: "storm", Description: "crash 3% + dup 3% + hang 3% + error 5% + drain", Faults: []fault.Fault{fault.Crash{Rate: .03}, fault.Duplicate{Rate: .03}, fault.Hang{Rate: .03}, fault.Error{Rate: .05}, fault.Drain{Rate: 1}}, Drain: true},
	}
}

func Batch(cfg Config) ([]Result, error) {
	names := Strategies
	if cfg.Strategy != "all" {
		valid := false
		for _, n := range names {
			if n == cfg.Strategy {
				valid = true
			}
		}
		if !valid {
			return nil, fmt.Errorf("unknown strategy %q", cfg.Strategy)
		}
		names = []string{cfg.Strategy}
	}
	scenarios := []Scenario{}
	for _, s := range Scenarios() {
		if cfg.Scenario == "all" || cfg.Scenario == s.Name {
			scenarios = append(scenarios, s)
		}
	}
	if len(scenarios) == 0 {
		return nil, fmt.Errorf("unknown scenario %q", cfg.Scenario)
	}
	// An identical, precomputed drain time prevents strategy-dependent injection.
	baseline := cfg
	baseline.Strategy = "naive"
	baseline.Scenario = "baseline"
	baseline.Faults = nil
	baseline.DrainAt = 0
	calibration, err := Run(baseline)
	if err != nil {
		return nil, err
	}
	out := []Result{}
	for _, s := range scenarios {
		for _, name := range names {
			cell := cfg
			cell.Strategy = name
			cell.Scenario = s.Name
			cell.Faults = s.Faults
			cell.DrainAt = 0
			if s.Drain {
				cell.DrainAt = time.Duration(calibration.BatchMS) * time.Millisecond / 2
			}
			r, err := Run(cell)
			if err != nil {
				return nil, err
			}
			out = append(out, r)
		}
	}
	return out, nil
}
