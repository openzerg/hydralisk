package api

import (
	"context"
	"encoding/json"
	"net/http"

	"connectrpc.com/connect"
	"github.com/openzerg/hydralisk/internal/api/generated"
	"github.com/openzerg/hydralisk/internal/api/generated/generatedconnect"
	"github.com/openzerg/hydralisk/internal/service"
)

type AgentServiceHandler struct {
	services *service.ServiceLayer
}

func NewAgentServiceHandler(services *service.ServiceLayer) *AgentServiceHandler {
	return &AgentServiceHandler{services: services}
}

func (h *AgentServiceHandler) ListSessions(ctx context.Context, req *connect.Request[generated.ListSessionsRequest]) (*connect.Response[generated.SessionListResponse], error) {
	result, err := h.services.Session.List(ctx, 0, 0)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	sessions := make([]*generated.SessionInfo, len(result.Sessions))
	for i, s := range result.Sessions {
		sessions[i] = &generated.SessionInfo{
			Id:           s.ID,
			Purpose:      s.Purpose,
			State:        string(s.State),
			CreatedAt:    s.CreatedAt,
			MessageCount: int64(s.MessageCount),
		}
	}

	return connect.NewResponse(&generated.SessionListResponse{
		Sessions: sessions,
		Total:    int64(result.Total),
	}), nil
}

func (h *AgentServiceHandler) GetSession(ctx context.Context, req *connect.Request[generated.GetSessionRequest]) (*connect.Response[generated.SessionInfo], error) {
	result, err := h.services.Session.Get(ctx, req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if result == nil {
		return nil, connect.NewError(connect.CodeNotFound, nil)
	}

	return connect.NewResponse(&generated.SessionInfo{
		Id:           result.ID,
		Purpose:      result.Purpose,
		State:        string(result.State),
		CreatedAt:    result.CreatedAt,
		MessageCount: int64(result.MessageCount),
	}), nil
}

func (h *AgentServiceHandler) GetSessionMessages(ctx context.Context, req *connect.Request[generated.GetSessionMessagesRequest]) (*connect.Response[generated.MessageListResponse], error) {
	result, err := h.services.Session.GetMessages(ctx, req.Msg.SessionId, 0, 0)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	messages := make([]*generated.MessageInfo, len(result.Messages))
	for i, m := range result.Messages {
		toolCallsJSON := ""
		if m.ToolCalls != nil {
			data, _ := json.Marshal(m.ToolCalls)
			toolCallsJSON = string(data)
		}
		messages[i] = &generated.MessageInfo{
			Id:            m.ID,
			SessionId:     m.SessionID,
			Role:          string(m.Role),
			Content:       m.Content,
			Timestamp:     m.Timestamp,
			ToolCallsJson: toolCallsJSON,
		}
	}

	return connect.NewResponse(&generated.MessageListResponse{
		Messages: messages,
		Total:    int64(result.Total),
	}), nil
}

func (h *AgentServiceHandler) SendSessionChat(ctx context.Context, req *connect.Request[generated.SendSessionChatRequest]) (*connect.Response[generated.Empty], error) {
	err := h.services.Session.SendChat(ctx, req.Msg.SessionId, req.Msg.Content)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&generated.Empty{}), nil
}

func (h *AgentServiceHandler) InterruptSession(ctx context.Context, req *connect.Request[generated.InterruptSessionRequest]) (*connect.Response[generated.Empty], error) {
	err := h.services.Session.Interrupt(ctx, req.Msg.SessionId, req.Msg.Message)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&generated.Empty{}), nil
}

