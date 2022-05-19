package chat

import (
	"time"

	"zombiezen.com/go/sqlite"
)

type Message struct {
	ID       int64
	Topic    string
	SendTime time.Time
	SenderID string
	Body     string
}

func (m Message) newInsertion() map[string]interface{} {
	return map[string]interface{}{
		"$topic":     m.Topic,
		"$send_time": m.SendTime.UnixMilli(),
		"$sender_id": m.SenderID,
		"$body":      m.Body,
	}
}

type messagesSelector struct {
	messages []Message
}

func newMessagesSelector() *messagesSelector {
	return &messagesSelector{
		messages: make([]Message, 0),
	}
}

func (sel *messagesSelector) Step(s *sqlite.Stmt) error {
	m := Message{
		ID:       s.GetInt64("id"),
		Topic:    s.GetText("topic"),
		SendTime: time.UnixMilli(s.GetInt64("send_time")),
		SenderID: s.GetText("sender_id"),
		Body:     s.GetText("body"),
	}
	sel.messages = append(sel.messages, m)
	return nil
}

func (sel *messagesSelector) Messages() []Message {
	return sel.messages
}
