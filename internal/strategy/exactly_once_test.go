package strategy

import (
	"context"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/clock"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/provider"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/queue"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/store"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/workload"
	"sync"
	"testing"
	"time"
)

func TestEightConcurrentDispatchesOneCommitOneCall(t *testing.T) {
	c := &clock.Clock{}
	m := store.New()
	p := provider.New()
	r := &Runtime{Clock: c, Queue: queue.New(c), Log: m, Ledger: m, Provider: p, Seed: 42, Timeout: 2 * time.Second, Active: map[int]*Work{}}
	steps := make([]workload.Step, 8)
	for i := range steps {
		steps[i] = workload.Step{RunID: "r0000", ID: "tool_0", Kind: "tool", WorkerID: i, Receipt: i + 1}
		r.Active[i+1] = &Work{Step: steps[i]}
	}
	e := &ExactlyOnce{R: r}
	var wg sync.WaitGroup
	start := make(chan struct{})
	for _, s := range steps {
		wg.Add(1)
		go func(s workload.Step) {
			defer wg.Done()
			<-start
			if err := e.Dispatch(context.Background(), s); err != nil {
				t.Error(err)
			}
		}(s)
	}
	close(start)
	wg.Wait()
	if err := c.Run(); err != nil {
		t.Fatal(err)
	}
	record := m.Read(steps[0].Key())
	calls, _ := p.Counts()
	if record.Commits != 1 || calls != 1 || m.Count(steps[0]) != 1 {
		t.Fatalf("commits=%d calls=%d effects=%d", record.Commits, calls, m.Count(steps[0]))
	}
}

func TestRecoveryAcrossEffectLogCrashWindows(t *testing.T) {
	for _, tc := range []struct {
		name   string
		effect bool
		phase  store.Phase
	}{
		{"cached-provider-result", false, store.ProviderCalled},
		{"effect-before-log", true, store.ProviderCalled},
		{"effect-recorded", true, store.EffectWritten},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := &clock.Clock{}
			m := store.New()
			p := provider.New()
			s := workload.Step{RunID: "recover", ID: "tool_0", Kind: "tool", Receipt: 1}
			m.Prepare(s)
			m.Claim(s, 0, time.Second)
			m.Advance(s.Key(), 0, store.ProviderCalled, 0)
			job := p.Begin(s, 0)
			p.Cache(job)
			if tc.effect {
				m.Apply(s, true)
			}
			if tc.phase == store.EffectWritten {
				m.Advance(s.Key(), 0, tc.phase, job.ID)
			}
			m.Release(s.Key(), 0)
			s.Attempt = 1
			s.WorkerID = 2
			r := &Runtime{Clock: c, Queue: queue.New(c), Log: m, Ledger: m, Provider: p, Seed: 42, Timeout: 2 * time.Second, Active: map[int]*Work{1: {Step: s}}, OnDone: func(_ workload.Step, id int) { p.Use(id) }}
			if err := (&ExactlyOnce{R: r}).Dispatch(context.Background(), s); err != nil {
				t.Fatal(err)
			}
			if err := c.Run(); err != nil {
				t.Fatal(err)
			}
			calls, _ := p.Counts()
			record := m.Read(s.Key())
			if calls != 1 || record.Commits != 1 || record.Phase != store.Committed || m.Count(s) != 1 {
				t.Fatalf("calls=%d record=%+v effects=%d", calls, record, m.Count(s))
			}
		})
	}
}