func (h *AgentServiceHandler) GetSessionContext(ctx context.Context, req *connect.Request[generated.GetSessionContextRequest]) (*connect.Response[generated.SessionContextResponse], error) {
	result, err := h.services.Session.GetContext(ctx, req.Msg.SessionId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if result == nil {
		return nil, connect.NewError(connect.CodeNotFound, nil)
	}

	data, _ := json.Marshal(result)
	return connect.NewResponse(&generated.SessionContextResponse{
		ContextJson: string(data),
	}), nil
}

func (h *AgentServiceHandler) ListProcesses(ctx context.Context, req *connect.Request[generated.ListProcessesRequest]) (*connect.Response[generated.ProcessListResponse], error) {
	result, err := h.services.Process.List(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	processes := make([]*generated.ProcessInfo, len(result.Processes))
	for i, p := range result.Processes {
		processes[i] = &generated.ProcessInfo{
			Id:        p.ID,
			Command:   p.Command,
			Status:    string(p.Status),
			StartedAt: p.StartedAt,
		}
	}

	return connect.NewResponse(&generated.ProcessListResponse{
		Processes: processes,
		Total:     int64(result.Total),
	}), nil
}

func (h *AgentServiceHandler) GetProcess(ctx context.Context, req *connect.Request[generated.GetProcessRequest]) (*connect.Response[generated.ProcessInfo], error) {
	result, err := h.services.Process.Get(ctx, req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if result == nil {
		return nil, connect.NewError(connect.CodeNotFound, nil)
	}

	return connect.NewResponse(&generated.ProcessInfo{
		Id:        result.ID,
		Command:   result.Command,
		Status:    string(result.Status),
		StartedAt: result.StartedAt,
	}), nil
}

func (h *AgentServiceHandler) GetProcessOutput(ctx context.Context, req *connect.Request[generated.GetProcessOutputRequest]) (*connect.Response[generated.ProcessOutputResponse], error) {
	stream := ""
	if req.Msg.Stream != nil {
		stream = *req.Msg.Stream
	}
	offset := int64(0)
	if req.Msg.Offset != nil {
		offset = *req.Msg.Offset
	}
	limit := int64(50)
	if req.Msg.Limit != nil {
		limit = *req.Msg.Limit
	}

	result, err := h.services.Process.GetOutput(req.Msg.ProcessId, stream, int(offset), int(limit))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	var content string
	for _, l := range result.Lines {
		content += l.Content + "\n"
	}

	return connect.NewResponse(&generated.ProcessOutputResponse{
		Content:   content,
		TotalSize: int64(result.TotalLines),
	}), nil
}

func (h *AgentServiceHandler) ListTasks(ctx context.Context, req *connect.Request[generated.ListTasksRequest]) (*connect.Response[generated.TaskListResponse], error) {
	return connect.NewResponse(&generated.TaskListResponse{
		Tasks: []*generated.TaskInfo{},
		Total: 0,
	}), nil
}

func (h *AgentServiceHandler) GetTask(ctx context.Context, req *connect.Request[generated.GetTaskRequest]) (*connect.Response[generated.TaskInfo], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *AgentServiceHandler) ListActivities(ctx context.Context, req *connect.Request[generated.ListActivitiesRequest]) (*connect.Response[generated.ActivityListResponse], error) {
	return connect.NewResponse(&generated.ActivityListResponse{
		Activities: []*generated.ActivityInfo{},
		Total:      0,
	}), nil
}

func (h *AgentServiceHandler) SendMessage(ctx context.Context, req *connect.Request[generated.SendMessageRequest]) (*connect.Response[generated.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *AgentServiceHandler) SendRemind(ctx context.Context, req *connect.Request[generated.SendRemindRequest]) (*connect.Response[generated.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *AgentServiceHandler) ListBuiltinTools(ctx context.Context, req *connect.Request[generated.Empty]) (*connect.Response[generated.BuiltinToolListResponse], error) {
	tools := h.services.Tool.List()

	toolInfos := make([]*generated.BuiltinToolInfo, len(tools))
	for i, t := range tools {
		paramsJSON, _ := json.Marshal(t.Function.Parameters)
		toolInfos[i] = &generated.BuiltinToolInfo{
			Name:           t.Function.Name,
			Description:    t.Function.Description,
			ParametersJson: string(paramsJSON),
		}
	}

	return connect.NewResponse(&generated.BuiltinToolListResponse{
		Tools: toolInfos,
	}), nil
}

func (h *AgentServiceHandler) ExecuteBuiltinTool(ctx context.Context, req *connect.Request[generated.ExecuteBuiltinToolRequest]) (*connect.Response[generated.ExecuteBuiltinToolResponse], error) {
	var args map[string]any
	if req.Msg.ArgsJson != "" {
		json.Unmarshal([]byte(req.Msg.ArgsJson), &args)
	}

	sessionID := ""
	if req.Msg.SessionId != nil {
		sessionID = *req.Msg.SessionId
	}

	result, err := h.services.Tool.Execute(ctx, req.Msg.ToolName, args, sessionID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	metadataJSON := ""
	if result.Metadata != nil {
		data, _ := json.Marshal(result.Metadata)
		metadataJSON = string(data)
	}

	attachmentsJSON := ""
	if result.Attachments != nil {
		data, _ := json.Marshal(result.Attachments)
		attachmentsJSON = string(data)
	}

	return connect.NewResponse(&generated.ExecuteBuiltinToolResponse{
		Title:           result.Title,
		Output:          result.Output,
		MetadataJson:    metadataJSON,
		AttachmentsJson: attachmentsJSON,
		Truncated:       result.Truncated,
	}), nil
}

func (h *AgentServiceHandler) CheckHealth(ctx context.Context, req *connect.Request[generated.Empty]) (*connect.Response[generated.HealthResponse], error) {
	return connect.NewResponse(&generated.HealthResponse{
		Healthy: true,
		Version: "0.1.0",
	}), nil
}

func (h *AgentServiceHandler) SubscribeSessionEvents(ctx context.Context, req *connect.Request[generated.SubscribeSessionEventsRequest], stream *connect.ServerStream[generated.SessionEvent]) error {
	<-ctx.Done()
	return nil
}

func (h *AgentServiceHandler) SubscribeGlobalEvents(ctx context.Context, req *connect.Request[generated.Empty], stream *connect.ServerStream[generated.GlobalEvent]) error {
	<-ctx.Done()
	return nil
}

var _ generatedconnect.AgentHandler = (*AgentServiceHandler)(nil)

func NewAgentHandler(services *service.ServiceLayer) (string, http.Handler) {
	return generatedconnect.NewAgentHandler(NewAgentServiceHandler(services))
}
