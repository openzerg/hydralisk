package tools

import (
	"context"
	"fmt"
	"sync"

	"github.com/openzerg/hydralisk/internal/core/interfaces"
	"github.com/openzerg/hydralisk/internal/core/types"
)

type BatchTool struct{}

func (t *BatchTool) Name() string { return "batch" }
func (t *BatchTool) Description() string {
	return "Execute multiple tool calls in parallel. Use this when you have multiple independent operations that can be performed simultaneously. Maximum 25 tools per batch."
}

func (t *BatchTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"tool_calls": {
				Type:        "array",
				Description: "Array of tool calls to execute in parallel",
				Items: &types.JSONSchema{
					Type: "object",
					Properties: map[string]*types.JSONSchema{
						"tool":       {Type: "string", Description: "The name of the tool to execute"},
						"parameters": {Type: "object", Description: "Parameters for the tool"},
					},
				},
			},
		},
		Required: []string{"tool_calls"},
	}
}

func (t *BatchTool) Execute(ctx context.Context, args map[string]interface{}, toolCtx *interfaces.ToolContext) (*types.ToolResult, error) {
	toolCalls, ok := args["tool_calls"].([]interface{})
	if !ok || len(toolCalls) == 0 {
		return &types.ToolResult{Title: "Error", Output: "tool_calls array is required and must not be empty"}, nil
	}

	if toolCtx.ToolRegistry == nil {
		return &types.ToolResult{Title: "Error", Output: "Tool registry not available"}, nil
	}

	disallowed := map[string]bool{"batch": true}
	maxCalls := 25
	if len(toolCalls) > maxCalls {
		toolCalls = toolCalls[:maxCalls]
	}

	var wg sync.WaitGroup
	results := make([]map[string]interface{}, len(toolCalls))
	mu := sync.Mutex{}

	for i, call := range toolCalls {
		wg.Add(1)
		go func(idx int, c interface{}) {
			defer wg.Done()
			callMap, ok := c.(map[string]interface{})
			if !ok {
				mu.Lock()
				results[idx] = map[string]interface{}{"tool": "unknown", "success": false, "error": "invalid call format"}
				mu.Unlock()
				return
			}

			toolName, _ := callMap["tool"].(string)
			params, _ := callMap["parameters"].(map[string]interface{})

			result := map[string]interface{}{"tool": toolName}

			if disallowed[toolName] {
				result["success"] = false
				result["error"] = fmt.Sprintf("Tool '%s' is not allowed in batch", toolName)
			} else {
				r, err := toolCtx.ToolRegistry.Execute(ctx, toolName, params, toolCtx)
				if err != nil {
					result["success"] = false
					result["error"] = err.Error()
				} else {
					result["success"] = true
					if len(r.Output) > 200 {
						result["output"] = r.Output[:200]
					} else {
						result["output"] = r.Output
					}
				}
			}

			mu.Lock()
			results[idx] = result
			mu.Unlock()
		}(i, call)
	}
	wg.Wait()

	successful := 0
	for _, r := range results {
		if r["success"] == true {
			successful++
		}
	}

	output := fmt.Sprintf("Executed %d/%d tools successfully.", successful, len(results))

	return &types.ToolResult{
		Title:  fmt.Sprintf("Batch execution (%d/%d successful)", successful, len(results)),
		Output: output,
		Metadata: map[string]interface{}{
			"total":      len(results),
			"successful": successful,
			"failed":     len(results) - successful,
			"results":    results,
		},
	}, nil
}

type TaskTool struct{}

func (t *TaskTool) Name() string { return "task" }
func (t *TaskTool) Description() string {
	return "Launch a specialized sub-agent to handle a complex, multi-step task."
}

func (t *TaskTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"description": {Type: "string", Description: "A short (3-5 words) description of the task"},
			"prompt":      {Type: "string", Description: "The detailed task instructions for the sub-agent"},
			"task_id":     {Type: "string", Description: "Optional: Resume a previous task by passing its task_id"},
		},
		Required: []string{"description", "prompt"},
	}
}

func (t *TaskTool) Execute(ctx context.Context, args map[string]interface{}, toolCtx *interfaces.ToolContext) (*types.ToolResult, error) {
	description, _ := args["description"].(string)
	prompt, _ := args["prompt"].(string)

	if description == "" || prompt == "" {
		return &types.ToolResult{Title: "Error", Output: "description and prompt are required"}, nil
	}

	if toolCtx.SessionService == nil {
		return &types.ToolResult{Title: "Error", Output: "Session service not available"}, nil
	}

	return &types.ToolResult{
		Title:  description,
		Output: fmt.Sprintf("Task created: %s\n\nNote: Full sub-agent implementation requires LLM integration", description),
		Metadata: map[string]interface{}{
			"task_id":    "",
			"prompt":     prompt,
			"session_id": toolCtx.SessionID,
		},
	}, nil
}

type MessageTool struct{}

func (t *MessageTool) Name() string { return "message" }
func (t *MessageTool) Description() string {
	return "Send messages to WebUI or Proxy subscribers. Actions: send, reply, sendFile."
}

func (t *MessageTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"action":   {Type: "string", Description: "Message action: send, reply, sendFile"},
			"to":       {Type: "string", Description: "Target recipient"},
			"content":  {Type: "string", Description: "Message content"},
			"file":     {Type: "string", Description: "File path (for sendFile action)"},
			"reply_to": {Type: "string", Description: "Message ID to reply to"},
		},
		Required: []string{"action", "to"},
	}
}

func (t *MessageTool) Execute(ctx context.Context, args map[string]interface{}, toolCtx *interfaces.ToolContext) (*types.ToolResult, error) {
	action, _ := args["action"].(string)
	to, _ := args["to"].(string)
	content, _ := args["content"].(string)
	file, _ := args["file"].(string)
	replyTo, _ := args["reply_to"].(string)

	if action == "" || to == "" {
		return &types.ToolResult{Title: "Error", Output: "action and to are required"}, nil
	}

	if toolCtx.MessageBus == nil {
		return &types.ToolResult{Title: "Error", Output: "Message bus not available"}, nil
	}

	msg := map[string]interface{}{
		"action":     action,
		"to":         to,
		"content":    content,
		"file":       file,
		"reply_to":   replyTo,
		"session_id": toolCtx.SessionID,
	}

	toolCtx.EventBus.Emit("message_send", msg)

	return &types.ToolResult{
		Title:  "Message Sent",
		Output: fmt.Sprintf("Message '%s' sent to %s", action, to),
		Metadata: map[string]interface{}{
			"action": action,
			"to":     to,
		},
	}, nil
}
