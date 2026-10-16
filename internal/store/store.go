// Package store models durable state and atomic idempotency constraints.
package store

import (
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/workload"
	"sync"
	"time"
)

type Phase int

const (
	Pending Phase = iota
	Claimed
	ProviderCalled
	EffectWritten
	Committed
)

type Record struct {
	Step       workload.Step
	Phase      Phase
	Version    int
	Owner      int
	LeaseUntil time.Duration
	JobID      int
	Commits    int
}

// Store deliberately has four operations. A persistent adapter must preserve
// version checks, uniqueness, and each transition's transaction boundary.
type Store interface {
	Read(string) Record
	Claim(workload.Step, time.Duration, time.Duration) (Record, bool)
	Advance(string, int, Phase, int) bool
	Release(string, int) (int, bool)
}

type Memory struct {
	mu      sync.Mutex
	records map[string]Record
	Effects map[string]int
	Bills   map[string]int
}

func New() *Memory {
	return &Memory{records: map[string]Record{}, Effects: map[string]int{}, Bills: map[string]int{}}
}
func (m *Memory) Read(key string) Record { m.mu.Lock(); defer m.mu.Unlock(); return m.records[key] }
func (m *Memory) Claim(s workload.Step, now, lease time.Duration) (Record, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, exists := m.records[s.Key()]
	if !exists {
		r = Record{Step: s, Version: 0, Owner: -1}
	}
	if r.Phase == Committed || s.Attempt != r.Version {
		return r, false
	}
	// Even the same worker cannot claim twice. Expiry needs a fenced Release.
	if r.Owner >= 0 {
		return r, false
	}
	r.Owner = s.WorkerID
	r.LeaseUntil = now + lease
	if r.Phase == Pending {
		r.Phase = Claimed
	}
	m.records[s.Key()] = r
	return r, true
}
func (m *Memory) Advance(key string, version int, p Phase, job int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.records[key]
	if !ok || r.Version != version || r.Owner < 0 || p < r.Phase || p > r.Phase+1 {
		return false
	}
	if p == Committed && r.Phase != EffectWritten {
		return false
	}
	if p == Committed {
		r.Commits++
		r.Owner = -1
	}
	r.Phase = p
	if job != 0 {
		r.JobID = job
	}
	m.records[key] = r
	return true
}
func (m *Memory) Release(key string, version int) (int, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.records[key]
	if !ok || r.Version != version || r.Phase == Committed {
		return r.Version, false
	}
	r.Version++
	r.Owner = -1
	m.records[key] = r
	return r.Version, true
}

// Apply represents INSERT ... ON CONFLICT DO NOTHING. Late external writes
// use the same key. Billing also checks committed predecessors under this lock.
func (m *Memory) Apply(s workload.Step, unique bool) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	ledger := m.Effects
	key := s.Key()
	if s.Kind == "bill" {
		if unique {
			for _, dep := range s.Dependencies {
				if m.records[dep].Phase != Committed {
					return false
				}
			}
		}
		ledger = m.Bills
		key = s.RunID
	} else if s.Kind != "tool" {
		return true
	}
	if unique && ledger[key] > 0 {
		return true
	}
	ledger[key]++
	return true
}

func (m *Memory) Count(s workload.Step) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s.Kind == "bill" {
		return m.Bills[s.RunID]
	}
	return m.Effects[s.Key()]
}

// Prepare persists pending before the scheduler publishes a new logical step.
func (m *Memory) Prepare(s workload.Step) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.records[s.Key()]; !ok {
		m.records[s.Key()] = Record{Step: s, Owner: -1}
	}
}
