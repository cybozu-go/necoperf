package resource

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cybozu-go/necoperf/internal/constants"
)

const (
	perfName = "perf"
)

type PerfExecuter struct {
	logger  *slog.Logger
	binPath string
}

func lookupBinary() (string, error) {
	path, err := exec.LookPath(perfName)
	if err != nil {
		return "", err
	}
	return path, nil
}

func NewPerfExecuter(logger *slog.Logger) (*PerfExecuter, error) {
	path, err := lookupBinary()
	if err != nil {
		return nil, err
	}

	return &PerfExecuter{
		logger:  logger,
		binPath: path,
	}, nil
}

func (p *PerfExecuter) ExecRecord(ctx context.Context, workDir string, pid int, timeout time.Duration) (string, error) {
	profilingFileName := fmt.Sprintf("necoperf-%d.data", pid)
	profilingPath := filepath.Join(workDir + "/" + profilingFileName)

	t := timeout.Seconds()
	perfArgs := []string{
		constants.RecordSubcommand,
		"-ag",
		"-F", "99",
		"--call-graph", "dwarf",
		"-p", strconv.Itoa(pid),
		"-o", profilingPath,
		"--", "sleep", strconv.Itoa(int(t)),
	}
	c := exec.CommandContext(ctx, p.binPath, perfArgs...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	p.logger.Info("Executing perf record", "cmd", c.String())

	return profilingPath, c.Run()
}

func (p *PerfExecuter) GetEvent(ctx context.Context, filePath string) (*bytes.Buffer, error) {
	var stdoutBuff bytes.Buffer
	perfArgs := []string{
		constants.ScriptSubcommand,
		"-F", "event",
		"-i", filePath,
	}

	c := exec.CommandContext(ctx, p.binPath, perfArgs...)
	c.Stdout = &stdoutBuff
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		return nil, err
	}
	return &stdoutBuff, nil
}

// Check if events are contained in the perf.data file.
func (p *PerfExecuter) HasPerfEvent(ctx context.Context, buf *bytes.Buffer) bool {
	return strings.Contains(buf.String(), constants.CpuEvent)
}

func (p *PerfExecuter) ExecScript(ctx context.Context, filePath string) (string, error) {
	var stdoutBuff bytes.Buffer

	buf, err := p.GetEvent(ctx, filePath)
	if err != nil {
		return "", err
	}

	if !p.HasPerfEvent(ctx, buf) {
		return "", fmt.Errorf("perf.data file does not contain events")
	}

	perfArgs := []string{
		constants.ScriptSubcommand,
		"--no-inline",
		"-i", filePath,
	}

	c := exec.CommandContext(ctx, p.binPath, perfArgs...)
	c.Stdout = &stdoutBuff
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return "", err
	}

	scriptFilePath := strings.Replace(filePath, "data", "script", 1)
	f, err := os.OpenFile(scriptFilePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = f.Write(stdoutBuff.Bytes())
	if err != nil {
		return "", err
	}

	return scriptFilePath, nil
}
