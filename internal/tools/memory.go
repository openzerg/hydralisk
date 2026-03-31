package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/openzerg/hydralisk/internal/core/interfaces"
	"github.com/openzerg/hydralisk/internal/core/types"
)

type MemorySearchTool struct{}

func (t *MemorySearchTool) Name() string { return "memory_search" }
func (t *MemorySearchTool) Description() string {
	return "Search for information in memory files using ripgrep."
}

func (t *MemorySearchTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"query":       {Type: "string", Description: "Search query"},
			"max_results": {Type: "number", Description: "Max results (default 10)"},
		},
		Required: []string{"query"},
	}
}

func (t *MemorySearchTool) Execute(ctx context.Context, args map[string]interface{}, toolCtx *interfaces.ToolContext) (*types.ToolResult, error) {
	query, _ := args["query"].(string)
	maxResults, _ := args["max_results"].(float64)

	if query == "" {
		return &types.ToolResult{Title: "Error", Output: "query is required", Metadata: map[string]interface{}{"error": true}}, nil
	}

	if maxResults == 0 {
		maxResults = 10
	}

	memoryDir := filepath.Join(toolCtx.WorkingDir, "memory")
	if _, err := os.Stat(memoryDir); os.IsNotExist(err) {
		return &types.ToolResult{Title: "No Results", Output: "No memory/ directory found", Metadata: map[string]interface{}{}}, nil
	}

	cmd := exec.Command("rg", "-i", "-n", "-C", "2", "--max-count", fmt.Sprintf("%d", int(maxResults)), query, memoryDir)
	output, err := cmd.Output()
	if err != nil {
		return &types.ToolResult{Title: "No Results", Output: fmt.Sprintf("No matches found for '%s'", query), Metadata: map[string]interface{}{}}, nil
	}

	return &types.ToolResult{
		Title:    "Found matches",
		Output:   string(output),
		Metadata: map[string]interface{}{"query": query},
	}, nil
}

type MemoryGetTool struct{}

func (t *MemoryGetTool) Name() string { return "memory_get" }
func (t *MemoryGetTool) Description() string {
	return "Read a snippet from a memory file."
}

func (t *MemoryGetTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"path":  {Type: "string", Description: "Memory file path"},
			"from":  {Type: "number", Description: "Start line (1-indexed)"},
			"lines": {Type: "number", Description: "Number of lines"},
		},
		Required: []string{"path"},
	}
}

func (t *MemoryGetTool) Execute(ctx context.Context, args map[string]interface{}, toolCtx *interfaces.ToolContext) (*types.ToolResult, error) {
	path, _ := args["path"].(string)
	from, _ := args["from"].(float64)
	lines, _ := args["lines"].(float64)

	if path == "" {
		return &types.ToolResult{Title: "Error", Output: "path is required", Metadata: map[string]interface{}{"error": true}}, nil
	}

	if from < 1 {
		from = 1
	}
	if lines == 0 {
		lines = 50
	}

	fullPath := filepath.Join(toolCtx.WorkingDir, path)
	file, err := os.Open(fullPath)
	if err != nil {
		return &types.ToolResult{Title: "Error", Output: fmt.Sprintf("Failed to read file: %v", err), Metadata: map[string]interface{}{"error": true}}, nil
	}
	defer file.Close()

	var resultLines []string
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if lineNum >= int(from) && lineNum < int(from)+int(lines) {
			resultLines = append(resultLines, scanner.Text())
		}
	}

	return &types.ToolResult{
		Title:     path,
		Output:    strings.Join(resultLines, "\n"),
		Truncated: lineNum >= int(from)+int(lines),
		Metadata:  map[string]interface{}{"path": path, "from": from, "lines": len(resultLines)},
	}, nil
}
