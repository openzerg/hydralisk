package interfaces

import (
	"context"

	"github.com/openzerg/hydralisk/internal/core/types"
)

type ISessionService interface {
	Get(ctx context.Context, id string) (*types.Session, error)
	Create(ctx context.Context, name, purpose string, systemPrompt *string, parentID *string) (*types.Session, error)
	Delete(ctx context.Context, id string) error
	GetMessages(ctx context.Context, sessionID string, offset, limit int) (*types.MessageListResponse, error)
	SendChat(ctx context.Context, sessionID, content string) error
	Interrupt(ctx context.Context, sessionID, message string) error
	GetContext(ctx context.Context, sessionID string) (*types.SessionContext, error)
	GetTodos(ctx context.Context, sessionID string) ([]*types.Todo, error)
	GetActivities(ctx context.Context, sessionID string) ([]*types.Activity, error)
}

type ToolContext struct {
	SessionID      string
	SessionName    string
	WorkingDir     string
	ProcessManager IProcessManager
	ToolRegistry   IToolRegistry
	EventBus       IEventBus
	Storage        IStorage
	SessionService ISessionService
	LLMClient      ILLMClient
	MessageBus     any
}

type ITool interface {
	Name() string
	Description() string
	Parameters() *types.JSONSchema
	Execute(ctx context.Context, args map[string]interface{}, toolCtx *ToolContext) (*types.ToolResult, error)
}

type IToolRegistry interface {
	Register(tool ITool)
	Unregister(name string)
	Get(name string) ITool
	GetDefinitions() []*types.ToolDefinition
	Execute(ctx context.Context, name string, args map[string]interface{}, toolCtx *ToolContext) (*types.ToolResult, error)
	Has(name string) bool
	List() []string
}
