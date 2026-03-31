package processmanager

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/openzerg/hydralisk/internal/core/types"
)

type internalHandle struct {
	handle   *types.ProcessHandle
	cmd      *exec.Cmd
	stdout   *os.File
	stderr   *os.File
	exitCode int
	done     chan struct{}
}

type ProcessManager struct {
	mu        sync.RWMutex
	processes map[string]*internalHandle
	outputDir string
}

func NewProcessManager(outputDir string) *ProcessManager {
	if outputDir == "" {
		outputDir = "/tmp/openzerg/processes"
	}
	return &ProcessManager{
		processes: make(map[string]*internalHandle),
		outputDir: outputDir,
	}
}

func (pm *ProcessManager) Spawn(ctx context.Context, command string, opts types.SpawnOptions) (*types.ProcessHandle, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	processID := uuid.New().String()
	outputDir := filepath.Join(pm.outputDir, processID)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	args := parseCommand(command)
	if len(args) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = opts.Workdir

	env := os.Environ()
	env = append(env, fmt.Sprintf("OPENZERG_PROCESS_ID=%s", processID))
	for k, v := range opts.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = env

	stdout, err := os.Create(filepath.Join(outputDir, "stdout.log"))
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout file: %w", err)
	}
	stderr, err := os.Create(filepath.Join(outputDir, "stderr.log"))
	if err != nil {
		stdout.Close()
		return nil, fmt.Errorf("failed to create stderr file: %w", err)
	}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Start(); err != nil {
		stdout.Close()
		stderr.Close()
		return nil, fmt.Errorf("failed to start process: %w", err)
	}

	handle := &types.ProcessHandle{
		ID:        processID,
		UnitName:  fmt.Sprintf("openzerg-%s.scope", processID[:8]),
		OutputDir: outputDir,
		StartedAt: time.Now(),
		TimeoutMs: opts.Timeout,
		SessionID: opts.SessionID,
	}

	ih := &internalHandle{
		handle:   handle,
		cmd:      cmd,
		stdout:   stdout,
		stderr:   stderr,
		exitCode: -1,
		done:     make(chan struct{}),
	}

	pm.processes[processID] = ih

	go func() {
		defer close(ih.done)
		err := cmd.Wait()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				ih.exitCode = exitErr.ExitCode()
			}
		} else {
			ih.exitCode = 0
		}
		stdout.Close()
		stderr.Close()
	}()

	return handle, nil
}

func (pm *ProcessManager) Wait(processID string, timeout int) (*types.ProcessResult, error) {
	pm.mu.RLock()
	ih, ok := pm.processes[processID]
	pm.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("process %s not found", processID)
	}

	start := time.Now()

	var ctx context.Context
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
		defer cancel()
	} else {
		ctx = context.Background()
	}

	select {
	case <-ih.done:
		duration := time.Since(start).Milliseconds()
		status := types.ProcessStatusCompleted
		if ih.exitCode != 0 {
			status = types.ProcessStatusFailed
		}
		return &types.ProcessResult{
			ExitCode:   ih.exitCode,
			Status:     status,
			DurationMs: duration,
		}, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("process timed out")
	}
}

func (pm *ProcessManager) Kill(processID, signal string) error {
	pm.mu.RLock()
	ih, ok := pm.processes[processID]
	pm.mu.RUnlock()

	if !ok {
		return fmt.Errorf("process %s not found", processID)
	}

	if signal == "" {
		signal = "SIGTERM"
	}

	return ih.cmd.Process.Signal(parseSignal(signal))
}

func (pm *ProcessManager) GetStatus(processID string) types.ProcessStatus {
	pm.mu.RLock()
	ih, ok := pm.processes[processID]
	pm.mu.RUnlock()

	if !ok {
		return ""
	}

	select {
	case <-ih.done:
		if ih.exitCode == 0 {
			return types.ProcessStatusCompleted
		}
		return types.ProcessStatusFailed
	default:
		return types.ProcessStatusRunning
	}
}

