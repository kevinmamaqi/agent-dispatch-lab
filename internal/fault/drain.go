package fault

import (
	"context"
)

type Drain struct{ Rate float64 }

func (f Drain) Apply(_ context.Context, e Event) Decision {
	return Decision{Drain: fires(e, f.Rate, "drain", "drain")}
}
