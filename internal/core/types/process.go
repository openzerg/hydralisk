package types

import "time"

type ProcessStatus string

const (
	ProcessStatusRunning   ProcessStatus = "Running"
	ProcessStatusCompleted ProcessStatus = "Completed"
	ProcessStatusFailed    ProcessStatus = "Failed"
	ProcessStatusTimeout   ProcessStatus = "Timeout"
	ProcessStatusKilled    ProcessStatus = "Killed"
)

type Process struct {
	ID              string        `json:"id" bun:",pk"`
	Command         string        `json:"command"`
	Cwd             string        `json:"cwd"`
	Status          ProcessStatus `json:"status"`
	ExitCode        *int          `json:"exit_code"`
	StartedAt       string        `json:"started_at"`
	FinishedAt      *string       `json:"finished_at"`
	ParentSessionID *string       `json:"parent_session_id"`
	UnitName        string        `json:"unit_name"`
	TimeoutMs       int           `json:"timeout_ms"`
	OutputDir       string        `json:"output_dir"`
	StdoutSize      int64         `json:"stdout_size"`
	StderrSize      int64         `json:"stderr_size"`
	StdoutLines     int           `json:"stdout_lines"`
	StderrLines     int           `json:"stderr_lines"`
}

type ProcessHandle struct {
	ID        string    `json:"id"`
	PID       int       `json:"pid"`
	UnitName  string    `json:"unit_name"`
	OutputDir string    `json:"output_dir"`
	StartedAt time.Time `json:"started_at"`
	TimeoutMs int       `json:"timeout_ms"`
	SessionID *string   `json:"session_id"`
}

type ProcessResult struct {
	ExitCode   int           `json:"exit_code"`
	Status     ProcessStatus `json:"status"`
	DurationMs int64         `json:"duration_ms"`
}

type OutputStats struct {
	StdoutSize  int64 `json:"stdout_size"`
	StderrSize  int64 `json:"stderr_size"`
	StdoutLines int   `json:"stdout_lines"`
	StderrLines int   `json:"stderr_lines"`
}

type ProcessOutputLine struct {
	Num     int    `json:"num"`
	Content string `json:"content"`
}

type ProcessOutput struct {
	ProcessID  string              `json:"process_id"`
	Stream     string              `json:"stream"`
	Lines      []ProcessOutputLine `json:"lines"`
	TotalLines int                 `json:"total_lines"`
	HasMore    bool                `json:"has_more"`
	Offset     int                 `json:"offset"`
	Limit      int                 `json:"limit"`
}

type CreateProcessData struct {
	ID              string  `json:"id"`
	Command         string  `json:"command"`
	Cwd             string  `json:"cwd"`
	ParentSessionID *string `json:"parent_session_id"`
	UnitName        string  `json:"unit_name"`
	OutputDir       string  `json:"output_dir"`
	TimeoutMs       *int    `json:"timeout_ms"`
}

type SpawnOptions struct {
	Workdir   string            `json:"workdir"`
	Timeout   int               `json:"timeout"`
	Env       map[string]string `json:"env"`
	SessionID *string           `json:"session_id"`
}
