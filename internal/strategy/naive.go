package strategy

import (
	"context"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/provider"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/workload"
)

type Naive struct{ R *Runtime }

func (n *Naive) Name() string { return "naive" }
func (n *Naive) Dispatch(ctx context.Context, s workload.Step) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r := n.R
	r.Invoke(s, func(late bool) bool { return r.Ledger.Apply(s, false) }, func(j provider.Job, err error) {
		if err != nil {
			r.Ack(s)
			return
		}
		r.Complete(s, j.ID)
	}, func() { r.Ack(s) })
	return nil
}
func (n *Naive) OnDrain(context.Context) error { n.R.Drain(true); return nil }
