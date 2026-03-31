package types

type ConnectedEvent struct {
	Message string `json:"message"`
}

type ThinkingEvent struct {
	SessionID   string `json:"session_id"`
	Content     string `json:"content"`
	SessionName string `json:"session_name"`
}

type ResponseEvent struct {
	SessionID   string `json:"session_id"`
	Content     string `json:"content"`
	SessionName string `json:"session_name"`
}

type ToolCallEvent struct {
	SessionID string `json:"session_id"`
	Tool      string `json:"tool"`
	Args      string `json:"args"`
	CallID    string `json:"call_id"`
}

type ToolResultEvent struct {
	SessionID string `json:"session_id"`
	Content   string `json:"content"`
	CallID    string `json:"call_id"`
	Success   bool   `json:"success"`
}

type DoneEvent struct {
	SessionID string `json:"session_id"`
}

type ErrorEvent struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

type ProcessNotificationEvent struct {
	ProcessID string `json:"process_id"`
	Event     struct {
		Type     string `json:"type"`
		Status   string `json:"status,omitempty"`
		ExitCode *int   `json:"exit_code,omitempty"`
	} `json:"event"`
	OutputPreview *string `json:"output_preview"`
}

type SessionActivityEvent struct {
	SessionID    string `json:"session_id"`
	SessionName  string `json:"session_name"`
	ActivityType string `json:"activity_type"`
	Message      string `json:"message"`
}

type InterruptedEvent struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

type TodoUpdateEvent struct {
	SessionID string `json:"session_id"`
	Todos     []struct {
		ID       string `json:"id"`
		Content  string `json:"content"`
		Status   string `json:"status"`
		Priority string `json:"priority"`
	} `json:"todos"`
}

type GlobalEvent struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}
