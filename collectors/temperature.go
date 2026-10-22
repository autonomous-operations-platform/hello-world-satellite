package collectors

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/autonomous-operations/autonomous-operations-satellite/pkg/framework/registry"
	"github.com/autonomous-operations/autonomous-operations-satellite/pkg/framework/store"
)

func init() { registry.RegisterCollector(&TemperatureCollector{}) }

// TemperatureCollector reads from the mock temperature sensor every 5s.
type TemperatureCollector struct{}

func (c *TemperatureCollector) Name() string { return "temperature-collector" }

func (c *TemperatureCollector) Setup(ctx context.Context, logger *slog.Logger, s store.CollectorStore) {
	go func() {
		t := time.NewTicker(5 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				c.collect(ctx, logger, s)
			}
		}
	}()
}

func (c *TemperatureCollector) collect(ctx context.Context, logger *slog.Logger, s store.CollectorStore) {
	celsius := readTemperature()
	logger.InfoContext(ctx, "temperature reading", "celsius", fmt.Sprintf("%.1f", celsius))
	s.AddTelemetry(ctx, []store.NewTelemetryRecord{{
		Signal: store.SignalMetrics,
		Attributes: map[string]string{
			"source": "cpu-temp",
			"value":  fmt.Sprintf("%.1f", celsius),
			"unit":   "celsius",
		},
	}})
}

// Mock sensor state — stand-ins for a real hardware/API call.
var (
	mockCelsius   = 60.0
	mockDirection = 14.3
)

// readTemperature simulates a sensor API: steps by 14.3°C, bouncing between 30 and 95.
func readTemperature() float64 {
	mockCelsius += mockDirection
	if mockCelsius >= 95 {
		mockCelsius = 95
		mockDirection = -14.3
	} else if mockCelsius <= 30 {
		mockCelsius = 30
		mockDirection = 14.3
	}
	return mockCelsius
}
