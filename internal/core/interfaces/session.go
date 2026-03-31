package interfaces

import (
	"context"

	"github.com/openzerg/hydralisk/internal/core/types"
)

type ISessionManager interface {
	Create(ctx context.Context, data *types.CreateSessionData) (*types.Session, error)
	Get(ctx context.Context, id string) (*types.Session, error)
	List(ctx context.Context) ([]*types.Session, error)
	Update(ctx context.Context, id string, data map[string]interface{}) error
	Delete(ctx context.Context, id string) error

	AddMessage(ctx context.Context, sessionID string, message *types.CreateMessageData) (*types.Message, error)
	GetMessages(ctx context.Context, sessionID string) ([]*types.Message, error)

	Start(ctx context.Context, id string) error
	Interrupt(ctx context.Context, id string, message *string) error

	GetState(id string) types.SessionState
	SetActive(sessionID string)
	GetActive() *string
}

type ISessionProcessor interface {
	Process(ctx context.Context, sessionID, userInput string) error
	HandleToolResult(ctx context.Context, sessionID, toolCallID string, result *types.ToolResult) error
	BuildContext(ctx context.Context, sessionID string) ([]*types.LLMMessage, error)
}
