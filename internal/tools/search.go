package tools

import (
	"context"
	"os/exec"

	"github.com/openzerg/hydralisk/internal/core/interfaces"
	"github.com/openzerg/hydralisk/internal/core/types"
)

type GlobTool struct{}

func (t *GlobTool) Name() string { return "glob" }
func (t *GlobTool) Description() string {
	return "Fast file pattern matching using ripgrep."
}

func (t *GlobTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"pattern": {Type: "string", Description: "Glob pattern to match files"},
			"path":    {Type: "string", Description: "Directory to search (default: current directory)"},
		},
		Required: []string{"pattern"},
	}
}

func (t *GlobTool) Execute(ctx context.Context, args map[string]interface{}, toolCtx *interfaces.ToolContext) (*types.ToolResult, error) {
	pattern, _ := args["pattern"].(string)
	path, _ := args["path"].(string)

	if pattern == "" {
		return &types.ToolResult{Title: "Error", Output: "pattern is required", Metadata: map[string]interface{}{"error": true}}, nil
	}

	if path == "" {
		path = toolCtx.WorkingDir
	}

	cmd := exec.Command("rg", "--files", "--glob", pattern, path)
	output, err := cmd.Output()
	if err != nil {
		return &types.ToolResult{Title: "Error", Output: string(output), Metadata: map[string]interface{}{"error": true}}, nil
	}

	return &types.ToolResult{
		Title:    "Files found",
		Output:   string(output),
		Metadata: map[string]interface{}{"pattern": pattern, "path": path},
	}, nil
}

type GrepTool struct{}

func (t *GrepTool) Name() string { return "grep" }
func (t *GrepTool) Description() string {
	return "Fast content search using ripgrep."
}

func (t *GrepTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"pattern":     {Type: "string", Description: "Regular expression pattern to search for"},
			"path":        {Type: "string", Description: "File or directory to search"},
			"ignore_case": {Type: "boolean", Description: "Case insensitive search (default false)"},
		},
		Required: []string{"pattern"},
	}
}

func (t *GrepTool) Execute(ctx context.Context, args map[string]interface{}, toolCtx *interfaces.ToolContext) (*types.ToolResult, error) {
	pattern, _ := args["pattern"].(string)
	path, _ := args["path"].(string)
	ignoreCase, _ := args["ignore_case"].(bool)

	if pattern == "" {
		return &types.ToolResult{Title: "Error", Output: "pattern is required", Metadata: map[string]interface{}{"error": true}}, nil
	}

	if path == "" {
		path = toolCtx.WorkingDir
	}

	args_list := []string{"-n", "--no-heading"}
	if ignoreCase {
		args_list = append(args_list, "-i")
	}
	args_list = append(args_list, pattern, path)

	cmd := exec.Command("rg", args_list...)
	output, err := cmd.Output()
	if err != nil {
		return &types.ToolResult{Title: "No matches", Output: "No matches found", Metadata: map[string]interface{}{}}, nil
	}

	return &types.ToolResult{
		Title:    "Matches found",
		Output:   string(output),
		Metadata: map[string]interface{}{"pattern": pattern, "path": path},
	}, nil
}

type LsTool struct{}

func (t *LsTool) Name() string { return "ls" }
func (t *LsTool) Description() string {
	return "List directory contents."
}

func (t *LsTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"path": {Type: "string", Description: "Directory path to list"},
		},
		Required: []string{},
	}
}

func (t *LsTool) Execute(ctx context.Context, args map[string]interface{}, toolCtx *interfaces.ToolContext) (*types.ToolResult, error) {
	path, _ := args["path"].(string)

	if path == "" {
		path = toolCtx.WorkingDir
	}

	cmd := exec.Command("rg", "--files", path)
	output, err := cmd.Output()
	if err != nil {
		return &types.ToolResult{Title: "Error", Output: string(output), Metadata: map[string]interface{}{"error": true}}, nil
	}

	return &types.ToolResult{
		Title:    "Directory listing",
		Output:   string(output),
		Metadata: map[string]interface{}{"path": path},
	}, nil
}
