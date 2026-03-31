package types

type LLMMessageRole string

const (
	LLMMessageRoleSystem    LLMMessageRole = "system"
	LLMMessageRoleUser      LLMMessageRole = "user"
	LLMMessageRoleAssistant LLMMessageRole = "assistant"
	LLMMessageRoleTool      LLMMessageRole = "tool"
	LLMMessageRoleThinking  LLMMessageRole = "thinking"
)

type LLMToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type LLMToolCall struct {
	ID       string               `json:"id"`
	Type     string               `json:"type"`
	Function *LLMToolCallFunction `json:"function"`
}

type LLMMessage struct {
	Role       LLMMessageRole `json:"role"`
	Content    *string        `json:"content"`
	ToolCalls  []LLMToolCall  `json:"tool_calls,omitempty"`
	ToolCallID *string        `json:"tool_call_id,omitempty"`
	Name       *string        `json:"name,omitempty"`
}

type ChatCompletionToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type ChatCompletionTool struct {
	Type     string                      `json:"type"`
	Function *ChatCompletionToolFunction `json:"function"`
}

type ChatCompletionRequest struct {
	Model       string                `json:"model"`
	Messages    []LLMMessage          `json:"messages"`
	Tools       []*ChatCompletionTool `json:"tools,omitempty"`
	Stream      *bool                 `json:"stream,omitempty"`
	Temperature *float64              `json:"temperature,omitempty"`
	MaxTokens   *int                  `json:"max_tokens,omitempty"`
	TopP        *float64              `json:"top_p,omitempty"`
}

type ChatCompletionChoice struct {
	Index        int         `json:"index"`
	Message      *LLMMessage `json:"message"`
	FinishReason *string     `json:"finish_reason"`
}

type ChatCompletionUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChatCompletionResponse struct {
	ID      string                  `json:"id"`
	Object  string                  `json:"object"`
	Created int64                   `json:"created"`
	Model   string                  `json:"model"`
	Choices []*ChatCompletionChoice `json:"choices"`
	Usage   *ChatCompletionUsage    `json:"usage,omitempty"`
}

type StreamChunkDelta struct {
	Role      *string `json:"role,omitempty"`
	Content   *string `json:"content,omitempty"`
	ToolCalls []struct {
		Index    int     `json:"index"`
		ID       *string `json:"id,omitempty"`
		Type     *string `json:"type,omitempty"`
		Function *struct {
			Name      *string `json:"name,omitempty"`
			Arguments *string `json:"arguments,omitempty"`
		} `json:"function,omitempty"`
	} `json:"tool_calls,omitempty"`
}

type StreamChunkChoice struct {
	Index        int               `json:"index"`
	Delta        *StreamChunkDelta `json:"delta"`
	FinishReason *string           `json:"finish_reason"`
}

type StreamChunk struct {
	ID      string               `json:"id"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Model   string               `json:"model"`
	Choices []*StreamChunkChoice `json:"choices"`
}

type LLMConfig struct {
	BaseURL     string                 `json:"base_url"`
	APIKey      string                 `json:"api_key"`
	Model       string                 `json:"model"`
	MaxTokens   *int                   `json:"max_tokens,omitempty"`
	Temperature *float64               `json:"temperature,omitempty"`
	TopP        *float64               `json:"top_p,omitempty"`
	TopK        *int                   `json:"top_k,omitempty"`
	ExtraParams map[string]interface{} `json:"extra_params,omitempty"`
}

type Provider struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	BaseURL     string   `json:"base_url"`
	APIKey      string   `json:"api_key"`
	Model       string   `json:"model"`
	MaxTokens   *int     `json:"max_tokens,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`
	TopP        *float64 `json:"top_p,omitempty"`
	TopK        *int     `json:"top_k,omitempty"`
	ExtraParams *string  `json:"extra_params"`
	IsActive    bool     `json:"is_active"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}
