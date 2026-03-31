package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/openzerg/hydralisk/internal/core/types"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

type Session struct {
	bun.BaseModel `bun:"table:sessions"`

	ID           string             `bun:"id,pk"`
	Name         string             `bun:"name,notnull"`
	Purpose      string             `bun:"purpose,default:'Default'"`
	State        types.SessionState `bun:"state,default:'Idle'"`
	CreatedAt    string             `bun:"created_at,default:datetime('now')"`
	StartedAt    *string            `bun:"started_at"`
	FinishedAt   *string            `bun:"finished_at"`
	MessageCount int                `bun:"message_count,default:0"`
	SystemPrompt string             `bun:"system_prompt,default:''"`
	ParentID     *string            `bun:"parent_id"`
	InputTokens  int                `bun:"input_tokens,default:0"`
	OutputTokens int                `bun:"output_tokens,default:0"`
}

type Message struct {
	bun.BaseModel `bun:"table:messages"`

	ID        string            `bun:"id,pk"`
	SessionID string            `bun:"session_id,notnull"`
	Role      types.MessageRole `bun:"role,notnull"`
	Content   string            `bun:"content,notnull"`
	Timestamp string            `bun:"timestamp,default:datetime('now')"`
	ToolCalls []types.ToolCall  `bun:"tool_calls,type:json"`
}

type Process struct {
	bun.BaseModel `bun:"table:processes"`

	ID              string              `bun:"id,pk"`
	Command         string              `bun:"command,notnull"`
	Cwd             string              `bun:"cwd,notnull"`
	Status          types.ProcessStatus `bun:"status,default:'Running'"`
	ExitCode        *int                `bun:"exit_code"`
	StartedAt       string              `bun:"started_at,default:datetime('now')"`
	FinishedAt      *string             `bun:"finished_at"`
	ParentSessionID *string             `bun:"parent_session_id"`
	UnitName        string              `bun:"unit_name,notnull"`
	TimeoutMs       int                 `bun:"timeout_ms,default:120000"`
	OutputDir       string              `bun:"output_dir,notnull"`
	StdoutSize      int64               `bun:"stdout_size,default:0"`
	StderrSize      int64               `bun:"stderr_size,default:0"`
	StdoutLines     int                 `bun:"stdout_lines,default:0"`
	StderrLines     int                 `bun:"stderr_lines,default:0"`
}

type Todo struct {
	bun.BaseModel `bun:"table:todos"`

	ID        string             `bun:"id,pk"`
	SessionID string             `bun:"session_id,notnull"`
	Content   string             `bun:"content,notnull"`
	Status    types.TodoStatus   `bun:"status,default:'pending'"`
	Priority  types.TodoPriority `bun:"priority,default:'medium'"`
	CreatedAt string             `bun:"created_at,default:datetime('now')"`
	UpdatedAt string             `bun:"updated_at,default:datetime('now')"`
}

type Provider struct {
	bun.BaseModel `bun:"table:providers"`

	ID          string   `bun:"id,pk"`
	Name        string   `bun:"name,notnull,unique"`
	BaseURL     string   `bun:"base_url,notnull"`
	APIKey      string   `bun:"api_key,notnull"`
	Model       string   `bun:"model,notnull"`
	MaxTokens   *int     `bun:"max_tokens"`
	Temperature *float64 `bun:"temperature"`
	TopP        *float64 `bun:"top_p"`
	TopK        *int     `bun:"top_k"`
	ExtraParams *string  `bun:"extra_params"`
	IsActive    bool     `bun:"is_active,default:false"`
	CreatedAt   string   `bun:"created_at,default:datetime('now')"`
	UpdatedAt   string   `bun:"updated_at,default:datetime('now')"`
}

type Activity struct {
	bun.BaseModel `bun:"table:activities"`

	ID           string  `bun:"id,pk"`
	SessionID    *string `bun:"session_id"`
	ActivityType string  `bun:"activity_type,notnull"`
	Description  string  `bun:"description,notnull"`
	Details      string  `bun:"details,default:'{}'"`
	Timestamp    string  `bun:"timestamp,default:datetime('now')"`
}

type Repository struct {
	db *bun.DB
}

