package store

import (
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/workload"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestConcurrentClaimAndVersionFence(t *testing.T) {
	m := New()
	s := workload.Step{RunID: "r", ID: "tool_0", Kind: "tool"}
	var wins atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			x := s
			x.WorkerID = id
			if _, ok := m.Claim(x, 0, time.Second); ok {
				wins.Add(1)
			}
		}(i)
	}
	wg.Wait()
	if wins.Load() != 1 {
		t.Fatalf("claims won=%d", wins.Load())
	}
	v, ok := m.Release(s.Key(), 0)
	if !ok || v != 1 {
		t.Fatal("release must increment version")
	}
	if _, ok := m.Claim(s, 2*time.Second, time.Second); ok {
		t.Fatal("accepted stale version")
	}
	if m.Advance(s.Key(), 0, ProviderCalled, 1) {
		t.Fatal("stale owner advanced")
	}
	s.Attempt = 1
	if _, ok := m.Claim(s, 2*time.Second, time.Second); !ok {
		t.Fatal("cannot resume")
	}
	if m.Advance(s.Key(), 1, Committed, 1) {
		t.Fatal("skipped durable transitions")
	}
	for _, p := range []Phase{ProviderCalled, EffectWritten, Committed} {
		if !m.Advance(s.Key(), 1, p, 1) {
			t.Fatalf("phase %d rejected", p)
		}
	}
	if m.Advance(s.Key(), 1, Committed, 1) {
		t.Fatal("duplicate commit")
	}
}
func TestCommitBeforeBill(t *testing.T) {
	m := New()
	s := workload.Step{RunID: "r", ID: "plan", Kind: "plan"}
	bill := workload.Step{RunID: "r", ID: "bill", Kind: "bill", Dependencies: []string{s.Key()}}
	if m.Apply(bill, true) {
		t.Fatal("billing before prerequisite commit")
	}
	m.Claim(s, 0, time.Second)
	for _, p := range []Phase{ProviderCalled, EffectWritten, Committed} {
		m.Advance(s.Key(), 0, p, 1)
	}
	if !m.Apply(bill, true) || !m.Apply(bill, true) || m.Count(bill) != 1 {
		t.Fatal("billing uniqueness failed")
	}
}
