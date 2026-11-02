package utility

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
)

// LogWriter is an io.Writer backed by a file that can be reopened on
// SIGHUP so logging continues into a fresh file after logrotate rotates the old
// one out from under us. The mutex serializes writes against a concurrent
// reopen, which matters because every subsystem goroutine logs through the same
// slog default handler.
type LogWriter struct {
	mu   sync.Mutex
	path string
	f    *os.File
}

func newLogWriter(path string) (*LogWriter, error) {
	w := &LogWriter{path: path}
	if err := w.open(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *LogWriter) open() error {
	f, err := os.OpenFile(w.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open log file %q: %w", w.path, err)
	}
	w.f = f
	return nil
}

func (w *LogWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.f.Write(p)
}

// Reopen closes the current file descriptor and opens the configured path
// again. Used on SIGHUP: logrotate (in "create" mode) renames the live file and
// creates a fresh one, and this makes subsequent writes land in the new file
// instead of the rotated-away descriptor.
func (w *LogWriter) Reopen() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if err := w.f.Close(); err != nil {
		return fmt.Errorf("close log file during reopen: %w", err)
	}
	return w.open()
}

func (w *LogWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.f.Close()
}

// SetupLogging opens the configured log file and installs a JSON slog handler
// as the process-wide default. Output is written to the file (durable, for the
// log connector to ship) and mirrored to stderr so logs remain visible in the
// terminal and via `docker logs`.
func SetupLogging(cfg Config) (*LogWriter, error) {
	w, err := newLogWriter(cfg.LogFilePath)
	if err != nil {
		return nil, err
	}
	slog.SetDefault(newJSONLogger(w, os.Stderr, cfg.LogLevel))
	return w, nil
}

// newJSONLogger builds a JSON slog logger that fans every record out to both
// destinations. Split out from SetupLogging so tests can supply an in-memory
// console writer instead of stderr.
func newJSONLogger(file, console io.Writer, level slog.Level) *slog.Logger {
	dst := io.MultiWriter(file, console)
	return slog.New(slog.NewJSONHandler(dst, &slog.HandlerOptions{Level: level}))
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
