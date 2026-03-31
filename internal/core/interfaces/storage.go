package interfaces

import (
	"context"

	"github.com/openzerg/hydralisk/internal/core/types"
)

type IStorage interface {
	CreateSession(ctx context.Context, data *types.CreateSessionData) (*types.Session, error)
	GetSession(ctx context.Context, id string) (*types.Session, error)
	ListSessions(ctx context.Context, filter *types.SessionFilter) ([]*types.Session, error)
	UpdateSession(ctx context.Context, id string, data *types.UpdateSessionData) error
	DeleteSession(ctx context.Context, id string) error
	IncrementMessageCount(ctx context.Context, sessionID string) error

	SaveMessage(ctx context.Context, message *types.CreateMessageData) (*types.Message, error)
	GetMessages(ctx context.Context, sessionID string, filter *types.MessageFilter) ([]*types.Message, error)
	DeleteMessages(ctx context.Context, sessionID string) error

	SaveProcess(ctx context.Context, process *types.CreateProcessData) (*types.Process, error)
	GetProcess(ctx context.Context, id string) (*types.Process, error)
	ListProcesses(ctx context.Context, filter map[string]interface{}) ([]*types.Process, error)
	UpdateProcessStatus(ctx context.Context, id string, status types.ProcessStatus, exitCode *int) error
	UpdateProcessOutputStats(ctx context.Context, id string, stats *types.OutputStats) error

	CreateTodo(ctx context.Context, sessionID, content, priority string) (*types.Todo, error)
	GetTodos(ctx context.Context, sessionID string) ([]*types.Todo, error)
	UpdateTodo(ctx context.Context, id string, data map[string]interface{}) error
	DeleteTodo(ctx context.Context, id string) error

	GetProvider(ctx context.Context, name string) (*types.Provider, error)
	ListProviders(ctx context.Context) ([]*types.Provider, error)
	SaveProvider(ctx context.Context, provider *types.Provider) (*types.Provider, error)
	UpdateProvider(ctx context.Context, id string, data map[string]interface{}) error
	DeleteProvider(ctx context.Context, id string) error

	RecordActivity(ctx context.Context, sessionID *string, activityType, description string, details map[string]interface{}) error
	GetActivities(ctx context.Context, sessionID string) ([]*types.Activity, error)
}
