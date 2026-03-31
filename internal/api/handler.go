package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/openzerg/hydralisk/internal/service"
	"github.com/openzerg/hydralisk/packages/api/generated"
	"github.com/openzerg/hydralisk/packages/api/generated/generatedconnect"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type protojsonCodec struct {
	marshalOptions   protojson.MarshalOptions
	unmarshalOptions protojson.UnmarshalOptions
}

func newProtojsonCodec() *protojsonCodec {
	return &protojsonCodec{
		marshalOptions: protojson.MarshalOptions{
			EmitDefaultValues: true,
			UseProtoNames:     true,
		},
		unmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}
}

func (c *protojsonCodec) Name() string { return "json" }

func (c *protojsonCodec) Marshal(msg any) ([]byte, error) {
	protoMsg, ok := msg.(proto.Message)
	if !ok {
		return nil, connect.NewError(connect.CodeInternal, nil)
	}
	return c.marshalOptions.Marshal(protoMsg)
}

func (c *protojsonCodec) Unmarshal(data []byte, msg any) error {
	protoMsg, ok := msg.(proto.Message)
	if !ok {
		return connect.NewError(connect.CodeInternal, nil)
	}
	return c.unmarshalOptions.Unmarshal(data, protoMsg)
}

type AgentServiceHandler struct {
	generatedconnect.UnimplementedAgentHandler
	services      *service.ServiceLayer
	mu            sync.RWMutex
	registries    map[string]*generated.RegistryInfo
	providers     map[string]*generated.ProviderInfo
	sessionAgents map[string]string
}

func NewAgentServiceHandler(services *service.ServiceLayer) *AgentServiceHandler {
	return &AgentServiceHandler{
		services:      services,
		registries:    make(map[string]*generated.RegistryInfo),
		providers:     make(map[string]*generated.ProviderInfo),
		sessionAgents: make(map[string]string),
	}
}

func (h *AgentServiceHandler) ListSessions(ctx context.Context, req *connect.Request[generated.ListSessionsRequest]) (*connect.Response[generated.SessionListResponse], error) {
	result, err := h.services.Session.List(ctx, 0, 0)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	sessions := make([]*generated.SessionInfo, len(result.Sessions))
	for i, s := range result.Sessions {
		sessions[i] = &generated.SessionInfo{
			Id:                    s.ID,
			Purpose:               s.Purpose,
			State:                 string(s.State),
			CreatedAt:             s.CreatedAt,
			Agent:                 "build",
			InputTokens:           int32(s.InputTokens),
			OutputTokens:          int32(s.OutputTokens),
			HasCompactedHistory:   false,
			CompactedMessageCount: 0,
		}
	}

	return connect.NewResponse(&generated.SessionListResponse{
		Sessions: sessions,
		Total:    int32(result.Total),
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

	h.mu.RLock()
	agent := h.sessionAgents[req.Msg.Id]
	h.mu.RUnlock()
	if agent == "" {
		agent = "build"
	}

	return connect.NewResponse(&generated.SessionInfo{
		Id:                    result.ID,
		Purpose:               result.Purpose,
		State:                 string(result.State),
		CreatedAt:             result.CreatedAt,
		Agent:                 agent,
		InputTokens:           int32(result.InputTokens),
		OutputTokens:          int32(result.OutputTokens),
		HasCompactedHistory:   false,
		CompactedMessageCount: 0,
	}), nil
}

func (h *AgentServiceHandler) CreateSession(ctx context.Context, req *connect.Request[generated.CreateSessionRequest]) (*connect.Response[generated.SessionInfo], error) {
	name := ""
	result, err := h.services.Session.Create(ctx, name, req.Msg.Purpose, nil, nil)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&generated.SessionInfo{
		Id:                    result.ID,
		Purpose:               result.Purpose,
		State:                 string(result.State),
		CreatedAt:             result.CreatedAt,
		Agent:                 "build",
		InputTokens:           int32(result.InputTokens),
		OutputTokens:          int32(result.OutputTokens),
		HasCompactedHistory:   false,
		CompactedMessageCount: 0,
	}), nil
}

func (h *AgentServiceHandler) DeleteSession(ctx context.Context, req *connect.Request[generated.DeleteSessionRequest]) (*connect.Response[generated.Empty], error) {
	err := h.services.Session.Delete(ctx, req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&generated.Empty{}), nil
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
		Total:    int32(result.Total),
	}), nil
}

func (h *AgentServiceHandler) SendSessionChat(ctx context.Context, req *connect.Request[generated.SendSessionChatRequest]) (*connect.Response[generated.Empty], error) {
	if req.Msg.Content == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("content is required"))
	}
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

