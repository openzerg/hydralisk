package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/openzerg/hydralisk/internal/core/interfaces"
	"github.com/openzerg/hydralisk/internal/core/types"
	"github.com/openzerg/hydralisk/internal/db"
	"github.com/openzerg/hydralisk/internal/event-bus"
	"github.com/openzerg/hydralisk/internal/llm-client"
	"github.com/openzerg/hydralisk/internal/process-manager"
	"github.com/openzerg/hydralisk/internal/tools"
)

type SessionService struct {
	storage       *db.Repository
	eventBus      *eventbus.EventBus
	llmClient     *llmclient.Client
	toolRegistry  *tools.ToolRegistry
	messageBus    *MessageBus
	sessions      map[string]*types.Session
	states        map[string]types.SessionState
	activeSession string
	mu            sync.RWMutex
}

func NewSessionService(
	storage *db.Repository,
	eventBus *eventbus.EventBus,
	llmClient *llmclient.Client,
	toolRegistry *tools.ToolRegistry,
	messageBus *MessageBus,
) *SessionService {
	return &SessionService{
		storage:      storage,
		eventBus:     eventBus,
		llmClient:    llmClient,
		toolRegistry: toolRegistry,
		messageBus:   messageBus,
		sessions:     make(map[string]*types.Session),
		states:       make(map[string]types.SessionState),
	}
}

func (s *SessionService) List(ctx context.Context, offset, limit int) (*types.SessionListResponse, error) {
	sessions, err := s.storage.ListSessions(ctx, nil)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	for _, session := range sessions {
		s.sessions[session.ID] = session
		if _, ok := s.states[session.ID]; !ok {
			s.states[session.ID] = session.State
		}
	}
	s.mu.Unlock()

	return &types.SessionListResponse{Sessions: sessions, Total: len(sessions)}, nil
}

func (s *SessionService) Get(ctx context.Context, id string) (*types.Session, error) {
	s.mu.RLock()
	if session, ok := s.sessions[id]; ok {
		s.mu.RUnlock()
		return session, nil
	}
	s.mu.RUnlock()

	session, err := s.storage.GetSession(ctx, id)
	if err != nil {
		return nil, err
	}

	if session != nil {
		s.mu.Lock()
		s.sessions[id] = session
		if _, ok := s.states[id]; !ok {
			s.states[id] = session.State
		}
		s.mu.Unlock()
	}

	return session, nil
}

func (s *SessionService) Create(ctx context.Context, name, purpose string, systemPrompt *string, parentID *string) (*types.Session, error) {
	id := uuid.New().String()
	if name == "" {
		name = fmt.Sprintf("session-%s", id[:8])
	}
	if purpose == "" {
		purpose = "Main"
	}

	data := &types.CreateSessionData{
		ID:           id,
		Name:         name,
		Purpose:      &purpose,
		SystemPrompt: systemPrompt,
		ParentID:     parentID,
	}

	session, err := s.storage.CreateSession(ctx, data)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.sessions[id] = session
	s.states[id] = types.SessionStateIdle
	s.mu.Unlock()

	return session, nil
}

func (s *SessionService) Delete(ctx context.Context, id string) error {
	if err := s.storage.DeleteSession(ctx, id); err != nil {
		return err
	}

	s.mu.Lock()
	delete(s.sessions, id)
	delete(s.states, id)
	if s.activeSession == id {
		s.activeSession = ""
	}
	s.mu.Unlock()

	return nil
}

func (s *SessionService) GetMessages(ctx context.Context, sessionID string, offset, limit int) (*types.MessageListResponse, error) {
	messages, err := s.storage.GetMessages(ctx, sessionID, nil)
	if err != nil {
		return nil, err
	}
	return &types.MessageListResponse{Messages: messages, Total: len(messages)}, nil
}

func (s *SessionService) AddMessage(ctx context.Context, sessionID string, role types.MessageRole, content string) (*types.Message, error) {
	data := &types.CreateMessageData{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      role,
		Content:   content,
	}

	msg, err := s.storage.SaveMessage(ctx, data)
	if err != nil {
		return nil, err
	}

	_ = s.storage.IncrementMessageCount(ctx, sessionID)

	return msg, nil
}

func (s *SessionService) GetTodos(ctx context.Context, sessionID string) ([]*types.Todo, error) {
	return s.storage.GetTodos(ctx, sessionID)
}

func (s *SessionService) GetActivities(ctx context.Context, sessionID string) ([]*types.Activity, error) {
	return s.storage.GetActivities(ctx, sessionID)
}

