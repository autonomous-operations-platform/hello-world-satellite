package operators

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"

	"github.tools.sap/autonomous-operations/aos-framework/pkg/framework/registry"
	"github.tools.sap/autonomous-operations/aos-framework/pkg/framework/store"
)

func init() { registry.RegisterOperator(&FanSpeedController{}) }

// FanSpeedController reacts to temperature insights and sets the fan speed:
// 80% on temperature-high, 20% on temperature-low.
type FanSpeedController struct{}

func (o *FanSpeedController) Name() string { return "fan-speed-controller" }

func (o *FanSpeedController) Filter() store.InsightFilter {
	return store.InsightFilter{Patterns: []string{"temperature-high", "temperature-low"}}
}

func (o *FanSpeedController) Execute(ctx context.Context, logger *slog.Logger, insight store.Insight, s store.OperatorStore) error {
	var targetSpeed int
	switch insight.Pattern {
	case "temperature-high":
		targetSpeed = 80
	case "temperature-low":
		targetSpeed = 20
	}

	s.AddActionLog(ctx, store.NewActionLog{
		StepID:      "set-fan-speed",
		Description: fmt.Sprintf("Setting fan to %d%%: %s", targetSpeed, insight.Summary),
		State:       store.ActionLogStateAttempting,
	})

	// Mock: call fan speed API here.

	if rand.Float32() < 0.3 {
		s.AddActionLog(ctx, store.NewActionLog{
			StepID:      "set-fan-speed",
			Description: "Fan speed update failed: simulated transient error.",
			State:       store.ActionLogStateFailed,
		})
		return fmt.Errorf("simulated fan speed failure")
	}

	s.AddActionLog(ctx, store.NewActionLog{
		StepID:      "set-fan-speed",
		Description: fmt.Sprintf("Fan speed set to %d%%.", targetSpeed),
		State:       store.ActionLogStateSuccess,
	})
	return nil
}
