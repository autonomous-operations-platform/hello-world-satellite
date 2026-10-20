package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.tools.sap/autonomous-operations/aos-framework/pkg/framework/satellite"
	"github.tools.sap/autonomous-operations/aos-framework/pkg/store/memstore"

	_ "github.tools.sap/autonomous-operations/aos-demo/satellites/hello-world-satellite/collectors"
	_ "github.tools.sap/autonomous-operations/aos-demo/satellites/hello-world-satellite/observers"
	_ "github.tools.sap/autonomous-operations/aos-demo/satellites/hello-world-satellite/operators"
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