func (s *SessionService) GetContext(ctx context.Context, sessionID string) (*types.SessionContext, error) {
	session, err := s.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, fmt.Errorf("session not found")
	}

	messages, err := s.storage.GetMessages(ctx, sessionID, nil)
	if err != nil {
		return nil, err
	}

	todos, err := s.storage.GetTodos(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	var children []*types.Session
	if session.ParentID != nil {
		children, _ = s.storage.ListSessions(ctx, &types.SessionFilter{ParentID: session.ParentID})
	}

	return &types.SessionContext{
		Session:  session,
		Messages: messages,
		Todos:    todos,
		Children: children,
	}, nil
}

func (s *SessionService) SetActive(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeSession = sessionID
}

func (s *SessionService) GetActive() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.activeSession
}

func (s *SessionService) SendChat(ctx context.Context, sessionID, content string) error {
	_, err := s.AddMessage(ctx, sessionID, types.MessageRoleUser, content)
	if err != nil {
		return err
	}
	s.eventBus.Emit("chat", map[string]any{"session_id": sessionID, "content": content})
	return nil
}

func (s *SessionService) Interrupt(ctx context.Context, sessionID, message string) error {
	s.eventBus.Emit("interrupt", map[string]any{"session_id": sessionID, "message": message})
	return nil
}

type ProcessService struct {
	storage *db.Repository
	procMgr *processmanager.ProcessManager
}

func NewProcessService(storage *db.Repository, procMgr *processmanager.ProcessManager) *ProcessService {
	return &ProcessService{storage: storage, procMgr: procMgr}
}

func (p *ProcessService) List(ctx context.Context) (*types.ProcessListResponse, error) {
	processes, err := p.storage.ListProcesses(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &types.ProcessListResponse{Processes: processes, Total: len(processes)}, nil
}

func (p *ProcessService) Get(ctx context.Context, id string) (*types.Process, error) {
	return p.storage.GetProcess(ctx, id)
}

func (p *ProcessService) GetOutput(processID string, stream string, offset, limit int) (*types.ProcessOutput, error) {
	return p.procMgr.GetOutput(processID, stream, offset, limit)
}

func (p *ProcessService) Kill(processID, signal string) error {
	if signal == "" {
		signal = "SIGTERM"
	}
	return p.procMgr.Kill(processID, signal)
}

type ToolService struct {
	registry   *tools.ToolRegistry
	processMgr *processmanager.ProcessManager
	eventBus   *eventbus.EventBus
	sessionSvc *SessionService
	storage    *db.Repository
	llmClient  *llmclient.Client
	messageBus *MessageBus
}

func NewToolService(
	registry *tools.ToolRegistry,
	processMgr *processmanager.ProcessManager,
	eventBus *eventbus.EventBus,
	sessionSvc *SessionService,
	storage *db.Repository,
	llmClient *llmclient.Client,
	messageBus *MessageBus,
) *ToolService {
	return &ToolService{
		registry:   registry,
		processMgr: processMgr,
		eventBus:   eventBus,
		sessionSvc: sessionSvc,
		storage:    storage,
		llmClient:  llmClient,
		messageBus: messageBus,
	}
}

func (t *ToolService) List() []*types.ToolDefinition {
	return t.registry.GetDefinitions()
}

func (t *ToolService) Execute(ctx context.Context, name string, args map[string]interface{}, sessionID string) (*types.ToolResult, error) {
	session, _ := t.sessionSvc.Get(ctx, sessionID)
	sessionName := ""
	if session != nil {
		sessionName = session.Name
	}

	toolCtx := &interfaces.ToolContext{
		SessionID:      sessionID,
		SessionName:    sessionName,
		WorkingDir:     "",
		ProcessManager: t.processMgr,
		ToolRegistry:   t.registry,
		EventBus:       t.eventBus,
		Storage:        t.storage,
		SessionService: t.sessionSvc,
		LLMClient:      t.llmClient,
		MessageBus:     t.messageBus,
	}

	return t.registry.Execute(ctx, name, args, toolCtx)
}

type ServiceLayer struct {
	Session  *SessionService
	Process  *ProcessService
	Tool     *ToolService
	EventBus *eventbus.EventBus
}

func CreateServiceLayer(
	storage *db.Repository,
	eventBus *eventbus.EventBus,
	llmClient *llmclient.Client,
	toolRegistry *tools.ToolRegistry,
	processMgr *processmanager.ProcessManager,
) *ServiceLayer {
	messageBus := NewMessageBus()

	sessionSvc := NewSessionService(storage, eventBus, llmClient, toolRegistry, messageBus)
	processSvc := NewProcessService(storage, processMgr)
	toolSvc := NewToolService(toolRegistry, processMgr, eventBus, sessionSvc, storage, llmClient, messageBus)

	return &ServiceLayer{
		Session:  sessionSvc,
		Process:  processSvc,
		Tool:     toolSvc,
		EventBus: eventBus,
	}
}
