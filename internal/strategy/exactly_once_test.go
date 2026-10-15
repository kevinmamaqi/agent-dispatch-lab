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