func (h *AgentServiceHandler) SwitchAgent(ctx context.Context, req *connect.Request[generated.SwitchAgentRequest]) (*connect.Response[generated.Empty], error) {
	if req.Msg.Agent != "plan" && req.Msg.Agent != "build" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("agent must be 'plan' or 'build'"))
	}

	h.mu.Lock()
	h.sessionAgents[req.Msg.SessionId] = req.Msg.Agent
	h.mu.Unlock()

	return connect.NewResponse(&generated.Empty{}), nil
}

func (h *AgentServiceHandler) UploadFile(ctx context.Context, req *connect.Request[generated.UploadFileRequest]) (*connect.Response[generated.UploadFileResponse], error) {
	if len(req.Msg.Content) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("content is required"))
	}
	return connect.NewResponse(&generated.UploadFileResponse{
		FilePath: "/tmp/" + req.Msg.Filename,
	}), nil
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

func (h *AgentServiceHandler) CompactSession(ctx context.Context, req *connect.Request[generated.CompactSessionRequest]) (*connect.Response[generated.CompactSessionResponse], error) {
	return connect.NewResponse(&generated.CompactSessionResponse{
		MessagesCompacted: 0,
	}), nil
}

func (h *AgentServiceHandler) GetHistoryMessages(ctx context.Context, req *connect.Request[generated.GetHistoryMessagesRequest]) (*connect.Response[generated.MessageListResponse], error) {
	return connect.NewResponse(&generated.MessageListResponse{
		Messages: []*generated.MessageInfo{},
		Total:    0,
	}), nil
}

func (h *AgentServiceHandler) DeleteMessagesFrom(ctx context.Context, req *connect.Request[generated.DeleteMessagesFromRequest]) (*connect.Response[generated.DeleteMessagesFromResponse], error) {
	return connect.NewResponse(&generated.DeleteMessagesFromResponse{
		DeletedCount: 0,
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
		Total:     int32(result.Total),
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
	offset := int32(0)
	if req.Msg.Offset != nil {
		offset = *req.Msg.Offset
	}
	limit := int32(50)
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
		TotalSize: int32(result.TotalLines),
	}), nil
}

func (h *AgentServiceHandler) KillProcess(ctx context.Context, req *connect.Request[generated.KillProcessRequest]) (*connect.Response[generated.Empty], error) {
	signal := "SIGTERM"
	if req.Msg.Signal != nil {
		signal = *req.Msg.Signal
	}
	err := h.services.Process.Kill(req.Msg.ProcessId, signal)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&generated.Empty{}), nil
}

func (h *AgentServiceHandler) ListTasks(ctx context.Context, req *connect.Request[generated.ListTasksRequest]) (*connect.Response[generated.TaskListResponse], error) {
	return connect.NewResponse(&generated.TaskListResponse{
		Tasks: []*generated.TaskInfo{},
		Total: 0,
	}), nil
}

func (h *AgentServiceHandler) GetTask(ctx context.Context, req *connect.Request[generated.GetTaskRequest]) (*connect.Response[generated.TaskInfo], error) {
	return nil, connect.NewError(connect.CodeNotFound, nil)
}

func (h *AgentServiceHandler) SendMessage(ctx context.Context, req *connect.Request[generated.SendMessageRequest]) (*connect.Response[generated.Empty], error) {
	return connect.NewResponse(&generated.Empty{}), nil
}

