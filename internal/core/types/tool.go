package types

type JSONSchema struct {
	Type        string                 `json:"type"`
	Properties  map[string]*JSONSchema `json:"properties,omitempty"`
	Required    []string               `json:"required,omitempty"`
	Items       *JSONSchema            `json:"items,omitempty"`
	Description string                 `json:"description,omitempty"`
	Enum        []string               `json:"enum,omitempty"`
	Default     interface{}            `json:"default,omitempty"`
	AnyOf       []*JSONSchema          `json:"anyOf,omitempty"`
	OneOf       []*JSONSchema          `json:"oneOf,omitempty"`
}

type ToolFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  *JSONSchema `json:"parameters"`
}

type ToolDefinition struct {
	Type     string        `json:"type"`
	Function *ToolFunction `json:"function"`
}

type Attachment struct {
	Type     string  `json:"type"`
	Name     string  `json:"name"`
	Content  *string `json:"content,omitempty"`
	Path     *string `json:"path,omitempty"`
	URL      *string `json:"url,omitempty"`
	MimeType *string `json:"mime_type,omitempty"`
}

type ToolResult struct {
	Title       string                 `json:"title"`
	Output      string                 `json:"output"`
	Metadata    map[string]interface{} `json:"metadata"`
	Attachments []Attachment           `json:"attachments"`
	Truncated   bool                   `json:"truncated"`
}

type ToolExecuteRequest struct {
	ToolName  string                 `json:"tool_name"`
	Args      map[string]interface{} `json:"args"`
	SessionID *string                `json:"session_id"`
}
