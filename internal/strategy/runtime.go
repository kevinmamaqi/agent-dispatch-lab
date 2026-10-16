package strategy

import (
	"context"
	"errors"
	"fmt"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/clock"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/fault"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/provider"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/queue"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/store"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/workload"
	"sort"
	"sync"
	"time"
)

type Strategy interface {
	Name() string
	Dispatch(context.Context, workload.Step) error
	OnDrain(context.Context) error
}

var ErrTimeout = errors.New("provider timeout")
var ErrProvider = errors.New("retryable provider error")

type Work struct {
	Step     workload.Step
	Finished bool
	Retry    func(time.Duration)
}
type Runtime struct {
	mu                   sync.Mutex
	OnBilled             func(workload.Step)
	Clock                *clock.Clock
	Queue                *queue.Queue
	Provider             *provider.Provider
	Log                  store.Store
	Ledger               *store.Memory
	Seed                 int64
	Faults               []fault.Fault
	Timeout, DrainBudget time.Duration
	Retries              int
	Active               map[int]*Work
	OnDone               func(workload.Step, int)
	OnFree               func(workload.Step)
	Draining             bool
	DrainAbandoned       int
	Evidence             []string
}

func (r *Runtime) Decision(s workload.Step, point string) fault.Decision {
	return fault.Decide(context.Background(), r.Faults, fault.Event{Seed: r.Seed, Step: s, Point: point, At: r.Clock.Now})
}
func (r *Runtime) Trace(s workload.Step, what, cause string) {
	r.Evidence = append(r.Evidence, fmt.Sprintf("%-16s run=%s step=%s key=%s attempt=%d (%s at t=%s)", what, s.RunID, s.ID, s.Key(), s.Attempt, cause, r.Clock.Now))
}
func (r *Runtime) Finish(s workload.Step) {
	r.mu.Lock()
	w := r.Active[s.Receipt]
	if w == nil || w.Finished {
		r.mu.Unlock()
		return
	}
	w.Finished = true
	delete(r.Active, s.Receipt)
	r.mu.Unlock()
	if r.OnFree != nil {
		r.OnFree(s)
	}
}
func (r *Runtime) Ack(s workload.Step) { r.Queue.Ack(s.Receipt); r.Finish(s) }
func (r *Runtime) Complete(s workload.Step, job int) {
	if r.OnDone != nil {
		r.OnDone(s, job)
	}
	r.Ack(s)
}
func (r *Runtime) Backoff(s workload.Step) time.Duration {
	n := s.Attempt
	if n > 4 {
		n = 4
	}
	return time.Duration(100*(1<<n)) * time.Millisecond
}
func (r *Runtime) Requeue(s workload.Step, version int, delay time.Duration) {
	r.Queue.Return(s, version, delay)
	r.Finish(s)
}
func (r *Runtime) IsActive(s workload.Step) bool {
	w := r.Active[s.Receipt]
	return w != nil && !w.Finished
}

// Invoke owns the failure timeline, not policy. effect(true) is a late external
// write after an ambiguous timeout. It cannot be cancelled by killing a worker.
func (r *Runtime) Invoke(s workload.Step, effect func(bool) bool, reply func(provider.Job, error), crash func()) {
	d := r.Decision(s, "start")
	latency := workload.Latency(r.Seed, s)
	j := r.Provider.Begin(s, r.Clock.Now)
	timeout := r.Timeout
	if d.Hang {
		latency = timeout + latency
	}
	if !d.Hang && latency > timeout {
		d.Hang = true
		d.Late = true
	}
	// Billing is an external write acknowledgement, not a model call. It gets
	// the same ambiguous-timeout faults, but is excluded from provider-call cost.
	if d.Crash && !d.AfterEffect {
		r.Clock.After(latency/2, func() {
			if r.IsActive(s) {
				r.Trace(s, "worker crash", "before effect")
				crash()
			}
		})
	}
	if d.Hang {
		r.Clock.After(timeout, func() {
			if r.IsActive(s) {
				r.Trace(s, "timeout", "provider-hang")
				reply(j, ErrTimeout)
			}
		})
		if !d.Late {
			return
		}
	}
	r.Clock.After(latency, func() {
		if d.Error {
			if r.IsActive(s) {
				r.Trace(s, "provider error", "provider-error")
				reply(j, ErrProvider)
			}
			return
		}
		r.Provider.Cache(j)
		if d.Crash && !d.AfterEffect {
			return
		}
		if d.Hang {
			effect(true)
			return
		}
		if !r.IsActive(s) {
			return
		}
		if !effect(false) {
			reply(j, ErrProvider)
			return
		}
		if d.Crash {
			r.Trace(s, "worker crash", "after effect, before commit")
			crash()
			return
		}
		reply(j, nil)
	})
}

func (r *Runtime) Drain(abandon bool) {
	r.Draining = true
	r.Clock.After(r.DrainBudget, func() {
		ids := make([]int, 0, len(r.Active))
		for id := range r.Active {
			ids = append(ids, id)
		}
		sort.Ints(ids)
		for _, id := range ids {
			w := r.Active[id]
			if w == nil {
				continue
			}
			if abandon {
				r.DrainAbandoned++
				r.Trace(w.Step, "drain abandoned", "drain deadline")
				r.Ack(w.Step)
			} else {
				r.Trace(w.Step, "drain handback", "drain deadline")
				w.Retry(0)
			}
		}
		r.Draining = false
		if r.Queue.Wake != nil {
			r.Queue.Wake()
		}
	})
}

func (r *Runtime) SetRetry(s workload.Step, f func(time.Duration)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Active[s.Receipt].Retry = f
}
func (r *Runtime) Apply(s workload.Step, unique, late bool) bool {
	before := r.Ledger.Count(s)
	if !r.Ledger.Apply(s, unique) {
		return false
	}
	after := r.Ledger.Count(s)
	if s.Kind == "bill" && before == 0 && after > 0 && r.OnBilled != nil {
		r.OnBilled(s)
	}
	if after > before && after > 1 {
		label := fmt.Sprintf("dup effect attempts=%d", after)
		if s.Kind == "bill" {
			label = fmt.Sprintf("double bill writes=%d", after)
		}
		cause := "dup-delivery/retry"
		if late {
			cause = "late completion after timeout"
		}
		r.Trace(s, label, cause)
	}
	return true
}
