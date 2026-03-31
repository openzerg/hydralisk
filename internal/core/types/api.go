package types

type ApiResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   *string     `json:"error"`
}

type PaginatedResponse struct {
	Items  []interface{} `json:"items"`
	Total  int           `json:"total"`
	Offset *int          `json:"offset,omitempty"`
	Limit  *int          `json:"limit,omitempty"`
}

type SessionListResponse struct {
	Sessions []*Session `json:"sessions"`
	Total    int        `json:"total"`
}

type MessageListResponse struct {
	Messages []*Message `json:"messages"`
	Total    int        `json:"total"`
}

type ProcessListResponse struct {
	Processes []*Process `json:"processes"`
	Total     int        `json:"total"`
}

type ToolListResponse struct {
	Tools []*ToolDefinition `json:"tools"`
}

type ChatRequest struct {
	Content string `json:"content"`
}

type ChatResponse struct {
	SessionID string `json:"session_id"`
}

type InterruptRequest struct {
	Message string `json:"message"`
}

type InterruptResponse struct {
	Interrupted bool `json:"interrupted"`
}

type TodoStatus string

const (
	TodoStatusPending    TodoStatus = "pending"
	TodoStatusInProgress TodoStatus = "in_progress"
	TodoStatusCompleted  TodoStatus = "completed"
)

type TodoPriority string

const (
	TodoPriorityLow    TodoPriority = "low"
	TodoPriorityMedium TodoPriority = "medium"
	TodoPriorityHigh   TodoPriority = "high"
)

type Todo struct {
	ID        string       `json:"id"`
	SessionID string       `json:"session_id"`
	Content   string       `json:"content"`
	Status    TodoStatus   `json:"status"`
	Priority  TodoPriority `json:"priority"`
	CreatedAt string       `json:"created_at"`
	UpdatedAt string       `json:"updated_at"`
}

type TodoListResponse struct {
	Todos []*Todo `json:"todos"`
}

type Activity struct {
	ID           string  `json:"id"`
	SessionID    *string `json:"session_id"`
	ActivityType string  `json:"activity_type"`
	Description  string  `json:"description"`
	Details      string  `json:"details"`
	Timestamp    string  `json:"timestamp"`
}

type ActivityListResponse struct {
	Activities []*Activity `json:"activities"`
}

type SessionContext struct {
	Session  *Session   `json:"session"`
	Messages []*Message `json:"messages"`
	Todos    []*Todo    `json:"todos"`
	Children []*Session `json:"children"`
}

type LimitsConfig struct {
	MaxTokensPerRequest   int `json:"max_tokens_per_request"`
	MaxRequestsPerMinute  int `json:"max_requests_per_minute"`
	MaxConcurrentSessions int `json:"max_concurrent_sessions"`
	DefaultTimeoutMs      int `json:"default_timeout_ms"`
}

type MessageRequest struct {
	Content   string  `json:"content"`
	SessionID *string `json:"session_id"`
}

type RemindRequest struct {
	Message   string  `json:"message"`
	SessionID *string `json:"session_id"`
}
