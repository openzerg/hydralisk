package types

type MessageRole string

const (
	MessageRoleSystem    MessageRole = "system"
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleTool      MessageRole = "tool"
	MessageRoleThinking  MessageRole = "thinking"
)

type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type Message struct {
	ID        string      `json:"id" bun:",pk"`
	SessionID string      `json:"session_id"`
	Role      MessageRole `json:"role"`
	Content   string      `json:"content"`
	Timestamp string      `json:"timestamp"`
	ToolCalls []ToolCall  `json:"tool_calls,omitempty"`
}

type CreateMessageData struct {
	ID        string      `json:"id"`
	SessionID string      `json:"session_id"`
	Role      MessageRole `json:"role"`
	Content   string      `json:"content"`
	ToolCalls []ToolCall  `json:"tool_calls,omitempty"`
}

type MessageFilter struct {
	Role   *MessageRole `json:"role"`
	Limit  *int         `json:"limit"`
	Offset *int         `json:"offset"`
}
