package tools

import (
	"context"
	"fmt"

	"github.com/openzerg/hydralisk/internal/core/interfaces"
	"github.com/openzerg/hydralisk/internal/core/types"
)

type JobTool struct{}

func (t *JobTool) Name() string { return "job" }
func (t *JobTool) Description() string {
	return `Manage background jobs: run commands, list jobs, get output, kill jobs.

Actions:
- run: Start a new background job
- list: List all jobs
- output: Get job output
- kill: Terminate a running job
- status: Get job status
- wait: Wait for job to complete`
}

func (t *JobTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"action":  {Type: "string", Description: "Job action (run, list, output, kill, status, wait)"},
			"command": {Type: "string", Description: "Command to run (for run action)"},
			"workdir": {Type: "string", Description: "Working directory"},
			"job_id":  {Type: "string", Description: "Job ID"},
			"stream":  {Type: "string", Description: "Output stream (stdout, stderr, both)"},
			"offset":  {Type: "number", Description: "Output offset"},
			"limit":   {Type: "number", Description: "Output limit"},
			"signal":  {Type: "string", Description: "Signal to send (for kill)"},
			"timeout": {Type: "number", Description: "Timeout in ms"},
		},
		Required: []string{"action"},
	}
}

func (t *JobTool) Execute(ctx context.Context, args map[string]interface{}, toolCtx *interfaces.ToolContext) (*types.ToolResult, error) {
	action, _ := args["action"].(string)
	pm := toolCtx.ProcessManager

	if pm == nil {
		return &types.ToolResult{Title: "Error", Output: "Process manager not available", Metadata: map[string]interface{}{"error": true}}, nil
	}

	switch action {
	case "run":
		return t.runJob(ctx, args, toolCtx, pm)
	case "list":
		return t.listJobs(pm)
	case "output":
		return t.getOutput(args, pm)
	case "kill":
		return t.killJob(args, pm)
	case "status":
		return t.getStatus(args, pm)
	case "wait":
		return t.waitJob(args, pm)
	default:
		return &types.ToolResult{Title: "Error", Output: "Unknown action: " + action, Metadata: map[string]interface{}{"error": true}}, nil
	}
}

func (t *JobTool) runJob(ctx context.Context, args map[string]interface{}, toolCtx *interfaces.ToolContext, pm interfaces.IProcessManager) (*types.ToolResult, error) {
	command, _ := args["command"].(string)
	workdir, _ := args["workdir"].(string)
	timeout, _ := args["timeout"].(float64)

	if command == "" {
		return &types.ToolResult{Title: "Error", Output: "command is required", Metadata: map[string]interface{}{"error": true}}, nil
	}

	if workdir == "" {
		workdir = toolCtx.WorkingDir
	}

	opts := types.SpawnOptions{
		Workdir: workdir,
		Timeout: int(timeout),
	}

	handle, err := pm.Spawn(ctx, command, opts)
	if err != nil {
		return &types.ToolResult{Title: "Error", Output: fmt.Sprintf("Failed to start job: %v", err), Metadata: map[string]interface{}{"error": true}}, nil
	}

	return &types.ToolResult{
		Title:    "Job Started",
		Output:   fmt.Sprintf("Job %s started in background", handle.ID),
		Metadata: map[string]interface{}{"job_id": handle.ID},
	}, nil
}

func (t *JobTool) listJobs(pm interfaces.IProcessManager) (*types.ToolResult, error) {
	handles := pm.ListProcesses()
	if len(handles) == 0 {
		return &types.ToolResult{Title: "No Jobs", Output: "No running jobs", Metadata: map[string]interface{}{}}, nil
	}

	output := "Running Jobs:\n"
	for _, h := range handles {
		status := pm.GetStatus(h.ID)
		output += fmt.Sprintf("  %s [%s] started %v\n", h.ID, status, h.StartedAt)
	}

	return &types.ToolResult{Title: "Job List", Output: output, Metadata: map[string]interface{}{}}, nil
}

