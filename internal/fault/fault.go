// Package fault holds keyed, composable fault decisions.
package fault

import (
	"context"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/workload"
	"time"
)

type Event struct {
	Seed  int64
	Step  workload.Step
	Point string
	At    time.Duration
}
type Decision struct{ Crash, AfterEffect, Duplicate, Delayed, Hang, Late, Error, Drain bool }
type Fault interface {
	Apply(context.Context, Event) Decision
}

func fires(e Event, rate float64, name, point string) bool {
	return e.Point == point && workload.Unit(e.Seed, e.Step.Key(), e.Step.Attempt, name) < rate
}
func Decide(ctx context.Context, fs []Fault, e Event) Decision {
	var d Decision
	for _, f := range fs {
		x := f.Apply(ctx, e)
		d.Crash = d.Crash || x.Crash
		d.AfterEffect = d.AfterEffect || x.AfterEffect
		d.Duplicate = d.Duplicate || x.Duplicate
		d.Delayed = d.Delayed || x.Delayed
		d.Hang = d.Hang || x.Hang
		d.Late = d.Late || x.Late
		d.Error = d.Error || x.Error
		d.Drain = d.Drain || x.Drain
	}
	return d
}
