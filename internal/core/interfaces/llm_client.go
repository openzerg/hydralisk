package interfaces

import (
	"context"

	"github.com/openzerg/hydralisk/internal/core/types"
)

type ILLMClient interface {
	Complete(ctx context.Context, messages []*types.LLMMessage, tools []*types.ToolDefinition, options map[string]interface{}) (*types.ChatCompletionResponse, error)
	Stream(ctx context.Context, messages []*types.LLMMessage, tools []*types.ToolDefinition, options map[string]interface{}) (<-chan *types.StreamChunk, error)
	UpdateConfig(config map[string]interface{})
	GetConfig() *types.LLMConfig
	CountTokens(messages []*types.LLMMessage) int
}
