package tools

import (
	"context"
	"sync"

	"github.com/openzerg/hydralisk/internal/core/interfaces"
	"github.com/openzerg/hydralisk/internal/core/types"
)

type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]interfaces.ITool
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]interfaces.ITool),
	}
}

func (r *ToolRegistry) Register(tool interfaces.ITool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[tool.Name()] = tool
}

func (r *ToolRegistry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tools, name)
}

func (r *ToolRegistry) Get(name string) interfaces.ITool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.tools[name]
}

func (r *ToolRegistry) GetDefinitions() []*types.ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	defs := make([]*types.ToolDefinition, 0, len(r.tools))
	for _, tool := range r.tools {
		defs = append(defs, &types.ToolDefinition{
			Type: "function",
			Function: &types.ToolFunction{
				Name:        tool.Name(),
				Description: tool.Description(),
				Parameters:  tool.Parameters(),
			},
		})
	}
	return defs
}

func (r *ToolRegistry) Execute(ctx context.Context, name string, args map[string]interface{}, toolCtx *interfaces.ToolContext) (*types.ToolResult, error) {
	r.mu.RLock()
	tool, ok := r.tools[name]
	r.mu.RUnlock()

	if !ok {
		return &types.ToolResult{
			Title:    "Error",
			Output:   "Tool not found: " + name,
			Metadata: map[string]interface{}{"error": true},
		}, nil
	}

	return tool.Execute(ctx, args, toolCtx)
}

func (r *ToolRegistry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.tools[name]
	return ok
}

func (r *ToolRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}
