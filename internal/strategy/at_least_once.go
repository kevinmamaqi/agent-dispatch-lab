package strategy

import (
	"context"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/provider"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/workload"
	"time"
)

type AtLeastOnce struct{ R *Runtime }

func (a *AtLeastOnce) Name() string { return "at-least-once" }
func (a *AtLeastOnce) Dispatch(ctx context.Context, s workload.Step) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r := a.R
	retry := func(delay time.Duration) {
		if s.Attempt >= r.Retries {
			r.Trace(s, "retry exhausted", "bounded retry policy")
			r.Ack(s)
			return
		}
		r.Requeue(s, s.Attempt+1, delay)
	}
	r.SetRetry(s, retry)
	r.Invoke(s, func(late bool) bool { return r.Apply(s, false, late) }, func(j provider.Job, err error) {
		if err != nil {
			retry(r.Backoff(s))
			return
		}
		r.Complete(s, j.ID)
	}, func() { retry(r.Timeout) }) // visibility-timeout redelivery after a crash
	return nil
}
func (a *AtLeastOnce) OnDrain(context.Context) error { a.R.Drain(false); return nil }
