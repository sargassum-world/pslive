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

func NewMessage(s *sqlite.Stmt) Message {
	return Message{
		ID:       s.GetInt64("id"),
		Topic:    s.GetText("topic"),
		SendTime: time.UnixMilli(s.GetInt64("send_time")),
		SenderID: s.GetText("sender_id"),
		Body:     s.GetText("body"),
	}
}

func (m Message) NewInsertion() map[string]interface{} {
	return map[string]interface{}{
		"$topic":     m.Topic,
		"$send_time": m.SendTime.UnixMilli(),
		"$sender_id": m.SenderID,
		"$body":      m.Body,
	}
}
