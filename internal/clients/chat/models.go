package chat

import (
	"time"
)

type Message struct {
	Topic            string
	SendTime         time.Time
	SenderID         string
	SenderIdentifier string
	Body             string
}
