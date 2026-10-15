// Package clock is a single-threaded discrete-event clock, never a wall clock.
package clock

import (
	"container/heap"
	"time"
)

type event struct {
	at  time.Duration
	seq int
	run func()
}
type events []event

func (e events) Len() int { return len(e) }
func (e events) Less(i, j int) bool {
	if e[i].at == e[j].at {
		return e[i].seq < e[j].seq
	}
	return e[i].at < e[j].at
}
func (e events) Swap(i, j int) { e[i], e[j] = e[j], e[i] }
func (e *events) Push(x any)   { *e = append(*e, x.(event)) }
func (e *events) Pop() any     { a := *e; x := a[len(a)-1]; *e = a[:len(a)-1]; return x }

type Clock struct {
	Now     time.Duration
	seq     int
	pending events
}

func (c *Clock) After(d time.Duration, f func()) {
	if d < 0 {
		panic("negative virtual delay")
	}
	c.seq++
	heap.Push(&c.pending, event{c.Now + d, c.seq, f})
}
func (c *Clock) Run() error {
	for n := 0; len(c.pending) > 0; n++ {
		if n >= 5000000 {
			return fmtError()
		}
		e := heap.Pop(&c.pending).(event)
		c.Now = e.at
		e.run()
	}
	return nil
}
