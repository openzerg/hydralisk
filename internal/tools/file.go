package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/openzerg/hydralisk/internal/core/interfaces"
	"github.com/openzerg/hydralisk/internal/core/types"
)

type ReadTool struct{}

func (t *ReadTool) Name() string { return "read" }
func (t *ReadTool) Description() string {
	return "Read a file from the local filesystem."
}

func (t *ReadTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"file_path": {Type: "string", Description: "The absolute path to the file to read"},
			"offset":    {Type: "number", Description: "Line number to start reading from (1-indexed)"},
			"limit":     {Type: "number", Description: "Number of lines to read"},
		},
		Required: []string{"file_path"},
	}
}

func (t *ReadTool) Execute(ctx context.Context, args map[string]interface{}, toolCtx *interfaces.ToolContext) (*types.ToolResult, error) {
	filePath, _ := args["file_path"].(string)
	offset, _ := args["offset"].(float64)
	limit, _ := args["limit"].(float64)

	if filePath == "" {
		return &types.ToolResult{Title: "Error", Output: "file_path is required", Metadata: map[string]interface{}{"error": true}}, nil
	}

	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(toolCtx.WorkingDir, filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return &types.ToolResult{Title: "Error", Output: fmt.Sprintf("Failed to read file: %v", err), Metadata: map[string]interface{}{"error": true}}, nil
	}
	defer file.Close()

	startLine := int(offset)
	if startLine < 1 {
		startLine = 1
	}
	maxLines := int(limit)
	if maxLines == 0 {
		maxLines = 2000
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if lineNum >= startLine && lineNum < startLine+maxLines {
			lines = append(lines, fmt.Sprintf("%d: %s", lineNum, scanner.Text()))
		}
	}

	truncated := lineNum >= startLine+maxLines
	return &types.ToolResult{
		Title:     filepath.Base(filePath),
		Output:    strings.Join(lines, "\n"),
		Truncated: truncated,
		Metadata:  map[string]interface{}{"path": filePath, "lines": len(lines)},
	}, nil
}

type WriteTool struct{}

func (t *WriteTool) Name() string { return "write" }
func (t *WriteTool) Description() string {
	return "Write a file to the local filesystem."
}

func (t *WriteTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"file_path": {Type: "string", Description: "The absolute path to the file to write"},
			"content":   {Type: "string", Description: "The content to write to the file"},
		},
		Required: []string{"file_path", "content"},
	}
}

func (t *WriteTool) Execute(ctx context.Context, args map[string]interface{}, toolCtx *interfaces.ToolContext) (*types.ToolResult, error) {
	filePath, _ := args["file_path"].(string)
	content, _ := args["content"].(string)

	if filePath == "" {
		return &types.ToolResult{Title: "Error", Output: "file_path is required", Metadata: map[string]interface{}{"error": true}}, nil
	}

	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(toolCtx.WorkingDir, filePath)
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &types.ToolResult{Title: "Error", Output: fmt.Sprintf("Failed to create directory: %v", err), Metadata: map[string]interface{}{"error": true}}, nil
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return &types.ToolResult{Title: "Error", Output: fmt.Sprintf("Failed to write file: %v", err), Metadata: map[string]interface{}{"error": true}}, nil
	}

	return &types.ToolResult{
		Title:    "File written",
		Output:   fmt.Sprintf("Successfully wrote to %s", filePath),
		Metadata: map[string]interface{}{"path": filePath},
	}, nil
}

type EditTool struct{}

func (t *EditTool) Name() string { return "edit" }
func (t *EditTool) Description() string {
	return "Perform exact string replacements in a file."
}

func (t *EditTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"file_path":   {Type: "string", Description: "The absolute path to the file to edit"},
			"old_string":  {Type: "string", Description: "The text to replace"},
			"new_string":  {Type: "string", Description: "The text to replace it with"},
			"replace_all": {Type: "boolean", Description: "Replace all occurrences (default false)"},
		},
		Required: []string{"file_path", "old_string", "new_string"},
	}
}

func (t *EditTool) Execute(ctx context.Context, args map[string]interface{}, toolCtx *interfaces.ToolContext) (*types.ToolResult, error) {
	filePath, _ := args["file_path"].(string)
	oldString, _ := args["old_string"].(string)
	newString, _ := args["new_string"].(string)
	replaceAll, _ := args["replace_all"].(bool)

	if filePath == "" || oldString == "" {
		return &types.ToolResult{Title: "Error", Output: "file_path and old_string are required", Metadata: map[string]interface{}{"error": true}}, nil
	}

	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(toolCtx.WorkingDir, filePath)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return &types.ToolResult{Title: "Error", Output: fmt.Sprintf("Failed to read file: %v", err), Metadata: map[string]interface{}{"error": true}}, nil
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, oldString) {
		return &types.ToolResult{Title: "Error", Output: "old_string not found in file", Metadata: map[string]interface{}{"error": true}}, nil
	}

	var newContent string
	if replaceAll {
		newContent = strings.ReplaceAll(contentStr, oldString, newString)
	} else {
		count := strings.Count(contentStr, oldString)
		if count > 1 {
			return &types.ToolResult{Title: "Error", Output: fmt.Sprintf("old_string appears %d times. Use replace_all or provide more context.", count), Metadata: map[string]interface{}{"error": true}}, nil
		}
		newContent = strings.Replace(contentStr, oldString, newString, 1)
	}

	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return &types.ToolResult{Title: "Error", Output: fmt.Sprintf("Failed to write file: %v", err), Metadata: map[string]interface{}{"error": true}}, nil
	}

	return &types.ToolResult{
		Title:    "File edited",
		Output:   fmt.Sprintf("Successfully edited %s", filePath),
		Metadata: map[string]interface{}{"path": filePath},
	}, nil
}
