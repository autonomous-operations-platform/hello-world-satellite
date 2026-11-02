package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/autonomous-operations-platform/autonomous-operations-satellite/pkg/framework/satellite"
	"github.com/autonomous-operations-platform/autonomous-operations-satellite/pkg/store/memstore"

	_ "github.com/autonomous-operations-platform/hello-world-satellite/collectors"
	_ "github.com/autonomous-operations-platform/hello-world-satellite/observers"
	_ "github.com/autonomous-operations-platform/hello-world-satellite/operators"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := satellite.Run(ctx, satellite.Config{
		Store:  memstore.New(),
		Logger: slog.Default(),
	})
	if err != nil {
		slog.Error("satellite run failed", "err", err)
		os.Exit(1)
	}
	/*
			A more elaborate setup could use Apache Doris instead of the memory store.
			Add the module via `go get github.com/autonomous-operations-platform/autonomous-operations-satellite/pkg/store/doris`
			The following code loads a configuration object that is populated from environment variables, defined in internal/utility.
			Then, it sets up file-logging that supports rotation by logrotate through SIGHUP.
			Afterwards, it connects to Apache Doris (can be started from the enclosed Docker Compose file).

			// Loading the configuration
			cfg, err := utility.LoadConfig()
			if err != nil {
				slog.Error("config load failed", "err", err)
				os.Exit(1)
			}

			// setting up logging
			logWriter, err := utility.SetupLogging(cfg)
			if err != nil {
				slog.Error("logging setup failed", "err", err)
				os.Exit(1)
			}
			defer logWriter.Close()

			// logrotate (create mode) renames the live log file and signals SIGHUP;
			// reopen so writes continue into the fresh file.
			hup := make(chan os.Signal, 1)
			signal.Notify(hup, syscall.SIGHUP)
			go func() {
				for range hup {
					if err := logWriter.Reopen(); err != nil {
						slog.Error("log file reopen failed", "component", "main", "err", err)
						continue
					}
					slog.Info("log file reopened after SIGHUP", "component", "main", "path", cfg.LogFilePath)
				}
			}()

		 	st, err := doris.Connect(ctx, doris.Params{
				Addr:     cfg.DorisAddr,
				User:     cfg.DorisUser,
				Password: cfg.DorisPassword,
				DB:       cfg.DorisDB,
			}, slog.Default())
			if err != nil {
				slog.Error("failed to connect to Doris", "err", err)
				os.Exit(1)
			}
			defer st.Close()

			err = satellite.Run(ctx, satellite.Config{
				Store:  st,
				Logger: slog.Default(),
			})
	*/
}
