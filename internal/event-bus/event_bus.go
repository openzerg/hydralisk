package eventbus

import (
	"github.com/google/uuid"
)

type handlerEntry struct {
	id      string
	handler func(map[string]interface{})
}

type EventBus struct {
	handlers map[string][]handlerEntry
}

func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[string][]handlerEntry),
	}
}

func (eb *EventBus) Emit(event string, data map[string]interface{}) {
	if entries, ok := eb.handlers[event]; ok {
		for _, entry := range entries {
			go entry.handler(data)
		}
	}
}

func (eb *EventBus) EmitGlobal(event interface{}) {
	if entries, ok := eb.handlers["global"]; ok {
		for _, entry := range entries {
			go entry.handler(map[string]interface{}{"event": event})
		}
	}
}

func (eb *EventBus) On(event string, handler func(map[string]interface{})) string {
	id := uuid.New().String()
	entry := handlerEntry{id: id, handler: handler}
	eb.handlers[event] = append(eb.handlers[event], entry)
	return id
}

func (eb *EventBus) Once(event string, handler func(map[string]interface{})) {
	id := uuid.New().String()
	var onceHandler func(map[string]interface{})
	onceHandler = func(data map[string]interface{}) {
		eb.Off(event, id)
		handler(data)
	}
	entry := handlerEntry{id: id, handler: onceHandler}
	eb.handlers[event] = append(eb.handlers[event], entry)
}

func (eb *EventBus) Off(event string, handlerID string) {
	if entries, ok := eb.handlers[event]; ok {
		for i, entry := range entries {
			if entry.id == handlerID {
				eb.handlers[event] = append(entries[:i], entries[i+1:]...)
				break
			}
		}
	}
}

func (eb *EventBus) RemoveAllListeners(event *string) {
	if event != nil {
		delete(eb.handlers, *event)
	} else {
		eb.handlers = make(map[string][]handlerEntry)
	}
}
