// Package chat provides a high-level store of chat messages
package chat

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"

	"github.com/sargassum-world/pslive/internal/clients/database"
)

type Store struct {
	db *database.DB
}

func NewStore(db *database.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) AddMessage(ctx context.Context, m Message) error {
	conn, err := s.db.AcquireWriter(ctx)
	if err != nil {
		return errors.Wrap(err, "couldn't acquire writer to add chat message")
	}
	defer s.db.ReleaseWriter(conn)

	return errors.Wrapf(
		// TODO: the query should be embedded in this package!
		s.db.Execute(conn, "chat/insert-message.sql", &sqlitex.ExecOptions{
			Named: map[string]interface{}{
				"$topic":             m.Topic,
				"$send_time":         m.SendTime.UnixMilli(),
				"$sender_id":         m.SenderID,
				"$sender_identifier": m.SenderIdentifier,
				"$body":              m.Body,
			},
		}),
		"couldn't execute query to add chat message with topic %s", m.Topic,
	)
	// TODO: return the frontend-facing message ID (it should be a salted SHA-256 hash of message_id
	// to mitigate the insecure direct object reference vulnerability)
}

const DefaultMessagesLimit = 10

func (s *Store) GetMessagesByTopic(
	ctx context.Context, topic string, messagesLimit int64,
) ([]Message, error) {
	conn, err := s.db.AcquireReader(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't acquire reader to get chat messages by topic")
	}
	defer s.db.ReleaseReader(conn)

	messages := make([]Message, 0)
	// TODO: the query should be embedded in this package!
	if err := s.db.Execute(conn, "chat/select-messages-by-topic.sql", &sqlitex.ExecOptions{
		Named: map[string]interface{}{
			"$topic":      topic,
			"$rows_limit": messagesLimit,
		},
		ResultFunc: func(s *sqlite.Stmt) error {
			message := Message{
				Topic:            s.GetText("topic"),
				SendTime:         time.UnixMilli(s.GetInt64("send_time")),
				SenderID:         s.GetText("sender_id"),
				SenderIdentifier: s.GetText("sender_identifier"),
				Body:             s.GetText("body"),
			}
			messages = append(messages, message)
			return nil
		},
	}); err != nil {
		return nil, errors.Wrapf(
			err, "couldn't execute query to get chat messages with topic %s", topic,
		)
	}
	return messages, nil
}
