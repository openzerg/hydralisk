package interfaces

type OutboundMessage struct {
	Action    string `json:"action"`
	To        string `json:"to"`
	Content   string `json:"content,omitempty"`
	File      string `json:"file,omitempty"`
	ReplyTo   string `json:"reply_to,omitempty"`
	SessionID string `json:"session_id"`
	Timestamp int64  `json:"timestamp"`
}

type MessageHandler struct {
	ID        string
	OnMessage func(message *OutboundMessage) error
}

type IMessageBus interface {
	Subscribe(handler *MessageHandler) string
	Unsubscribe(handlerID string)
	Broadcast(message *OutboundMessage) error
	SubscriberCount() int
}
