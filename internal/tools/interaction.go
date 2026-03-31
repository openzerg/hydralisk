package tools

import (
	"context"
	"fmt"

	"github.com/openzerg/hydralisk/internal/core/interfaces"
	"github.com/openzerg/hydralisk/internal/core/types"
)

type QuestionTool struct{}

func (t *QuestionTool) Name() string { return "question" }
func (t *QuestionTool) Description() string {
	return "Ask the user a question and wait for a response."
}

func (t *QuestionTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"question": {Type: "string", Description: "The question to ask"},
			"header":   {Type: "string", Description: "Short header for the question"},
			"options": {
				Type:        "array",
				Items:       &types.JSONSchema{Type: "string"},
				Description: "List of options (optional)",
			},
		},
		Required: []string{"question"},
	}
}

func (t *QuestionTool) Execute(ctx context.Context, args map[string]interface{}, toolCtx *interfaces.ToolContext) (*types.ToolResult, error) {
	question, _ := args["question"].(string)
	header, _ := args["header"].(string)
	optionsRaw, _ := args["options"].([]interface{})

	if question == "" {
		return &types.ToolResult{Title: "Error", Output: "question is required", Metadata: map[string]interface{}{"error": true}}, nil
	}

	if header == "" {
		header = "Question"
	}

	options := make([]string, len(optionsRaw))
	for i, o := range optionsRaw {
		options[i] = o.(string)
	}

	// Emit question event
	if toolCtx.EventBus != nil {
		toolCtx.EventBus.Emit("question", map[string]interface{}{
			"session_id": toolCtx.SessionID,
			"question":   question,
			"header":     header,
			"options":    options,
		})
	}

	// For now, return a placeholder (in real implementation, this would wait for user input)
	return &types.ToolResult{
		Title:  header,
		Output: question,
		Metadata: map[string]interface{}{
			"question": question,
			"header":   header,
			"options":  options,
		},
	}, nil
}

type TodoWriteTool struct{}

func (t *TodoWriteTool) Name() string { return "todowrite" }
func (t *TodoWriteTool) Description() string {
	return "Write or update the todo list for the current session."
}

func (t *TodoWriteTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"todos": {
				Type: "array",
				Items: &types.JSONSchema{
					Type: "object",
					Properties: map[string]*types.JSONSchema{
						"content":  {Type: "string", Description: "Task content"},
						"status":   {Type: "string", Description: "pending, in_progress, completed"},
						"priority": {Type: "string", Description: "low, medium, high"},
					},
				},
				Description: "Array of todo items",
			},
		},
		Required: []string{"todos"},
	}
}

func (t *TodoWriteTool) Execute(ctx context.Context, args map[string]interface{}, toolCtx *interfaces.ToolContext) (*types.ToolResult, error) {
	todosRaw, _ := args["todos"].([]interface{})

	if len(todosRaw) == 0 {
		return &types.ToolResult{Title: "Error", Output: "todos array is required", Metadata: map[string]interface{}{"error": true}}, nil
	}

	// Emit todo update event
	if toolCtx.EventBus != nil {
		toolCtx.EventBus.Emit("todo_update", map[string]interface{}{
			"session_id": toolCtx.SessionID,
			"todos":      todosRaw,
		})
	}

	return &types.ToolResult{
		Title:    "Todos Updated",
		Output:   fmt.Sprintf("Updated %d todos", len(todosRaw)),
		Metadata: map[string]interface{}{"count": len(todosRaw)},
	}, nil
}

type TodoReadTool struct{}

func (t *TodoReadTool) Name() string { return "todoread" }
func (t *TodoReadTool) Description() string {
	return "Read the current todo list."
}

func (t *TodoReadTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type:       "object",
		Properties: map[string]*types.JSONSchema{},
		Required:   []string{},
	}
}

func (t *TodoReadTool) Execute(ctx context.Context, args map[string]interface{}, toolCtx *interfaces.ToolContext) (*types.ToolResult, error) {
	if toolCtx.Storage == nil {
		return &types.ToolResult{Title: "Error", Output: "Storage not available", Metadata: map[string]interface{}{"error": true}}, nil
	}

	todos, err := toolCtx.Storage.GetTodos(ctx, toolCtx.SessionID)
	if err != nil {
		return &types.ToolResult{Title: "Error", Output: fmt.Sprintf("Failed to get todos: %v", err), Metadata: map[string]interface{}{"error": true}}, nil
	}

	output := "Current Todos:\n"
	for _, todo := range todos {
		output += fmt.Sprintf("- [%s] %s (%s)\n", todo.Status, todo.Content, todo.Priority)
	}

	return &types.ToolResult{
		Title:    "Todo List",
		Output:   output,
		Metadata: map[string]interface{}{"count": len(todos)},
	}, nil
}
