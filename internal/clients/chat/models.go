package chat

import (
	"time"

	"zombiezen.com/go/sqlite"
)

type (
	MessageID int64
	SenderID  string
	Topic     string
)

// Message

type Message struct {
	ID       MessageID
	Topic    Topic
	SendTime time.Time
	SenderID SenderID
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

// Messages

func newMessagesByTopicSelection(topic Topic, messagesLimit int64) map[string]interface{} {
	return map[string]interface{}{
		"$topic":      topic,
		"$rows_limit": messagesLimit,
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
		ID:       MessageID(s.GetInt64("id")),
		Topic:    Topic(s.GetText("topic")),
		SendTime: time.UnixMilli(s.GetInt64("send_time")),
		SenderID: SenderID(s.GetText("sender_id")),
		Body:     s.GetText("body"),
	}
	sel.messages = append(sel.messages, m)
	return nil
}

func (sel *messagesSelector) Messages() []Message {
	return sel.messages
}