func (h *AgentServiceHandler) SendRemind(ctx context.Context, req *connect.Request[generated.SendRemindRequest]) (*connect.Response[generated.Empty], error) {
	return connect.NewResponse(&generated.Empty{}), nil
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

func (h *AgentServiceHandler) ExecuteTool(ctx context.Context, req *connect.Request[generated.ExecuteToolRequest]) (*connect.Response[generated.ExecuteToolResponse], error) {
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

	return connect.NewResponse(&generated.ExecuteToolResponse{
		Title:           result.Title,
		Output:          result.Output,
		MetadataJson:    metadataJSON,
		AttachmentsJson: attachmentsJSON,
		Truncated:       result.Truncated,
	}), nil
}

func (h *AgentServiceHandler) ListExternalTools(ctx context.Context, req *connect.Request[generated.Empty]) (*connect.Response[generated.ExternalToolListResponse], error) {
	return connect.NewResponse(&generated.ExternalToolListResponse{
		Tools: []*generated.ExternalToolInfo{},
	}), nil
}

func (h *AgentServiceHandler) RegisterExternalTool(ctx context.Context, req *connect.Request[generated.RegisterExternalToolRequest]) (*connect.Response[generated.ExternalToolInfo], error) {
	return connect.NewResponse(&generated.ExternalToolInfo{
		Name:        req.Msg.Name,
		Description: req.Msg.Description,
	}), nil
}

func (h *AgentServiceHandler) UnregisterExternalTool(ctx context.Context, req *connect.Request[generated.UnregisterExternalToolRequest]) (*connect.Response[generated.Empty], error) {
	return connect.NewResponse(&generated.Empty{}), nil
}

func (h *AgentServiceHandler) SyncExternalTools(ctx context.Context, req *connect.Request[generated.SyncExternalToolsRequest]) (*connect.Response[generated.ExternalToolListResponse], error) {
	return connect.NewResponse(&generated.ExternalToolListResponse{
		Tools: []*generated.ExternalToolInfo{},
	}), nil
}

func (h *AgentServiceHandler) SetToolVariable(ctx context.Context, req *connect.Request[generated.SetToolVariableRequest]) (*connect.Response[generated.Empty], error) {
	return connect.NewResponse(&generated.Empty{}), nil
}

func (h *AgentServiceHandler) ListProviders(ctx context.Context, req *connect.Request[generated.Empty]) (*connect.Response[generated.ProviderListResponse], error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	providers := make([]*generated.ProviderInfo, 0, len(h.providers))
	for _, p := range h.providers {
		providers = append(providers, p)
	}
	return connect.NewResponse(&generated.ProviderListResponse{
		Providers: providers,
	}), nil
}

func (h *AgentServiceHandler) RegisterProvider(ctx context.Context, req *connect.Request[generated.RegisterProviderRequest]) (*connect.Response[generated.ProviderInfo], error) {
	if req.Msg.Name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("name is required"))
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.providers[req.Msg.Name]; exists {
		return nil, connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("provider with name %s already exists", req.Msg.Name))
	}

	provider := &generated.ProviderInfo{
		Id:          fmt.Sprintf("provider-%d", time.Now().UnixNano()),
		Name:        req.Msg.Name,
		BaseUrl:     req.Msg.BaseUrl,
		ApiKey:      req.Msg.ApiKey,
		Model:       req.Msg.Model,
		MaxTokens:   req.Msg.MaxTokens,
		Temperature: req.Msg.Temperature,
		TopP:        req.Msg.TopP,
		TopK:        req.Msg.TopK,
		CreatedAt:   time.Now().Format(time.RFC3339),
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}
	h.providers[req.Msg.Name] = provider

	return connect.NewResponse(provider), nil
}

func (h *AgentServiceHandler) UpdateProvider(ctx context.Context, req *connect.Request[generated.UpdateProviderRequest]) (*connect.Response[generated.ProviderInfo], error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	provider, exists := h.providers[req.Msg.Name]
	if !exists {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("provider %s not found", req.Msg.Name))
	}

	if req.Msg.BaseUrl != nil {
		provider.BaseUrl = *req.Msg.BaseUrl
	}
	if req.Msg.Model != nil {
		provider.Model = *req.Msg.Model
	}
	if req.Msg.MaxTokens != nil {
		provider.MaxTokens = *req.Msg.MaxTokens
	}
	if req.Msg.Temperature != nil {
		provider.Temperature = *req.Msg.Temperature
	}
	if req.Msg.TopP != nil {
		provider.TopP = *req.Msg.TopP
	}
	if req.Msg.TopK != nil {
		provider.TopK = *req.Msg.TopK
	}
	provider.UpdatedAt = time.Now().Format(time.RFC3339)

	return connect.NewResponse(provider), nil
}

func (h *AgentServiceHandler) UnregisterProvider(ctx context.Context, req *connect.Request[generated.UnregisterExternalToolRequest]) (*connect.Response[generated.Empty], error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.providers, req.Msg.Name)
	return connect.NewResponse(&generated.Empty{}), nil
}

func (h *AgentServiceHandler) SetDefaultProvider(ctx context.Context, req *connect.Request[generated.SetDefaultProviderRequest]) (*connect.Response[generated.Empty], error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	provider, exists := h.providers[req.Msg.Name]
	if !exists {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("provider %s not found", req.Msg.Name))
	}

	for _, p := range h.providers {
		p.IsDefault = false
	}
	provider.IsDefault = true

	return connect.NewResponse(&generated.Empty{}), nil
}

func (h *AgentServiceHandler) TestProviderConnection(ctx context.Context, req *connect.Request[generated.TestProviderConnectionRequest]) (*connect.Response[generated.TestProviderConnectionResponse], error) {
	return connect.NewResponse(&generated.TestProviderConnectionResponse{
		Success: true,
	}), nil
}