func (t *JobTool) getOutput(args map[string]interface{}, pm interfaces.IProcessManager) (*types.ToolResult, error) {
	jobID, _ := args["job_id"].(string)
	stream, _ := args["stream"].(string)
	offset, _ := args["offset"].(float64)
	limit, _ := args["limit"].(float64)

	if jobID == "" {
		return &types.ToolResult{Title: "Error", Output: "job_id is required", Metadata: map[string]interface{}{"error": true}}, nil
	}

	if stream == "" {
		stream = "stdout"
	}

	output, err := pm.GetOutput(jobID, stream, int(offset), int(limit))
	if err != nil {
		return &types.ToolResult{Title: "Error", Output: fmt.Sprintf("Failed to get output: %v", err), Metadata: map[string]interface{}{"error": true}}, nil
	}

	var lines string
	for _, l := range output.Lines {
		lines += l.Content + "\n"
	}

	return &types.ToolResult{
		Title:     "Job Output",
		Output:    lines,
		Truncated: output.HasMore,
		Metadata:  map[string]interface{}{"job_id": jobID, "total_lines": output.TotalLines},
	}, nil
}

func (t *JobTool) killJob(args map[string]interface{}, pm interfaces.IProcessManager) (*types.ToolResult, error) {
	jobID, _ := args["job_id"].(string)
	signal, _ := args["signal"].(string)

	if jobID == "" {
		return &types.ToolResult{Title: "Error", Output: "job_id is required", Metadata: map[string]interface{}{"error": true}}, nil
	}

	if signal == "" {
		signal = "SIGTERM"
	}

	if err := pm.Kill(jobID, signal); err != nil {
		return &types.ToolResult{Title: "Error", Output: fmt.Sprintf("Failed to kill job: %v", err), Metadata: map[string]interface{}{"error": true}}, nil
	}

	return &types.ToolResult{
		Title:    "Job Killed",
		Output:   fmt.Sprintf("Job %s terminated with %s", jobID, signal),
		Metadata: map[string]interface{}{"job_id": jobID},
	}, nil
}

func (t *JobTool) getStatus(args map[string]interface{}, pm interfaces.IProcessManager) (*types.ToolResult, error) {
	jobID, _ := args["job_id"].(string)

	if jobID == "" {
		return &types.ToolResult{Title: "Error", Output: "job_id is required", Metadata: map[string]interface{}{"error": true}}, nil
	}

	status := pm.GetStatus(jobID)
	stats, _ := pm.GetOutputStats(jobID)

	return &types.ToolResult{
		Title:    "Job Status",
		Output:   fmt.Sprintf("Status: %s\nStdout: %d bytes\nStderr: %d bytes", status, stats.StdoutSize, stats.StderrSize),
		Metadata: map[string]interface{}{"job_id": jobID, "status": status},
	}, nil
}

func (t *JobTool) waitJob(args map[string]interface{}, pm interfaces.IProcessManager) (*types.ToolResult, error) {
	jobID, _ := args["job_id"].(string)
	timeout, _ := args["timeout"].(float64)

	if jobID == "" {
		return &types.ToolResult{Title: "Error", Output: "job_id is required", Metadata: map[string]interface{}{"error": true}}, nil
	}

	result, err := pm.Wait(jobID, int(timeout))
	if err != nil {
		return &types.ToolResult{Title: "Error", Output: fmt.Sprintf("Failed to wait: %v", err), Metadata: map[string]interface{}{"error": true}}, nil
	}

	return &types.ToolResult{
		Title:    "Job Completed",
		Output:   fmt.Sprintf("Exit code: %d\nStatus: %s\nDuration: %dms", result.ExitCode, result.Status, result.DurationMs),
		Metadata: map[string]interface{}{"job_id": jobID, "exit_code": result.ExitCode},
	}, nil
}
