package chat

import (
	"time"
)

type Message struct {
	MessageID        int64
	Topic            string
	SendTime         time.Time
	SenderID         string
	SenderIdentifier string // TODO: remove SenderIdentifier; look up identifiers separately
	Body             string
}
