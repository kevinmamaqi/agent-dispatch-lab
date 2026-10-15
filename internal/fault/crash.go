package fault

import (
	"context"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/workload"
)

type Crash struct{ Rate float64 }

func (f Crash) Apply(_ context.Context, e Event) Decision {
	return Decision{Crash: fires(e, f.Rate, "worker-crash", "start"), AfterEffect: workload.Unit(e.Seed, e.Step.Key(), e.Step.Attempt, "crash-after") < 0.5}
}