func (pm *ProcessManager) GetOutput(processID string, stream string, offset, limit int) (*types.ProcessOutput, error) {
	pm.mu.RLock()
	ih, ok := pm.processes[processID]
	pm.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("process %s not found", processID)
	}

	if stream == "" {
		stream = "stdout"
	}
	if limit == 0 {
		limit = 50
	}

	var filename string
	switch stream {
	case "stdout":
		filename = filepath.Join(ih.handle.OutputDir, "stdout.log")
	case "stderr":
		filename = filepath.Join(ih.handle.OutputDir, "stderr.log")
	case "both":
		return pm.readBothOutputs(ih.handle, offset, limit)
	default:
		filename = filepath.Join(ih.handle.OutputDir, "stdout.log")
	}

	return pm.readOutputFile(processID, stream, filename, offset, limit)
}

func (pm *ProcessManager) readOutputFile(processID, stream, filename string, offset, limit int) (*types.ProcessOutput, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []types.ProcessOutputLine
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		if lineNum >= offset && lineNum < offset+limit {
			lines = append(lines, types.ProcessOutputLine{
				Num:     lineNum,
				Content: scanner.Text(),
			})
		}
		lineNum++
	}

	return &types.ProcessOutput{
		ProcessID:  processID,
		Stream:     stream,
		Lines:      lines,
		TotalLines: lineNum,
		HasMore:    lineNum > offset+limit,
		Offset:     offset,
		Limit:      limit,
	}, nil
}

func (pm *ProcessManager) readBothOutputs(handle *types.ProcessHandle, offset, limit int) (*types.ProcessOutput, error) {
	stdout, err := pm.readOutputFile(handle.ID, "stdout", filepath.Join(handle.OutputDir, "stdout.log"), offset, limit)
	if err != nil {
		return nil, err
	}
	stderr, err := pm.readOutputFile(handle.ID, "stderr", filepath.Join(handle.OutputDir, "stderr.log"), offset, limit)
	if err != nil {
		return nil, err
	}

	lines := append(stdout.Lines, stderr.Lines...)
	return &types.ProcessOutput{
		ProcessID:  handle.ID,
		Stream:     "both",
		Lines:      lines,
		TotalLines: stdout.TotalLines + stderr.TotalLines,
		HasMore:    stdout.HasMore || stderr.HasMore,
		Offset:     offset,
		Limit:      limit,
	}, nil
}

func (pm *ProcessManager) GetOutputStats(processID string) (*types.OutputStats, error) {
	pm.mu.RLock()
	ih, ok := pm.processes[processID]
	pm.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("process %s not found", processID)
	}

	stdoutInfo, _ := os.Stat(filepath.Join(ih.handle.OutputDir, "stdout.log"))
	stderrInfo, _ := os.Stat(filepath.Join(ih.handle.OutputDir, "stderr.log"))

	stdoutSize := int64(0)
	stderrSize := int64(0)
	if stdoutInfo != nil {
		stdoutSize = stdoutInfo.Size()
	}
	if stderrInfo != nil {
		stderrSize = stderrInfo.Size()
	}

	return &types.OutputStats{
		StdoutSize: stdoutSize,
		StderrSize: stderrSize,
	}, nil
}

func (pm *ProcessManager) ListProcesses() []*types.ProcessHandle {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	handles := make([]*types.ProcessHandle, 0, len(pm.processes))
	for _, ih := range pm.processes {
		handles = append(handles, ih.handle)
	}
	return handles
}

func parseCommand(command string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(command); i++ {
		char := command[i]
		if inQuotes {
			if char == quoteChar {
				inQuotes = false
			} else {
				current.WriteByte(char)
			}
		} else if char == '"' || char == '\'' {
			inQuotes = true
			quoteChar = char
		} else if char == ' ' || char == '\t' {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(char)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

func parseSignal(signal string) os.Signal {
	switch signal {
	case "SIGTERM":
		return os.Interrupt
	case "SIGKILL":
		return os.Kill
	default:
		return os.Interrupt
	}
}
