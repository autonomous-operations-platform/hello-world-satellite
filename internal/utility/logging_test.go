package utility

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestSetupLoggingWritesJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sat.log")
	w, err := SetupLogging(Config{LogFilePath: path, LogLevel: slog.LevelInfo})
	if err != nil {
		t.Fatalf("SetupLogging: %v", err)
	}
	defer w.Close()

	slog.Info("hello", "component", "test", "n", 42)

	rec := readLastJSONLine(t, path)
	if rec["msg"] != "hello" {
		t.Errorf("msg = %v, want hello", rec["msg"])
	}
	if rec["level"] != "INFO" {
		t.Errorf("level = %v, want INFO", rec["level"])
	}
	if rec["component"] != "test" {
		t.Errorf("component = %v, want test", rec["component"])
	}
	if rec["n"] != float64(42) {
		t.Errorf("n = %v, want 42", rec["n"])
	}
}

func TestSetupLoggingRespectsLevel(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sat.log")
	w, err := SetupLogging(Config{LogFilePath: path, LogLevel: slog.LevelWarn})
	if err != nil {
		t.Fatalf("SetupLogging: %v", err)
	}
	defer w.Close()

	slog.Info("filtered out")
	slog.Warn("kept")

	lines := readJSONLines(t, path)
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1 (info should be filtered)", len(lines))
	}
	if lines[0]["msg"] != "kept" {
		t.Errorf("msg = %v, want kept", lines[0]["msg"])
	}
}

// TestReopenAfterRotate simulates logrotate "create" mode: the live file is
// renamed away, reopen() is called (as it would be on SIGHUP), and subsequent
// writes must land in a freshly created file at the original path.
func TestReopenAfterRotate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sat.log")
	w, err := SetupLogging(Config{LogFilePath: path, LogLevel: slog.LevelInfo})
	if err != nil {
		t.Fatalf("SetupLogging: %v", err)
	}
	defer w.Close()

	slog.Info("before rotate")

	// logrotate renames the current file out of the way.
	if err := os.Rename(path, path+".1"); err != nil {
		t.Fatalf("rename: %v", err)
	}
	if err := w.Reopen(); err != nil {
		t.Fatalf("reopen: %v", err)
	}

	slog.Info("after rotate")

	// The fresh file must exist and contain only the post-reopen record.
	lines := readJSONLines(t, path)
	if len(lines) != 1 {
		t.Fatalf("fresh file has %d lines, want 1", len(lines))
	}
	if lines[0]["msg"] != "after rotate" {
		t.Errorf("msg = %v, want 'after rotate'", lines[0]["msg"])
	}

	// The rotated file must still hold the pre-rotate record.
	rotated := readJSONLines(t, path+".1")
	if len(rotated) != 1 || rotated[0]["msg"] != "before rotate" {
		t.Errorf("rotated file = %v, want single 'before rotate' record", rotated)
	}
}

// TestLoggingTeesToFileAndConsole verifies each record lands in both the log
// file (durable) and the console writer (terminal / docker logs).
func TestLoggingTeesToFileAndConsole(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sat.log")
	w, err := newLogWriter(path)
	if err != nil {
		t.Fatalf("newReopenableWriter: %v", err)
	}
	defer w.Close()

	var console bytes.Buffer
	logger := newJSONLogger(w, &console, slog.LevelInfo)
	logger.Info("dual sink", "component", "test")

	// File sink.
	fileRec := readLastJSONLine(t, path)
	if fileRec["msg"] != "dual sink" {
		t.Errorf("file msg = %v, want 'dual sink'", fileRec["msg"])
	}

	// Console sink.
	var consoleRec map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(console.Bytes()), &consoleRec); err != nil {
		t.Fatalf("console unmarshal %q: %v", console.String(), err)
	}
	if consoleRec["msg"] != "dual sink" {
		t.Errorf("console msg = %v, want 'dual sink'", consoleRec["msg"])
	}
}

func readJSONLines(t *testing.T, path string) []map[string]any {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()

	var out []map[string]any
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if len(sc.Bytes()) == 0 {
			continue
		}
		var rec map[string]any
		if err := json.Unmarshal(sc.Bytes(), &rec); err != nil {
			t.Fatalf("unmarshal %q: %v", sc.Text(), err)
		}
		out = append(out, rec)
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan: %v", err)
	}
	return out
}

func readLastJSONLine(t *testing.T, path string) map[string]any {
	t.Helper()
	lines := readJSONLines(t, path)
	if len(lines) == 0 {
		t.Fatalf("no log lines in %s", path)
	}
	return lines[len(lines)-1]
}
