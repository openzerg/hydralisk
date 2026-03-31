package types

type SessionState string

const (
	SessionStateIdle      SessionState = "Idle"
	SessionStateRunning   SessionState = "Running"
	SessionStateDone      SessionState = "Done"
	SessionStateFailed    SessionState = "Failed"
	SessionStateCancelled SessionState = "Cancelled"
)

type Session struct {
	ID           string       `json:"id" bun:",pk"`
	Name         string       `json:"name"`
	Purpose      string       `json:"purpose"`
	State        SessionState `json:"state"`
	CreatedAt    string       `json:"created_at"`
	StartedAt    *string      `json:"started_at"`
	FinishedAt   *string      `json:"finished_at"`
	MessageCount int          `json:"message_count"`
	SystemPrompt string       `json:"system_prompt"`
	ParentID     *string      `json:"parent_id"`
	ChildIDs     []string     `json:"child_ids"`
	InputTokens  int          `json:"input_tokens"`
	OutputTokens int          `json:"output_tokens"`
}

type CreateSessionData struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Purpose      *string `json:"purpose"`
	SystemPrompt *string `json:"system_prompt"`
	ParentID     *string `json:"parent_id"`
}

type UpdateSessionData struct {
	Name         *string       `json:"name"`
	State        *SessionState `json:"state"`
	StartedAt    *string       `json:"started_at"`
	FinishedAt   *string       `json:"finished_at"`
	InputTokens  *int          `json:"input_tokens"`
	OutputTokens *int          `json:"output_tokens"`
}

type SessionFilter struct {
	State    *SessionState `json:"state"`
	Purpose  *string       `json:"purpose"`
	ParentID *string       `json:"parent_id"`
}