func NewRepository(databasePath string) (*Repository, error) {
	sqldb, err := sql.Open(sqliteshim.DriverName(), databasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := bun.NewDB(sqldb, sqlitedialect.New())

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &Repository{db: db}, nil
}

func createTables(db *bun.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			purpose TEXT NOT NULL DEFAULT 'Default',
			state TEXT NOT NULL DEFAULT 'Idle',
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			started_at TEXT,
			finished_at TEXT,
			message_count INTEGER DEFAULT 0,
			system_prompt TEXT DEFAULT '',
			parent_id TEXT REFERENCES sessions(id),
			input_tokens INTEGER DEFAULT 0,
			output_tokens INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			timestamp TEXT NOT NULL DEFAULT (datetime('now')),
			tool_calls TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id)`,
		`CREATE TABLE IF NOT EXISTS processes (
			id TEXT PRIMARY KEY,
			command TEXT NOT NULL,
			cwd TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'Running',
			exit_code INTEGER,
			started_at TEXT NOT NULL DEFAULT (datetime('now')),
			finished_at TEXT,
			parent_session_id TEXT REFERENCES sessions(id),
			unit_name TEXT NOT NULL,
			timeout_ms INTEGER DEFAULT 120000,
			output_dir TEXT NOT NULL,
			stdout_size INTEGER DEFAULT 0,
			stderr_size INTEGER DEFAULT 0,
			stdout_lines INTEGER DEFAULT 0,
			stderr_lines INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS todos (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			content TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			priority TEXT NOT NULL DEFAULT 'medium',
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			updated_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS providers (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			base_url TEXT NOT NULL,
			api_key TEXT NOT NULL,
			model TEXT NOT NULL,
			max_tokens INTEGER,
			temperature REAL,
			top_p REAL,
			top_k INTEGER,
			extra_params TEXT,
			is_active INTEGER DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			updated_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS activities (
			id TEXT PRIMARY KEY,
			session_id TEXT,
			activity_type TEXT NOT NULL,
			description TEXT NOT NULL,
			details TEXT DEFAULT '{}',
			timestamp TEXT NOT NULL DEFAULT (datetime('now'))
		)`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) CreateSession(ctx context.Context, data *types.CreateSessionData) (*types.Session, error) {
	session := &Session{
		ID:           data.ID,
		Name:         data.Name,
		Purpose:      "Default",
		State:        types.SessionStateIdle,
		CreatedAt:    time.Now().Format(time.RFC3339),
		SystemPrompt: "",
	}
	if data.Purpose != nil {
		session.Purpose = *data.Purpose
	}
	if data.SystemPrompt != nil {
		session.SystemPrompt = *data.SystemPrompt
	}
	if data.ParentID != nil {
		session.ParentID = data.ParentID
	}

	_, err := r.db.NewInsert().Model(session).Exec(ctx)
	if err != nil {
		return nil, err
	}

	return r.toSession(session), nil
}

func (r *Repository) GetSession(ctx context.Context, id string) (*types.Session, error) {
	session := new(Session)
	err := r.db.NewSelect().Model(session).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return r.toSession(session), nil
}

func (r *Repository) ListSessions(ctx context.Context, filter *types.SessionFilter) ([]*types.Session, error) {
	var sessions []*Session
	q := r.db.NewSelect().Model(&sessions)
	if filter != nil {
		if filter.State != nil {
			q = q.Where("state = ?", *filter.State)
		}
		if filter.ParentID != nil {
			q = q.Where("parent_id = ?", *filter.ParentID)
		}
	}
	err := q.Order("created_at DESC").Scan(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*types.Session, len(sessions))
	for i, s := range sessions {
		result[i] = r.toSession(s)
	}
	return result, nil
}

func (r *Repository) UpdateSession(ctx context.Context, id string, data *types.UpdateSessionData) error {
	session := new(Session)
	session.ID = id
	if data.Name != nil {
		session.Name = *data.Name
	}
	if data.State != nil {
		session.State = *data.State
	}
	if data.StartedAt != nil {
		session.StartedAt = data.StartedAt
	}
	if data.FinishedAt != nil {
		session.FinishedAt = data.FinishedAt
	}
	if data.InputTokens != nil {
		session.InputTokens = *data.InputTokens
	}
	if data.OutputTokens != nil {
		session.OutputTokens = *data.OutputTokens
	}

	_, err := r.db.NewUpdate().Model(session).WherePK().OmitZero().Exec(ctx)
	return err
}

func (r *Repository) DeleteSession(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Model((*Session)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}

func (r *Repository) IncrementMessageCount(ctx context.Context, sessionID string) error {
	_, err := r.db.NewUpdate().Model((*Session)(nil)).
		Set("message_count = message_count + 1").
		Where("id = ?", sessionID).
		Exec(ctx)
	return err
}

func (r *Repository) SaveMessage(ctx context.Context, data *types.CreateMessageData) (*types.Message, error) {
	msg := &Message{
		ID:        data.ID,
		SessionID: data.SessionID,
		Role:      data.Role,
		Content:   data.Content,
		Timestamp: time.Now().Format(time.RFC3339),
		ToolCalls: data.ToolCalls,
	}
	_, err := r.db.NewInsert().Model(msg).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return r.toMessage(msg), nil
}

func (r *Repository) GetMessages(ctx context.Context, sessionID string, filter *types.MessageFilter) ([]*types.Message, error) {
	var messages []*Message
	q := r.db.NewSelect().Model(&messages).Where("session_id = ?", sessionID)
	if filter != nil && filter.Limit != nil {
		q = q.Limit(*filter.Limit)
	}
	err := q.Order("timestamp ASC").Scan(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*types.Message, len(messages))
	for i, m := range messages {
		result[i] = r.toMessage(m)
	}
	return result, nil
}

func (r *Repository) DeleteMessages(ctx context.Context, sessionID string) error {
	_, err := r.db.NewDelete().Model((*Message)(nil)).Where("session_id = ?", sessionID).Exec(ctx)
	return err
}

func (r *Repository) SaveProcess(ctx context.Context, data *types.CreateProcessData) (*types.Process, error) {
	proc := &Process{
		ID:              data.ID,
		Command:         data.Command,
		Cwd:             data.Cwd,
		Status:          types.ProcessStatusRunning,
		StartedAt:       time.Now().Format(time.RFC3339),
		ParentSessionID: data.ParentSessionID,
		UnitName:        data.UnitName,
		OutputDir:       data.OutputDir,
	}
	if data.TimeoutMs != nil {
		proc.TimeoutMs = *data.TimeoutMs
	}

	_, err := r.db.NewInsert().Model(proc).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return r.toProcess(proc), nil
}

func (r *Repository) GetProcess(ctx context.Context, id string) (*types.Process, error) {
	proc := new(Process)
	err := r.db.NewSelect().Model(proc).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return r.toProcess(proc), nil
}

func (r *Repository) ListProcesses(ctx context.Context, filter map[string]interface{}) ([]*types.Process, error) {
	var processes []*Process
	q := r.db.NewSelect().Model(&processes)
	for k, v := range filter {
		q = q.Where(k+" = ?", v)
	}
	err := q.Order("started_at DESC").Scan(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*types.Process, len(processes))
	for i, p := range processes {
		result[i] = r.toProcess(p)
	}
	return result, nil
}

func (r *Repository) UpdateProcessStatus(ctx context.Context, id string, status types.ProcessStatus, exitCode *int) error {
	updates := map[string]interface{}{
		"status":      status,
		"finished_at": time.Now().Format(time.RFC3339),
	}
	if exitCode != nil {
		updates["exit_code"] = *exitCode
	}
	_, err := r.db.NewUpdate().Model((*Process)(nil)).Set("status = ?, finished_at = ?", status, time.Now().Format(time.RFC3339)).Where("id = ?", id).Exec(ctx)
	return err
}

func (r *Repository) UpdateProcessOutputStats(ctx context.Context, id string, stats *types.OutputStats) error {
	_, err := r.db.NewUpdate().Model((*Process)(nil)).
		Set("stdout_size = ?", stats.StdoutSize).
		Set("stderr_size = ?", stats.StderrSize).
		Set("stdout_lines = ?", stats.StdoutLines).
		Set("stderr_lines = ?", stats.StderrLines).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (r *Repository) CreateTodo(ctx context.Context, sessionID, content, priority string) (*types.Todo, error) {
	todo := &Todo{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Content:   content,
		Status:    types.TodoStatusPending,
		Priority:  types.TodoPriorityMedium,
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}
	if priority != "" {
		todo.Priority = types.TodoPriority(priority)
	}

	_, err := r.db.NewInsert().Model(todo).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return r.toTodo(todo), nil
}

func (r *Repository) GetTodos(ctx context.Context, sessionID string) ([]*types.Todo, error) {
	var todos []*Todo
	err := r.db.NewSelect().Model(&todos).Where("session_id = ?", sessionID).Order("created_at ASC").Scan(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*types.Todo, len(todos))
	for i, t := range todos {
		result[i] = r.toTodo(t)
	}
	return result, nil
}

func (r *Repository) UpdateTodo(ctx context.Context, id string, data map[string]interface{}) error {
	data["updated_at"] = time.Now().Format(time.RFC3339)
	q := r.db.NewUpdate().Model((*Todo)(nil)).Where("id = ?", id)
	for k, v := range data {
		q = q.Set(k+" = ?", v)
	}
	_, err := q.Exec(ctx)
	return err
}

func (r *Repository) DeleteTodo(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Model((*Todo)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}

func (r *Repository) GetProvider(ctx context.Context, name string) (*types.Provider, error) {
	provider := new(Provider)
	err := r.db.NewSelect().Model(provider).Where("name = ?", name).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return r.toProvider(provider), nil
}

func (r *Repository) ListProviders(ctx context.Context) ([]*types.Provider, error) {
	var providers []*Provider
	err := r.db.NewSelect().Model(&providers).Order("created_at ASC").Scan(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*types.Provider, len(providers))
	for i, p := range providers {
		result[i] = r.toProvider(p)
	}
	return result, nil
}

func (r *Repository) SaveProvider(ctx context.Context, provider *types.Provider) (*types.Provider, error) {
	p := &Provider{
		ID:        uuid.New().String(),
		Name:      provider.Name,
		BaseURL:   provider.BaseURL,
		APIKey:    provider.APIKey,
		Model:     provider.Model,
		MaxTokens: provider.MaxTokens,
		IsActive:  provider.IsActive,
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	_, err := r.db.NewInsert().Model(p).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return r.toProvider(p), nil
}

func (r *Repository) UpdateProvider(ctx context.Context, id string, data map[string]interface{}) error {
	data["updated_at"] = time.Now().Format(time.RFC3339)
	q := r.db.NewUpdate().Model((*Provider)(nil)).Where("id = ?", id)
	for k, v := range data {
		q = q.Set(k+" = ?", v)
	}
	_, err := q.Exec(ctx)
	return err
}

func (r *Repository) DeleteProvider(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Model((*Provider)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}

func (r *Repository) RecordActivity(ctx context.Context, sessionID *string, activityType, description string, details map[string]interface{}) error {
	detailsJSON := "{}"
	if details != nil {
		b, _ := json.Marshal(details)
		detailsJSON = string(b)
	}

	activity := &Activity{
		ID:           uuid.New().String(),
		SessionID:    sessionID,
		ActivityType: activityType,
		Description:  description,
		Details:      detailsJSON,
		Timestamp:    time.Now().Format(time.RFC3339),
	}

	_, err := r.db.NewInsert().Model(activity).Exec(ctx)
	return err
}

func (r *Repository) GetActivities(ctx context.Context, sessionID string) ([]*types.Activity, error) {
	var activities []*Activity
	err := r.db.NewSelect().Model(&activities).Where("session_id = ?", sessionID).Order("timestamp DESC").Scan(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*types.Activity, len(activities))
	for i, a := range activities {
		result[i] = r.toActivity(a)
	}
	return result, nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func (r *Repository) toSession(s *Session) *types.Session {
	return &types.Session{
		ID:           s.ID,
		Name:         s.Name,
		Purpose:      s.Purpose,
		State:        s.State,
		CreatedAt:    s.CreatedAt,
		StartedAt:    s.StartedAt,
		FinishedAt:   s.FinishedAt,
		MessageCount: s.MessageCount,
		SystemPrompt: s.SystemPrompt,
		ParentID:     s.ParentID,
		InputTokens:  s.InputTokens,
		OutputTokens: s.OutputTokens,
	}
}

func (r *Repository) toMessage(m *Message) *types.Message {
	return &types.Message{
		ID:        m.ID,
		SessionID: m.SessionID,
		Role:      m.Role,
		Content:   m.Content,
		Timestamp: m.Timestamp,
		ToolCalls: m.ToolCalls,
	}
}

func (r *Repository) toProcess(p *Process) *types.Process {
	return &types.Process{
		ID:              p.ID,
		Command:         p.Command,
		Cwd:             p.Cwd,
		Status:          p.Status,
		ExitCode:        p.ExitCode,
		StartedAt:       p.StartedAt,
		FinishedAt:      p.FinishedAt,
		ParentSessionID: p.ParentSessionID,
		UnitName:        p.UnitName,
		TimeoutMs:       p.TimeoutMs,
		OutputDir:       p.OutputDir,
		StdoutSize:      p.StdoutSize,
		StderrSize:      p.StderrSize,
		StdoutLines:     p.StdoutLines,
		StderrLines:     p.StderrLines,
	}
}

func (r *Repository) toTodo(t *Todo) *types.Todo {
	return &types.Todo{
		ID:        t.ID,
		SessionID: t.SessionID,
		Content:   t.Content,
		Status:    t.Status,
		Priority:  t.Priority,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

func (r *Repository) toProvider(p *Provider) *types.Provider {
	return &types.Provider{
		ID:          p.ID,
		Name:        p.Name,
		BaseURL:     p.BaseURL,
		APIKey:      p.APIKey,
		Model:       p.Model,
		MaxTokens:   p.MaxTokens,
		Temperature: p.Temperature,
		TopP:        p.TopP,
		TopK:        p.TopK,
		ExtraParams: p.ExtraParams,
		IsActive:    p.IsActive,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func (r *Repository) toActivity(a *Activity) *types.Activity {
	return &types.Activity{
		ID:           a.ID,
		SessionID:    a.SessionID,
		ActivityType: a.ActivityType,
		Description:  a.Description,
		Details:      a.Details,
		Timestamp:    a.Timestamp,
	}
}
