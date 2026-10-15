package fault

import (
	"context"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/workload"
)

type Hang struct{ Rate float64 }

func (f Hang) Apply(_ context.Context, e Event) Decision {
	return Decision{Hang: fires(e, f.Rate, "provider-hang", "start"), Late: workload.Unit(e.Seed, e.Step.Key(), e.Step.Attempt, "hang-late") < 0.5}
}
