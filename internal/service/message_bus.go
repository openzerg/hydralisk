package service

import (
	"sync"

	"github.com/openzerg/hydralisk/internal/core/interfaces"
)

type MessageHandler struct {
	ID        string
	OnMessage func(message *interfaces.OutboundMessage) error
}

type MessageBus struct {
	mu       sync.RWMutex
	handlers map[string]*MessageHandler
}

func NewMessageBus() *MessageBus {
	return &MessageBus{
		handlers: make(map[string]*MessageHandler),
	}
}

func (mb *MessageBus) Subscribe(handler *MessageHandler) string {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	mb.handlers[handler.ID] = handler
	return handler.ID
}

func (mb *MessageBus) Unsubscribe(handlerID string) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	delete(mb.handlers, handlerID)
}

func (mb *MessageBus) Broadcast(message *interfaces.OutboundMessage) error {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	for _, handler := range mb.handlers {
		go func(h *MessageHandler) {
			_ = h.OnMessage(message)
		}(handler)
	}
	return nil
}

func (mb *MessageBus) SubscriberCount() int {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	return len(mb.handlers)
}
