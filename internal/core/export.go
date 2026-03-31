package core

import (
	"github.com/openzerg/hydralisk/internal/core/interfaces"
	"github.com/openzerg/hydralisk/internal/core/types"
)

type Session = types.Session
type SessionState = types.SessionState
type CreateSessionData = types.CreateSessionData
type UpdateSessionData = types.UpdateSessionData
type SessionFilter = types.SessionFilter

type Message = types.Message
type MessageRole = types.MessageRole
type CreateMessageData = types.CreateMessageData
type MessageFilter = types.MessageFilter
type ToolCall = types.ToolCall

type Process = types.Process
type ProcessStatus = types.ProcessStatus
type ProcessHandle = types.ProcessHandle
type ProcessResult = types.ProcessResult
type OutputStats = types.OutputStats
type ProcessOutput = types.ProcessOutput
type ProcessOutputLine = types.ProcessOutputLine
type CreateProcessData = types.CreateProcessData
type SpawnOptions = types.SpawnOptions

type JSONSchema = types.JSONSchema
type ToolFunction = types.ToolFunction
type ToolDefinition = types.ToolDefinition
type Attachment = types.Attachment
type ToolResult = types.ToolResult
type ToolExecuteRequest = types.ToolExecuteRequest

type GlobalEvent = types.GlobalEvent

type LLMMessageRole = types.LLMMessageRole
type LLMToolCall = types.LLMToolCall
type LLMMessage = types.LLMMessage
type ChatCompletionRequest = types.ChatCompletionRequest
type ChatCompletionResponse = types.ChatCompletionResponse
type StreamChunk = types.StreamChunk
type LLMConfig = types.LLMConfig
type Provider = types.Provider

type Todo = types.Todo
type TodoStatus = types.TodoStatus
type TodoPriority = types.TodoPriority
type Activity = types.Activity
type SessionContext = types.SessionContext

type OutboundMessage = interfaces.OutboundMessage
type MessageHandler = interfaces.MessageHandler

type IStorage = interfaces.IStorage
type IProcessManager = interfaces.IProcessManager
type IEventBus = interfaces.IEventBus
type ILLMClient = interfaces.ILLMClient
type IMessageBus = interfaces.IMessageBus
type ISessionManager = interfaces.ISessionManager
type ISessionProcessor = interfaces.ISessionProcessor
type ISessionService = interfaces.ISessionService
type ITool = interfaces.ITool
type IToolRegistry = interfaces.IToolRegistry
type ToolContext = interfaces.ToolContext
