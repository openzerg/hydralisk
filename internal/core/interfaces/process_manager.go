package interfaces

import (
	"context"

	"github.com/openzerg/hydralisk/internal/core/types"
)

type IProcessManager interface {
	Spawn(ctx context.Context, command string, opts types.SpawnOptions) (*types.ProcessHandle, error)
	Wait(processID string, timeout int) (*types.ProcessResult, error)
	Kill(processID, signal string) error
	GetStatus(processID string) types.ProcessStatus
	GetOutput(processID string, stream string, offset, limit int) (*types.ProcessOutput, error)
	GetOutputStats(processID string) (*types.OutputStats, error)
	ListProcesses() []*types.ProcessHandle
}
