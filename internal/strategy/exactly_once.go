package strategy

import (
	"context"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/provider"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/store"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/workload"
	"time"
)

// ExactlyOnce guarantees effects, not one physical execution. The store must
// fence stale versions, and external effects must honor the stable key.
type ExactlyOnce struct{ R *Runtime }

func (e *ExactlyOnce) Name() string { return "exactly-once" }
func (e *ExactlyOnce) Dispatch(ctx context.Context, s workload.Step) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r := e.R
	record := r.Log.Read(s.Key())
	if record.Phase == store.Committed {
		r.Complete(s, record.JobID)
		return nil
	}
	record, won := r.Log.Claim(s, r.Clock.Now, r.Timeout+time.Second)
	if !won {
		r.Ack(s)
		return nil
	} // another claim or a stale attempt; no call
	retry := func(delay time.Duration) {
		version, ok := r.Log.Release(s.Key(), s.Attempt)
		if !ok {
			r.Ack(s)
			return
		}
		r.Requeue(s, version, delay) // fence the old owner before handing back
	}
	r.SetRetry(s, retry)
	commit := func(job int) {
		if r.Log.Advance(s.Key(), s.Attempt, store.Committed, job) {
			r.Complete(s, job)
		} else {
			r.Ack(s)
		}
	}
	if record.Phase == store.EffectWritten {
		commit(record.JobID)
		return nil
	}
	// Persist intent BEFORE crossing the provider boundary. A resumed owner first
	// reads the result cache. An unfinished/failed call can still require retry;
	// exactly-once effects do not imply exactly-once provider computation.
	r.Log.Advance(s.Key(), s.Attempt, store.ProviderCalled, record.JobID)
	apply := func(late bool) bool {
		if !r.Apply(s, true, late) {
			return false
		}
		if late {
			return true
		} // external write dedupes even after the lease owner died
		job, _ := r.Provider.Lookup(s.Key())
		return r.Log.Advance(s.Key(), s.Attempt, store.EffectWritten, job.ID)
	}
	if cached, ok := r.Provider.Lookup(s.Key()); ok {
		if apply(false) {
			commit(cached.ID)
		} else {
			retry(r.Backoff(s))
		}
		return nil
	}
	r.Invoke(s, apply, func(j provider.Job, err error) {
		if err != nil {
			retry(r.Backoff(s))
			return
		}
		commit(j.ID)
	}, func() { retry(r.Timeout) })
	return nil
}
func (e *ExactlyOnce) OnDrain(context.Context) error {
	// Stop claims, finish within D, then release/version+1 and return unfinished
	// deliveries. The replacement pool consumes the same durable log.
	e.R.Drain(false)
	return nil
}
