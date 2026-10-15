// Package queue models receipts, visibility and explicit redelivery.
package queue

import (
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/clock"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/workload"
	"time"
)

type Queue struct {
	Clock    *clock.Clock
	ready    []workload.Step
	receipts map[int]workload.Step
	sequence int
	Wake     func()
}

func New(c *clock.Clock) *Queue { return &Queue{Clock: c, receipts: map[int]workload.Step{}} }
func (q *Queue) Put(s workload.Step, delay time.Duration) {
	q.Clock.After(delay, func() {
		q.ready = append(q.ready, s)
		if q.Wake != nil {
			q.Wake()
		}
	})
}
func (q *Queue) Receive() (workload.Step, bool) {
	if len(q.ready) == 0 {
		return workload.Step{}, false
	}
	s := q.ready[0]
	q.ready = q.ready[1:]
	q.sequence++
	s.Receipt = q.sequence
	q.receipts[s.Receipt] = s
	return s, true
}
func (q *Queue) Ack(receipt int) { delete(q.receipts, receipt) }
func (q *Queue) Return(s workload.Step, version int, delay time.Duration) {
	if _, ok := q.receipts[s.Receipt]; !ok {
		return
	}
	q.Ack(s.Receipt)
	s.Attempt = version
	q.Put(s, delay)
}
func (q *Queue) Pending() int { return len(q.ready) + len(q.receipts) }
