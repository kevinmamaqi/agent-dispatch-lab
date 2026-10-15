package fault

import (
	"context"
)

type Error struct{ Rate float64 }

func (f Error) Apply(_ context.Context, e Event) Decision {
	return Decision{Error: fires(e, f.Rate, "provider-error", "start")}
}
