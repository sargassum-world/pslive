package videostreams

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/sargassum-world/godest/actioncable"
	"github.com/sargassum-world/godest/pubsub"
)

const ChannelName = "Video::StreamsChannel"

type subscriber func(ctx context.Context, topic string) <-chan Frame

type Channel struct {
	identifier string
	streamName string
	h          *pubsub.Hub[[]Frame]
	subscriber subscriber
	sessionID  string
	logger     pubsub.Logger
}

func parseStreamName(identifier string) (string, error) {
	var i struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(identifier), &i); err != nil {
		return "", errors.Wrap(err, "couldn't parse stream name from identifier")
	}
	return i.Name, nil
}

func NewChannel(
	identifier string, h *pubsub.Hub[[]Frame], subscriber subscriber, sessionID string,
	logger pubsub.Logger, checkers ...actioncable.IdentifierChecker,
) (*Channel, error) {
	name, err := parseStreamName(identifier)
	if err != nil {
		return nil, err
	}
	for _, checker := range checkers {
		if err := checker(identifier); err != nil {
			return nil, errors.Wrap(err, "stream identifier failed checks")
		}
	}
	return &Channel{
		identifier: identifier,
		streamName: name,
		h:          h,
		subscriber: subscriber,
		sessionID:  sessionID,
		logger:     logger,
	}, nil
}

func (c *Channel) Subscribe(
	ctx context.Context, sub actioncable.Subscription,
) (unsubscriber func(), err error) {
	if sub.Identifier() != c.identifier {
		return nil, errors.Errorf(
			"channel identifier %+v does not match subscription identifier %+v",
			c.identifier, sub.Identifier(),
		)
	}

	ctx, cancel := context.WithCancel(ctx)
	frames := c.subscriber(ctx, c.streamName)
	go func() {
		for frame := range frames {
			encoded, err := frame.AsBase64Frame()
			if err != nil {
				c.logger.Error(errors.Wrap(err, "couldn't base64-encode frame to send over action cable"))
				break
			}
			// TODO: attach metadata to the frame data by packing everything together in json with a
			// AsJSONFrame method on the videostreams.Frame interface
			if ok := sub.Receive(encoded.Im); !ok {
				break
			}
		}
		sub.Close()
	}()
	return cancel, nil
}

func (c *Channel) Perform(data string) error {
	return errors.New("video streams channel cannot perform any actions")
}

func NewChannelFactory(
	b *Broker, sessionID string, logger pubsub.Logger, checkers ...actioncable.IdentifierChecker,
) actioncable.ChannelFactory {
	return func(identifier string) (actioncable.Channel, error) {
		return NewChannel(identifier, b.Hub(), b.Subscribe, sessionID, logger, checkers...)
	}
}
