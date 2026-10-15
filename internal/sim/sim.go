// Package sim runs the same seeded workload through a bounded worker pool.
package sim

import (
	"context"
	"fmt"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/clock"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/fault"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/provider"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/queue"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/store"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/strategy"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/workload"
	"sort"
	"time"
)

type Config struct {
	Runs     int
	Seed     int64
	Workers  int
	Strategy string
	Scenario string
	Faults   []fault.Fault
	DrainAt  time.Duration
}
type Result struct {
	Scenario         string   `json:"scenario"`
	Strategy         string   `json:"strategy"`
	Runs             int      `json:"runs"`
	Seed             int64    `json:"seed"`
	Workers          int      `json:"workers"`
	Completed        int      `json:"runs_completed"`
	Lost             int      `json:"runs_lost"`
	DuplicateEffects int      `json:"steps_duplicated_effect"`
	DoubleBilled     int      `json:"runs_double_billed"`
	WastedCalls      int      `json:"provider_calls_wasted"`
	ProviderCalls    int      `json:"provider_calls"`
	P50MS            int64    `json:"p50_run_latency_ms"`
	P99MS            int64    `json:"p99_run_latency_ms"`
	DrainAbandoned   int      `json:"drain_abandoned"`
	BatchMS          int64    `json:"batch_latency_ms"`
	Evidence         []string `json:"-"`
}

func Run(cfg Config) (Result, error) {
	out := Result{Scenario: cfg.Scenario, Strategy: cfg.Strategy, Runs: cfg.Runs, Seed: cfg.Seed, Workers: cfg.Workers}
	if cfg.Runs < 1 || cfg.Workers < 1 {
		return out, fmt.Errorf("runs and workers must be positive")
	}
	c := &clock.Clock{}
	q := queue.New(c)
	p := provider.New()
	ledger := store.New()
	r := &strategy.Runtime{Clock: c, Queue: q, Provider: p, Log: ledger, Ledger: ledger, Seed: cfg.Seed, Faults: cfg.Faults,
		Timeout: 2 * time.Second, DrainBudget: 150 * time.Millisecond, Retries: 4, Active: map[int]*strategy.Work{}}
	var agent strategy.Strategy
	switch cfg.Strategy {
	case "naive":
		agent = &strategy.Naive{R: r}
	default:
		return out, fmt.Errorf("unknown strategy %q", cfg.Strategy)
	}
	done := map[string]bool{}
	scheduled := map[string]bool{}
	completed := map[string]time.Duration{}
	enqueue := func(s workload.Step) {
		q.Put(s, 0)
		d := r.Decision(s, "enqueue")
		if d.Duplicate {
			delay := time.Duration(0)
			if d.Delayed {
				delay = 350 * time.Millisecond
			}
			q.Put(s, delay)
			r.Trace(s, "duplicate queued", "dup-delivery")
		}
	}
	r.OnDone = func(s workload.Step, job int) {
		if done[s.Key()] {
			return
		}
		done[s.Key()] = true
		p.Use(job)
		if s.Kind == "bill" {
			completed[s.RunID] = c.Now
		}
		for _, next := range workload.Next(cfg.Seed, s, done) {
			if !scheduled[next.Key()] {
				scheduled[next.Key()] = true
				enqueue(next)
			}
		}
	}
	free := make([]int, cfg.Workers)
	for i := range free {
		free[i] = i
	}
	r.OnFree = func(s workload.Step) { free = append(free, s.WorkerID); q.Wake() }
	q.Wake = func() {
		for len(free) > 0 && !r.Draining {
			s, ok := q.Receive()
			if !ok {
				return
			}
			s.WorkerID = free[0]
			free = free[1:]
			r.Active[s.Receipt] = &strategy.Work{Step: s}
			if err := agent.Dispatch(context.Background(), s); err != nil {
				r.Trace(s, "dispatch error", err.Error())
				r.Ack(s)
			}
		}
	}
	for i := 0; i < cfg.Runs; i++ {
		s := workload.Step{RunID: fmt.Sprintf("r%04d", i), ID: "plan", Kind: "plan"}
		scheduled[s.Key()] = true
		enqueue(s)
	}
	if cfg.DrainAt > 0 {
		c.After(cfg.DrainAt, func() { _ = agent.OnDrain(context.Background()) })
	}
	if err := c.Run(); err != nil {
		return out, err
	}
	if q.Pending() != 0 || len(r.Active) != 0 {
		return out, fmt.Errorf("simulation quiesced with %d messages and %d active workers", q.Pending(), len(r.Active))
	}
	latencies := []int64{}
	for _, at := range completed {
		latencies = append(latencies, at.Milliseconds())
	}
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	if len(latencies) > 0 {
		out.P50MS = latencies[(len(latencies)-1)*50/100]
		out.P99MS = latencies[(len(latencies)-1)*99/100]
	}
	out.Completed = len(completed)
	out.Lost = cfg.Runs - out.Completed
	for _, n := range ledger.Effects {
		if n > 1 {
			out.DuplicateEffects += n - 1
		}
	}
	for _, n := range ledger.Bills {
		if n > 1 {
			out.DoubleBilled++
		}
	}
	out.ProviderCalls, out.WastedCalls = p.Counts()
	out.DrainAbandoned = r.DrainAbandoned
	out.BatchMS = c.Now.Milliseconds()
	out.Evidence = r.Evidence
	return out, nil
}
