package fault

import (
	"context"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/workload"
)

type Duplicate struct{ Rate float64 }

func (f Duplicate) Apply(_ context.Context, e Event) Decision {
	return Decision{Duplicate: fires(e, f.Rate, "dup-delivery", "enqueue"), Delayed: workload.Unit(e.Seed, e.Step.Key(), e.Step.Attempt, "dup-delay") < 0.5}
}
