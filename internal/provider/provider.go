// Package provider simulates calls and a gateway's completed-result cache.
package provider

import (
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/workload"
	"sync"
	"time"
)

type Job struct {
	ID      int
	Key     string
	At      time.Duration
	Counted bool
	Failed  bool
}
type Provider struct {
	mu       sync.Mutex
	sequence int
	jobs     map[int]Job
	cache    map[string]Job
	used     map[int]bool
}

func New() *Provider {
	return &Provider{jobs: map[int]Job{}, cache: map[string]Job{}, used: map[int]bool{}}
}
func (p *Provider) Begin(s workload.Step, at time.Duration) Job {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sequence++
	j := Job{ID: p.sequence, Key: s.Key(), At: at, Counted: s.Kind != "bill"}
	p.jobs[j.ID] = j
	return j
}
func (p *Provider) Cache(j Job) { p.mu.Lock(); defer p.mu.Unlock(); p.cache[j.Key] = j }
func (p *Provider) Lookup(key string) (Job, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	j, ok := p.cache[key]
	return j, ok
}
func (p *Provider) Use(id int) { p.mu.Lock(); defer p.mu.Unlock(); p.used[id] = true }
func (p *Provider) Counts() (calls, wasted int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for id, j := range p.jobs {
		if j.Counted {
			calls++
			if !p.used[id] {
				wasted++
			}
		}
	}
	return
}
