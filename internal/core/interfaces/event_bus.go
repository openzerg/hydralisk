package interfaces

type IEventBus interface {
	Emit(event string, data map[string]interface{})
	EmitGlobal(event interface{})
	On(event string, handler func(data map[string]interface{})) string
	Once(event string, handler func(data map[string]interface{}))
	Off(event string, handlerID string)
	RemoveAllListeners(event *string)
}
