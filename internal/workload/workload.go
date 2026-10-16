// Package workload defines the stable identities and dependency graph.
package workload

import (
	"fmt"
	"hash/fnv"
	"math"
	"time"
)

type Step struct {
	RunID        string
	ID           string
	Kind         string
	Attempt      int
	WorkerID     int
	Receipt      int
	Dependencies []string
}

func (s Step) Key() string { return s.RunID + "/" + s.ID }

// Unit is a keyed random draw: scheduling changes do not consume another
// strategy's RNG stream. Attempts share first-attempt faults across strategies.
func Unit(seed int64, key string, attempt int, purpose string) float64 {
	h := fnv.New64a()
	fmt.Fprintf(h, "%d/%s/%d/%s", seed, key, attempt, purpose)
	x := h.Sum64()
	// SplitMix64's finalizer removes correlations between adjacent identifiers.
	x = (x ^ (x >> 30)) * 0xbf58476d1ce4e5b9
	x = (x ^ (x >> 27)) * 0x94d049bb133111eb
	x ^= x >> 31
	return (float64(x>>12) + 0.5) / (1 << 52)
}

func Tools(seed int64, run string) int {
	return 3 + int(Unit(seed, run, 0, "tools")*4)
}

func Latency(seed int64, s Step) time.Duration {
	u := Unit(seed, s.Key(), s.Attempt, "latency-a")
	v := Unit(seed, s.Key(), s.Attempt, "latency-b")
	z := math.Sqrt(-2*math.Log(u)) * math.Cos(2*math.Pi*v)
	// Quantize to milliseconds for stable JSON across Go versions/platforms.
	ms := math.Round(300 * math.Exp(0.65*z))
	return time.Duration(math.Min(1800, math.Max(1, ms))) * time.Millisecond
}

func Next(seed int64, s Step, done map[string]bool) []Step {
	makeStep := func(id, kind string, deps []string) Step {
		return Step{RunID: s.RunID, ID: id, Kind: kind, Dependencies: deps}
	}
	switch s.Kind {
	case "plan":
		out := make([]Step, Tools(seed, s.RunID))
		for i := range out {
			out[i] = makeStep(fmt.Sprintf("tool_%d", i), "tool", []string{s.Key()})
		}
		return out
	case "tool":
		deps := []string{s.RunID + "/plan"}
		for i := 0; i < Tools(seed, s.RunID); i++ {
			key := fmt.Sprintf("%s/tool_%d", s.RunID, i)
			if !done[key] {
				return nil
			}
			deps = append(deps, key)
		}
		return []Step{makeStep("synthesize", "synthesize", deps)}
	case "synthesize":
		deps := []string{s.RunID + "/plan", s.Key()}
		for i := 0; i < Tools(seed, s.RunID); i++ {
			deps = append(deps, fmt.Sprintf("%s/tool_%d", s.RunID, i))
		}
		return []Step{makeStep("bill", "bill", deps)}
	}
	return nil
}
