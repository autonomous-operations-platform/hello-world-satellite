package observers

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/autonomous-operations-platform/autonomous-operations-satellite/pkg/framework/registry"
	"github.com/autonomous-operations-platform/autonomous-operations-satellite/pkg/framework/store"
)

func init() { registry.RegisterObserver(&TemperatureThreshold{}) }

const (
	highThreshold = 75.0
	lowThreshold  = 60.0
)

// TemperatureThreshold watches cpu-temp telemetry and emits temperature-high
// above 75°C and temperature-low below 60°C.
type TemperatureThreshold struct{}

func (o *TemperatureThreshold) Name() string { return "temperature-threshold" }

func (o *TemperatureThreshold) Filter() store.TelemetryFilter {
	return store.TelemetryFilter{
		Attributes: map[string]string{"source": "cpu-temp"},
	}
}

func (o *TemperatureThreshold) Detect(ctx context.Context, logger *slog.Logger, trigger store.TelemetryRecord, s store.ObserverStore) error {
	celsius, err := strconv.ParseFloat(trigger.Attributes["value"], 64)
	if err != nil {
		return fmt.Errorf("invalid temperature value: %w", err)
	}

	switch {
	case celsius > highThreshold:
		logger.InfoContext(ctx, "temperature above accepted range", "celsius", celsius, "high_threshold", highThreshold)
		_, err = s.AddInsight(ctx, store.NewInsight{
			Pattern:         "temperature-high",
			Severity:        store.SeverityWarning,
			Summary:         fmt.Sprintf("CPU temperature %.1f°C exceeds high threshold of %.0f°C.", celsius, highThreshold),
			WindowStart:     trigger.ReceivedAt,
			WindowEnd:       trigger.ReceivedAt,
			SourceRecordIDs: []store.RecordID{trigger.ID},
		})
	case celsius < lowThreshold:
		logger.InfoContext(ctx, "temperature below accepted range", "celsius", celsius, "low_threshold", lowThreshold)
		_, err = s.AddInsight(ctx, store.NewInsight{
			Pattern:         "temperature-low",
			Severity:        store.SeverityInfo,
			Summary:         fmt.Sprintf("CPU temperature %.1f°C dropped below low threshold of %.0f°C.", celsius, lowThreshold),
			WindowStart:     trigger.ReceivedAt,
			WindowEnd:       trigger.ReceivedAt,
			SourceRecordIDs: []store.RecordID{trigger.ID},
		})
	default:
		logger.InfoContext(ctx, "temperature within accepted range", "celsius", celsius, "low_threshold", lowThreshold, "high_threshold", highThreshold)
	}
	return err
}
