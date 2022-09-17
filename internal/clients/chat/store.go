// Package chat provides a high-level store of chat messages
package chat

import (
	"context"
	_ "embed"
	"strings"

	"github.com/pkg/errors"
	"github.com/sargassum-world/godest/database"
)

type Store struct {
	db *database.DB
}

func NewStore(db *database.DB) *Store {
	return &Store{
		db: db,
	}
}

//go:embed queries/insert-message.sql
var rawInsertMessageQuery string
var insertMessageQuery string = strings.TrimSpace(rawInsertMessageQuery)

func (s *Store) AddMessage(ctx context.Context, m Message) (messageID int64, err error) {
	rowID, err := s.db.ExecuteInsertion(ctx, insertMessageQuery, m.newInsertion())
	if err != nil {
		return 0, errors.Wrapf(err, "couldn't add chat message with topic %s", m.Topic)
	}
	// TODO: instead of returning the raw ID, return the frontend-facing ID as a salted SHA-256 hash
	// of the ID to mitigate the insecure direct object reference vulnerability and avoid leaking
	// info about instrument creation?
	return rowID, err
}

//go:embed queries/select-messages-by-topic.sql
var rawSelectMessagesByTopicQuery string
var selectMessagesByTopicQuery string = strings.TrimSpace(rawSelectMessagesByTopicQuery)

const DefaultMessagesLimit = 50

func (s *Store) GetMessagesByTopic(
	ctx context.Context, topic string, messagesLimit int64,
) (messages []Message, err error) {
	sel := newMessagesSelector()
	if err = s.db.ExecuteSelection(
		ctx, selectMessagesByTopicQuery,
		map[string]interface{}{
			"$topic":      topic,
			"$rows_limit": messagesLimit,
		},
		sel.Step,
	); err != nil {
		return nil, errors.Wrapf(err, "couldn't get chat messages with topic %s", topic)
	}
	return sel.Messages(), nil
}