func (h *AgentServiceHandler) SetSessionProvider(ctx context.Context, req *connect.Request[generated.SetSessionProviderRequest]) (*connect.Response[generated.Empty], error) {
	return connect.NewResponse(&generated.Empty{}), nil
}

func (h *AgentServiceHandler) ListRegistries(ctx context.Context, req *connect.Request[generated.Empty]) (*connect.Response[generated.RegistryListResponse], error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	registries := make([]*generated.RegistryInfo, 0, len(h.registries))
	for _, r := range h.registries {
		registries = append(registries, r)
	}
	return connect.NewResponse(&generated.RegistryListResponse{
		Registries: registries,
	}), nil
}

func (h *AgentServiceHandler) AddRegistry(ctx context.Context, req *connect.Request[generated.AddRegistryRequest]) (*connect.Response[generated.RegistryInfo], error) {
	if req.Msg.Name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("name is required"))
	}
	if req.Msg.Url == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("url is required"))
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.registries[req.Msg.Name]; exists {
		return nil, connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("registry with name %s already exists", req.Msg.Name))
	}

	registry := &generated.RegistryInfo{
		Id:        fmt.Sprintf("registry-%d", time.Now().UnixNano()),
		Name:      req.Msg.Name,
		Url:       req.Msg.Url,
		HasApiKey: req.Msg.ApiKey != nil && *req.Msg.ApiKey != "",
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	h.registries[req.Msg.Name] = registry

	return connect.NewResponse(registry), nil
}

func (h *AgentServiceHandler) RemoveRegistry(ctx context.Context, req *connect.Request[generated.RemoveRegistryRequest]) (*connect.Response[generated.Empty], error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for name, r := range h.registries {
		if r.Id == req.Msg.Id {
			delete(h.registries, name)
			break
		}
	}
	return connect.NewResponse(&generated.Empty{}), nil
}

func (h *AgentServiceHandler) ListInstalledSkills(ctx context.Context, req *connect.Request[generated.Empty]) (*connect.Response[generated.SkillListResponse], error) {
	return connect.NewResponse(&generated.SkillListResponse{
		Skills: []*generated.SkillInfo{},
	}), nil
}

func (h *AgentServiceHandler) ListRemoteSkills(ctx context.Context, req *connect.Request[generated.ListRemoteSkillsRequest]) (*connect.Response[generated.SkillListResponse], error) {
	return connect.NewResponse(&generated.SkillListResponse{
		Skills: []*generated.SkillInfo{},
	}), nil
}

func (h *AgentServiceHandler) InstallSkill(ctx context.Context, req *connect.Request[generated.InstallSkillRequest]) (*connect.Response[generated.SkillInfo], error) {
	return connect.NewResponse(&generated.SkillInfo{
		Name: req.Msg.Name,
	}), nil
}

func (h *AgentServiceHandler) UninstallSkill(ctx context.Context, req *connect.Request[generated.UninstallSkillRequest]) (*connect.Response[generated.Empty], error) {
	return connect.NewResponse(&generated.Empty{}), nil
}

func (h *AgentServiceHandler) GetSkill(ctx context.Context, req *connect.Request[generated.GetSkillRequest]) (*connect.Response[generated.SkillInfo], error) {
	return nil, connect.NewError(connect.CodeNotFound, nil)
}

func (h *AgentServiceHandler) ListTimers(ctx context.Context, req *connect.Request[generated.ListTimersRequest]) (*connect.Response[generated.TimerListResponse], error) {
	return connect.NewResponse(&generated.TimerListResponse{
		Timers: []*generated.TimerInfo{},
	}), nil
}

func (h *AgentServiceHandler) CancelTimer(ctx context.Context, req *connect.Request[generated.CancelTimerRequest]) (*connect.Response[generated.Empty], error) {
	return connect.NewResponse(&generated.Empty{}), nil
}

func (h *AgentServiceHandler) AnswerQuestion(ctx context.Context, req *connect.Request[generated.AnswerQuestionRequest]) (*connect.Response[generated.Empty], error) {
	return connect.NewResponse(&generated.Empty{}), nil
}

func (h *AgentServiceHandler) CheckHealth(ctx context.Context, req *connect.Request[generated.Empty]) (*connect.Response[generated.HealthResponse], error) {
	return connect.NewResponse(&generated.HealthResponse{
		Healthy: true,
		Version: "1.0.0",
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

func NewAgentHandler(services *service.ServiceLayer) (string, http.Handler) {
	return generatedconnect.NewAgentHandler(
		NewAgentServiceHandler(services),
		connect.WithCodec(newProtojsonCodec()),
	)
}
