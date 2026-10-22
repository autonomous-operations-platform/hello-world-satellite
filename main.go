package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/autonomous-operations/autonomous-operations-satellite/pkg/framework/satellite"
	"github.com/autonomous-operations/autonomous-operations-satellite/pkg/store/memstore"

	_ "github.com/autonomous-operations/aos-demo/satellites/hello-world-satellite/collectors"
	_ "github.com/autonomous-operations/aos-demo/satellites/hello-world-satellite/observers"
	_ "github.com/autonomous-operations/aos-demo/satellites/hello-world-satellite/operators"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Store could be replaced with the Doris store or any
	// other other implementation of Satellite Record Store
	err := satellite.Run(ctx, satellite.Config{
		Store:  memstore.New(),
		Logger: slog.Default(),
	})
	if err != nil {
		slog.Error("satellite run failed", "err", err)
		os.Exit(1)
	}
}
