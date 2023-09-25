package resource

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestPerfExecutor(t *testing.T) {
	t.Parallel()
	var timeout = 10 * time.Second

	content, err := os.ReadFile("/proc/sys/kernel/perf_event_paranoid")
	if err != nil {
		t.Fatal(err)
	}

	paranoid, err := strconv.Atoi(strings.Trim(string(content), "\n"))
	if err != nil {
		t.Fatal(err)
	}
	if paranoid >= 0 {
		t.Skipf("Skip test because perf_event_paranoid is %q", paranoid)
	}

	_, err = exec.LookPath("perf")
	if err != nil {
		t.Skip("Skip test because perf is not installed")
	}

	logger := slog.Default()
	perfExecuter, err := NewPerfExecuter(logger)
	if err != nil {
		t.Fatal(err)
	}
	if perfExecuter.binPath == "" {
		t.Skip("Skip test because perf is not installed")
	}

	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "yes")
	err = cmd.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Cancel()

	pid := cmd.Process.Pid
	path, err := perfExecuter.ExecRecord(ctx, os.TempDir(), pid, timeout)
	if err != nil {
		t.Fatal(err)
	}

	_, err = perfExecuter.ExecScript(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
}
